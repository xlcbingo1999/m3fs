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

import (
	"os"

	"github.com/open3fs/m3fs/pkg/errors"
)

// FSInterface provides interface about local fs, this is not implemented for remote runner.
type FSInterface interface {
	MkdirTemp(string, string) (string, error)
	MkdirAll(string) error
	RemoveAll(string) error
	WriteFile(string, []byte, os.FileMode) error
}

type fsExternal struct {
	externalBase

	returnUnimplemented bool
}

func (fe *fsExternal) init(em *Manager) {
	fe.externalBase.init(em)
	em.FS = fe
	if _, ok := em.Runner.(*RemoteRunner); ok {
		fe.returnUnimplemented = true
	}
}

func (fe *fsExternal) MkdirTemp(dir, prefix string) (string, error) {
	if fe.returnUnimplemented {
		return "", errors.New("unimplemented")
	}
	name, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		return "", errors.Trace(err)
	}
	return name, nil
}

func (fe *fsExternal) MkdirAll(dir string) error {
	if fe.returnUnimplemented {
		return errors.New("unimplemented")
	}
	return errors.Trace(os.MkdirAll(dir, 0755))
}

func (fe *fsExternal) WriteFile(path string, data []byte, perm os.FileMode) error {
	if fe.returnUnimplemented {
		return errors.New("unimplemented")
	}
	return errors.Trace(os.WriteFile(path, data, perm))
}

func (fe *fsExternal) RemoveAll(dir string) error {
	if fe.returnUnimplemented {
		return errors.New("unimplemented")
	}
	return errors.Trace(os.RemoveAll(dir))
}

func init() {
	registerNewExternalFunc(func() externalInterface {
		return new(fsExternal)
	})
}
