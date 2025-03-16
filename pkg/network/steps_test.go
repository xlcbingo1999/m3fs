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

package network

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/config"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenIbdev2netdevScriptStep(t *testing.T) {
	suiteRun(t, &genIbdev2netdevScriptStepSuite{})
}

type genIbdev2netdevScriptStepSuite struct {
	ttask.StepSuite

	step *genIbdev2netdevScriptStep
}

func (s *genIbdev2netdevScriptStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &genIbdev2netdevScriptStep{}
	s.Cfg.Nodes = []config.Node{
		{
			Name: "node1",
			Host: "1.1.1.1",
		},
	}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, s.Cfg.Nodes[0])
}

func (s *genIbdev2netdevScriptStepSuite) TestGenIbdev2netdevScript() {
	tmpDir := "/tmp/m3fs-prepare-network.123"
	s.MockLocalFS.On("MkdirTemp", "/tmp", "m3fs-prepare-network").Return(tmpDir, nil)
	scriptPath := tmpDir + "/ibdev2netdev"
	s.MockLocalFS.On("WriteFile", scriptPath,
		[]byte(ibdev2netdevScript), os.FileMode(0755)).Return(nil)

	binDir := s.Cfg.WorkDir + "/bin"
	remoteScriptPath := binDir + "/ibdev2netdev"
	s.MockRunner.On("Exec", "mkdir", []string{"-p", binDir}).Return("", nil)
	s.MockRunner.On("Scp", scriptPath, remoteScriptPath).Return(nil)
	s.MockRunner.On("Exec", "chmod", []string{"+x", remoteScriptPath}).Return("", nil)

	s.MockLocalFS.On("RemoveAll", tmpDir).Return(nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockLocalFS.AssertExpectations(s.T())
	s.MockRunner.AssertExpectations(s.T())
}

func TestLoadRdmaRxeModule(t *testing.T) {
	suiteRun(t, &loadRdmaRxeModuleStepSuite{})
}

type loadRdmaRxeModuleStepSuite struct {
	ttask.StepSuite

	step *loadRdmaRxeModuleStep
}

func (s *loadRdmaRxeModuleStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &loadRdmaRxeModuleStep{}
	s.Cfg.Nodes = []config.Node{
		{
			Name: "node1",
			Host: "1.1.1.1",
		},
	}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, s.Cfg.Nodes[0])
}

func (s *loadRdmaRxeModuleStepSuite) TestLoadRdmaRxeModule() {
	s.MockRunner.On("Exec", "modprobe", []string{"rdma_rxe"}).Return("", nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
}

func TestCreateRdmaRxeLinkStep(t *testing.T) {
	suiteRun(t, &createRdmaRxeLinkStepSuite{})
}

type createRdmaRxeLinkStepSuite struct {
	ttask.StepSuite

	step *createRdmaRxeLinkStep
}

func (s *createRdmaRxeLinkStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &createRdmaRxeLinkStep{}
	s.Cfg.Nodes = []config.Node{
		{
			Name: "node1",
			Host: "1.1.1.1",
		},
	}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, s.Cfg.Nodes[0])
}

func (s *createRdmaRxeLinkStepSuite) TestCreateRdmaRxeLinkStep() {
	tmpDir := "/tmp/m3fs-prepare-network.123"
	s.MockLocalFS.On("MkdirTemp", "/tmp", "m3fs-prepare-network").Return(tmpDir, nil)
	scriptPath := tmpDir + "/create_rdma_rxe_link"
	s.MockLocalFS.On("WriteFile", scriptPath,
		[]byte(createRdmaLinkScript), os.FileMode(0755)).Return(nil)

	binDir := s.Cfg.WorkDir + "/bin"
	remoteScriptPath := binDir + "/create_rdma_rxe_link"
	s.MockRunner.On("Exec", "mkdir", []string{"-p", binDir}).Return("", nil)
	s.MockRunner.On("Scp", scriptPath, remoteScriptPath).Return(nil)
	s.MockRunner.On("Exec", "chmod", []string{"+x", remoteScriptPath}).Return("", nil)
	s.MockRunner.On("Exec", "bash", []string{remoteScriptPath}).Return("", nil)

	s.MockLocalFS.On("RemoveAll", tmpDir).Return(nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockLocalFS.AssertExpectations(s.T())
	s.MockRunner.AssertExpectations(s.T())
}
