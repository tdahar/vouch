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

package standard_test

import (
	"context"
	"testing"
	"time"

	"github.com/attestantio/vouch/mock"
	mockaccountmanager "github.com/attestantio/vouch/services/accountmanager/mock"
	standardchaintime "github.com/attestantio/vouch/services/chaintime/standard"
	mockfeerecipientprovider "github.com/attestantio/vouch/services/feerecipientprovider/mock"
	prometheusmetrics "github.com/attestantio/vouch/services/metrics/prometheus"
	"github.com/attestantio/vouch/services/proposalpreparer/standard"
	"github.com/attestantio/vouch/testing/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	ctx := context.Background()

	zerolog.SetGlobalLevel(zerolog.Disabled)

	genesisTime := time.Now()
	slotDuration := 12 * time.Second
	slotsPerEpoch := uint64(32)
	genesisTimeProvider := mock.NewGenesisTimeProvider(genesisTime)
	slotDurationProvider := mock.NewSlotDurationProvider(slotDuration)
	slotsPerEpochProvider := mock.NewSlotsPerEpochProvider(slotsPerEpoch)

	mockValidatingAccountsProvider := mockaccountmanager.NewValidatingAccountsProvider()
	mockFeeRecipientProvider := mockfeerecipientprovider.New()
	mockProposalPreparationsSubmitter := mock.NewProposalPreparationsSubmitter()

	chainTime, err := standardchaintime.New(ctx,
		standardchaintime.WithGenesisTimeProvider(genesisTimeProvider),
		standardchaintime.WithSlotDurationProvider(slotDurationProvider),
		standardchaintime.WithSlotsPerEpochProvider(slotsPerEpochProvider),
	)
	require.NoError(t, err)

	prometheusMetrics, err := prometheusmetrics.New(ctx,
		prometheusmetrics.WithAddress(":12345"),
		prometheusmetrics.WithChainTime(chainTime),
	)
	require.NoError(t, err)

	tests := []struct {
		name     string
		params   []standard.Parameter
		err      string
		logEntry string
	}{
		{
			name: "MonitorNil",
			params: []standard.Parameter{
				standard.WithMonitor(nil),
				standard.WithLogLevel(zerolog.Disabled),
				standard.WithChainTimeService(chainTime),
				standard.WithValidatingAccountsProvider(mockValidatingAccountsProvider),
				standard.WithFeeRecipientProvider(mockFeeRecipientProvider),
				standard.WithProposalPreparationsSubmitter(mockProposalPreparationsSubmitter),
			},
			err: "problem with parameters: no monitor specified",
		},
		{
			name: "ChainTimeServiceMissing",
			params: []standard.Parameter{
				standard.WithLogLevel(zerolog.Disabled),
				standard.WithValidatingAccountsProvider(mockValidatingAccountsProvider),
				standard.WithFeeRecipientProvider(mockFeeRecipientProvider),
				standard.WithProposalPreparationsSubmitter(mockProposalPreparationsSubmitter),
			},
			err: "problem with parameters: no chain time service specified",
		},
		{
			name: "ValidatingAccountsProviderMissing",
			params: []standard.Parameter{
				standard.WithLogLevel(zerolog.Disabled),
				standard.WithChainTimeService(chainTime),
				standard.WithFeeRecipientProvider(mockFeeRecipientProvider),
				standard.WithProposalPreparationsSubmitter(mockProposalPreparationsSubmitter),
			},
			err: "problem with parameters: no validating accounts provider specified",
		},
		{
			name: "FeeRecipientProviderMissing",
			params: []standard.Parameter{
				standard.WithLogLevel(zerolog.Disabled),
				standard.WithChainTimeService(chainTime),
				standard.WithValidatingAccountsProvider(mockValidatingAccountsProvider),
				standard.WithProposalPreparationsSubmitter(mockProposalPreparationsSubmitter),
			},
			err: "problem with parameters: no fee recipient provider specified",
		},
		{
			name: "ProposalPreparationsSubmitterMissing",
			params: []standard.Parameter{
				standard.WithLogLevel(zerolog.Disabled),
				standard.WithChainTimeService(chainTime),
				standard.WithValidatingAccountsProvider(mockValidatingAccountsProvider),
				standard.WithFeeRecipientProvider(mockFeeRecipientProvider),
			},
			err: "problem with parameters: no proposal preparations submitter specified",
		},
		{
			name: "Good",
			params: []standard.Parameter{
				standard.WithMonitor(prometheusMetrics),
				standard.WithLogLevel(zerolog.Disabled),
				standard.WithChainTimeService(chainTime),
				standard.WithValidatingAccountsProvider(mockValidatingAccountsProvider),
				standard.WithFeeRecipientProvider(mockFeeRecipientProvider),
				standard.WithProposalPreparationsSubmitter(mockProposalPreparationsSubmitter),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capture := logger.NewLogCapture()
			_, err := standard.New(ctx, test.params...)
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
