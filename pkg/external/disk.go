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

// DiskInterface provides interface about disk.
type DiskInterface interface {
	GetNvmeDisks() ([]string, error)
}

type diskExternal struct {
	externalBase
}

func (de *diskExternal) init(em *Manager) {
	de.externalBase.init(em)
	em.Disk = de
}

func (de *diskExternal) GetNvmeDisks() ([]string, error) {
	// TODO: implement GetNvmeDisks
	return nil, nil
}

func init() {
	registerNewExternalFunc(func() externalInterface {
		return new(diskExternal)
	})
}
