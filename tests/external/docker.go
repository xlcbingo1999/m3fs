package external

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/open3fs/m3fs/pkg/external"
)

// MockDocker is an mock type for the DockerInterface
type MockDocker struct {
	mock.Mock
	external.DockerInterface
}

// Run mock.
func (m *MockDocker) Run(ctx context.Context, args *external.RunArgs) (string, error) {
	arg := m.Called(args)
	err1 := arg.Error(1)
	if err1 != nil {
		return "", err1
	}
	return arg.String(0), nil
}

// Rm mock.
func (m *MockDocker) Rm(ctx context.Context, name string, force bool) (string, error) {
	arg := m.Called(name, force)
	err1 := arg.Error(1)
	if err1 != nil {
		return "", err1
	}
	return arg.String(0), nil
}

// Exec mock.
func (m *MockDocker) Exec(ctx context.Context, container, cmd string, args ...string) (
	string, error) {

	arg := m.Called(container, cmd, args)
	err1 := arg.Error(1)
	if err1 != nil {
		return "", err1
	}
	return arg.String(0), nil
}
