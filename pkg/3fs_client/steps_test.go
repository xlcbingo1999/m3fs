package fsclient

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/config"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestUmountHostMountpointSuite(t *testing.T) {
	suiteRun(t, &umountHostMountpointSuite{})
}

type umountHostMountpointSuite struct {
	ttask.StepSuite

	step *umountHostMountponitStep
}

func (s *umountHostMountpointSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.Cfg.Services.Client.HostMountpoint = "/mnt/3fs"
	s.SetupRuntime()
	s.step = &umountHostMountponitStep{}
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
}

func (s *umountHostMountpointSuite) Test() {
	s.MockRunner.On("Exec", "mount", []string(nil)).Return("/mnt/3fs", nil)
	s.MockRunner.On("Exec", "umount", []string{"/mnt/3fs"}).Return("", nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
}

func (s *umountHostMountpointSuite) TestWithNotMount() {
	s.MockRunner.On("Exec", "mount", []string(nil)).Return("", nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
}
