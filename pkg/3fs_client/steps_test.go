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
