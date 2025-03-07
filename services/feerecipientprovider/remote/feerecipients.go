// Copyright © 2022 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type feeRecipientsResponseJSON struct {
	FeeRecipients []*feeRecipientResponseJSON `json:"fee_recipients"`
}

type feeRecipientResponseJSON struct {
	Index        uint64 `json:"index"`
	FeeRecipient string `json:"fee_recipient"`
}

var zeroFeeRecipient = bellatrix.ExecutionAddress{}

// FeeRecipients returns the fee recipients for the given validators.
func (s *Service) FeeRecipients(ctx context.Context,
	indices []phase0.ValidatorIndex,
) (
	map[phase0.ValidatorIndex]bellatrix.ExecutionAddress,
	error,
) {
	started := time.Now()

	feeRecipients, err := s.feeRecipients(ctx, indices)
	if err != nil {
		feeRecipientsCompleted(started, "failed")
		return nil, err
	}

	feeRecipientsCompleted(started, "succeeded")
	return feeRecipients, nil
}

func (s *Service) feeRecipients(ctx context.Context,
	indices []phase0.ValidatorIndex,
) (
	map[phase0.ValidatorIndex]bellatrix.ExecutionAddress,
	error,
) {
	res := make(map[phase0.ValidatorIndex]bellatrix.ExecutionAddress, len(indices))

	data, err := s.fetchFeeRecipientsFromRemote(ctx, indices)
	if err != nil {
		log.Error().Err(err).Msg("Failed to obtain fee recipients from remote")
		// This is an error, however we still have default and cached results so keep going.
	}
	if data != nil {
		parseFeeRecipients(res, data.FeeRecipients)
	}

	if len(res) != len(indices) {
		log.Trace().Int("fee_recipients", len(res)).Msg("Padding fee recipients")
		s.padResults(ctx, indices, res)
	}

	if len(res) != len(indices) {
		log.Error().Int("fee_recipients", len(res)).Int("validators", len(indices)).Msg("Failed to obtain fee recipients for all validators")
	}

	log.Trace().Int("fee_recipients", len(res)).Msg("Updated fee recipients")
	s.cacheMu.Lock()
	s.cache = res
	s.cacheMu.Unlock()

	return res, nil
}

// fetchFeeRecipientsFromRemote fetches fee recipients from the remote source.
func (s *Service) fetchFeeRecipientsFromRemote(ctx context.Context,
	indices []phase0.ValidatorIndex,
) (
	*feeRecipientsResponseJSON,
	error,
) {
	// Create the request body.
	body := new(bytes.Buffer)
	_, err := body.WriteString(`{"indices":[`)
	if err != nil {
		return nil, errors.Wrap(err, "buffer write failed")
	}
	for i, index := range indices {
		if i != 0 {
			body.WriteString(",")
		}
		body.WriteString(fmt.Sprintf("%d", index))
	}
	body.WriteString(`]}`)

	respBodyReader, err := s.post(ctx, "feerecipients", body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request fee recipients")
	}
	if respBodyReader == nil {
		return nil, errors.New("failed to obtain fee recipients")
	}

	var data feeRecipientsResponseJSON
	if err := json.NewDecoder(respBodyReader).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "failed to parse fee recipients response")
	}

	return &data, nil
}

// padResults fills in the results that have not been obtained remotely, either
// from the cache or the default results.
func (s *Service) padResults(ctx context.Context,
	indices []phase0.ValidatorIndex,
	res map[phase0.ValidatorIndex]bellatrix.ExecutionAddress,
) {
	for _, index := range indices {
		if _, exists := res[index]; !exists {
			s.padResult(ctx, index, res)
		}
	}
}

// padResult fills in a result that has not been obtained remotely, either
// from the cache or the default results.
func (s *Service) padResult(_ context.Context,
	index phase0.ValidatorIndex,
	res map[phase0.ValidatorIndex]bellatrix.ExecutionAddress,
) {
	s.cacheMu.RLock()
	feeRecipient, exists := s.cache[index]
	s.cacheMu.RUnlock()

	if exists {
		res[index] = feeRecipient
		feeRecipientObtained("cache")
	} else {
		res[index] = s.defaultFeeRecipient
		feeRecipientObtained("default")
	}
}

// post sends an HTTP post request and returns the body.
func (s *Service) post(ctx context.Context, endpoint string, body io.Reader) (io.Reader, error) {
	// #nosec G404
	log := log.With().Str("id", fmt.Sprintf("%02x", rand.Int31())).Str("endpoint", endpoint).Logger()
	if e := log.Trace(); e.Enabled() {
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			return nil, errors.New("failed to read request body")
		}
		body = bytes.NewReader(bodyBytes)

		e.Str("body", string(bodyBytes)).Msg("POST request")
	}

	url, err := url.Parse(fmt.Sprintf("%s/%s", s.baseURL.String(), endpoint))
	if err != nil {
		return nil, errors.Wrap(err, "invalid endpoint")
	}
	log.Trace().Str("url", url.String()).Msg("URL for POST")

	opCtx, cancel := context.WithTimeout(ctx, s.timeout)
	req, err := http.NewRequestWithContext(opCtx, http.MethodPost, url.String(), body)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to create POST request")
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to call POST endpoint")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "failed to read POST response")
	}

	statusFamily := resp.StatusCode / 100
	if statusFamily != 2 {
		log.Trace().Int("status_code", resp.StatusCode).Str("data", string(data)).Msg("POST failed")
		cancel()
		return nil, fmt.Errorf("POST failed with status %d: %s", resp.StatusCode, string(data))
	}
	cancel()

	log.Trace().Str("response", string(data)).Msg("POST response")

	return bytes.NewReader(data), nil
}

func parseFeeRecipients(feeRecipients map[phase0.ValidatorIndex]bellatrix.ExecutionAddress, entries []*feeRecipientResponseJSON) {
	for _, feeRecipientEntry := range entries {
		feeRecipientBytes, err := hex.DecodeString(strings.TrimPrefix(feeRecipientEntry.FeeRecipient, "0x"))
		if err != nil {
			log.Error().Err(err).Uint64("index", feeRecipientEntry.Index).Str("fee_recipient", feeRecipientEntry.FeeRecipient).Msg("Failed to parse fee recipient address")
			continue
		}
		if bytes.Equal(feeRecipientBytes, zeroFeeRecipient[:]) {
			log.Warn().Uint64("index", feeRecipientEntry.Index).Str("fee_recipient", feeRecipientEntry.FeeRecipient).Msg("Received 0 fee recipient address; ignoring")
			continue
		}
		feeRecipient := bellatrix.ExecutionAddress{}
		copy(feeRecipient[:], feeRecipientBytes)
		feeRecipients[phase0.ValidatorIndex(feeRecipientEntry.Index)] = feeRecipient
		feeRecipientObtained("remote")
	}
}
