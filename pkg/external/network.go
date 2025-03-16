// Copyright 2025 Open3FS Authors
//
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

package external

// NetInterface provides interface about network.
type NetInterface interface {
	GetRdmaLinks() ([]string, error)
}

type netExternal struct {
	externalBase
}

func (ne *netExternal) init(em *Manager) {
	ne.externalBase.init(em)
	em.Net = ne
}

func (ne *netExternal) GetRdmaLinks() ([]string, error) {
	// TODO: add get list of RDMA links logic
	return nil, nil
}

func init() {
	registerNewExternalFunc(func() externalInterface {
		return new(netExternal)
	})
}
