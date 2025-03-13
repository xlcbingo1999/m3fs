package external

import (
	"bytes"
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/open3fs/m3fs/pkg/external"
)

// MockRunner is an mock type for the RunnerInterface
type MockRunner struct {
	mock.Mock
	external.RunnerInterface
}

// Exec mock.
func (m *MockRunner) Exec(ctx context.Context, cmd string, args ...string) (
	out *bytes.Buffer, err error) {

	arg := m.Called(cmd, args)
	err1 := arg.Error(1)
	if err1 != nil {
		return nil, err1
	}
	return arg.Get(0).(*bytes.Buffer), nil
}

// Scp mock.
func (m *MockRunner) Scp(local, remote string) error {
	arg := m.Called(local, remote)
	return arg.Error(0)
}
