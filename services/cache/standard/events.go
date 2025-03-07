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

package standard

import (
	"fmt"

	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
)

// handleBlock handles a block update message.
func (s *Service) handleBlock(event *apiv1.Event) {
	if event.Data == nil {
		return
	}

	data := event.Data.(*apiv1.BlockEvent)
	log.Trace().Str("root", fmt.Sprintf("%#x", data.Block)).Uint64("slot", uint64(data.Slot)).Msg("Received block event")

	s.setBlockRootToSlot(data.Block, data.Slot)
}
