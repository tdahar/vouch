// Copyright © 2020 Attestant Limited.
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

package standard

import (
	"context"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

// ValidatorsByPubKey fetches the requested validators from local store given their public keys.
func (s *Service) ValidatorsByPubKey(_ context.Context, pubKeys []phase0.BLSPubKey) map[phase0.ValidatorIndex]*phase0.Validator {
	res := make(map[phase0.ValidatorIndex]*phase0.Validator)
	s.validatorsMutex.RLock()
	for _, pubKey := range pubKeys {
		if validator, exists := s.validatorsByPubKey[pubKey]; exists {
			res[s.validatorPubKeyToIndex[pubKey]] = validator
		}
	}
	s.validatorsMutex.RUnlock()

	return res
}
