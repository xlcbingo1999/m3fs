package external

import (
	"os"

	"github.com/stretchr/testify/mock"

	"github.com/open3fs/m3fs/pkg/external"
)

// MockLocal is an mock type for the LocalInterface
type MockLocal struct {
	mock.Mock
	external.LocalInterface
}

// MkdirTemp mock.
func (m *MockLocal) MkdirTemp(dir, prefix string) (string, error) {
	arg := m.Called(dir, prefix)
	return arg.String(0), arg.Error(1)
}

// MkdirAll mock.
func (m *MockLocal) MkdirAll(path string) error {
	return m.Called(path).Error(0)
}

// RemoveAll mock.
func (m *MockLocal) RemoveAll(path string) error {
	return m.Called(path).Error(0)
}

// WriteFile mock.
func (m *MockLocal) WriteFile(path string, data []byte, perm os.FileMode) error {
	return m.Called(path, data, perm).Error(0)
}
