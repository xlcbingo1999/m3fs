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
