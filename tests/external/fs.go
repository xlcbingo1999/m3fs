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

	"github.com/stretchr/testify/mock"

	"github.com/open3fs/m3fs/pkg/external"
)

// MockFS is an mock type for the FSInterface
type MockFS struct {
	mock.Mock
	external.FSInterface
}

// MkdirTemp mock.
func (m *MockFS) MkdirTemp(dir, prefix string) (string, error) {
	arg := m.Called(dir, prefix)
	return arg.String(0), arg.Error(1)
}

// MkdirAll mock.
func (m *MockFS) MkdirAll(path string) error {
	return m.Called(path).Error(0)
}

// RemoveAll mock.
func (m *MockFS) RemoveAll(path string) error {
	return m.Called(path).Error(0)
}

// WriteFile mock.
func (m *MockFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	return m.Called(path, data, perm).Error(0)
}
