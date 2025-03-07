// Copyright © 2021 Attestant Limited.
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

package prometheus_test

import (
	"context"
	"testing"
	"time"

	"github.com/attestantio/vouch/mock"
	standardchaintime "github.com/attestantio/vouch/services/chaintime/standard"
	"github.com/attestantio/vouch/services/metrics/prometheus"
	"github.com/attestantio/vouch/testing/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	ctx := context.Background()

	chainTime, err := standardchaintime.New(ctx,
		standardchaintime.WithLogLevel(zerolog.Disabled),
		standardchaintime.WithGenesisTimeProvider(mock.NewGenesisTimeProvider(time.Now())),
		standardchaintime.WithSlotDurationProvider(mock.NewSlotDurationProvider(12*time.Second)),
		standardchaintime.WithSlotsPerEpochProvider(mock.NewSlotsPerEpochProvider(32)),
	)
	require.NoError(t, err)

	tests := []struct {
		name     string
		params   []prometheus.Parameter
		err      string
		logEntry string
	}{
		{
			name: "AddressMissing",
			params: []prometheus.Parameter{
				prometheus.WithLogLevel(zerolog.Disabled),
				prometheus.WithChainTime(chainTime),
			},
			err: "problem with parameters: no address specified",
		},
		{
			name: "ChainTimeMissing",
			params: []prometheus.Parameter{
				prometheus.WithLogLevel(zerolog.Disabled),
				prometheus.WithAddress("http://localhost:12345/"),
			},
			err: "problem with parameters: no chain time service specified",
		},
		{
			name: "Good",
			params: []prometheus.Parameter{
				prometheus.WithLogLevel(zerolog.Disabled),
				prometheus.WithAddress("http://localhost:12345/"),
				prometheus.WithChainTime(chainTime),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capture := logger.NewLogCapture()
			_, err := prometheus.New(context.Background(), test.params...)
			if test.err != "" {
				require.EqualError(t, err, test.err)
				if test.logEntry != "" {
					capture.AssertHasEntry(t, test.logEntry)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
