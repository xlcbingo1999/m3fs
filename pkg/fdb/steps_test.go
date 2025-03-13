package fdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenClusterFileContentStep(t *testing.T) {
	suiteRun(t, &genClusterFileContentStepSuite{})
}

type genClusterFileContentStepSuite struct {
	ttask.StepSuite

	step *genClusterFileContentStep
}

func (s *genClusterFileContentStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &genClusterFileContentStep{}
	s.Cfg.Nodes = []config.Node{
		{
			Name: "node1",
			Host: "1.1.1.1",
		},
		{
			Name: "node2",
			Host: "1.1.1.2",
		},
	}
	s.Cfg.Services.Fdb.Nodes = []string{"node1", "node2"}
	s.Cfg.Services.Fdb.Port = 4500
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, s.Cfg.Nodes[0])
}

func (s *genClusterFileContentStepSuite) TestGenClusterFileContentStep() {
	s.NoError(s.step.Execute(s.Ctx()))

	contentI, ok := s.Runtime.Load(task.RuntimeFdbClusterFileContentKey)
	s.True(ok)
	s.Equal("test-cluster:test-cluster@1.1.1.1:4500,1.1.1.2:4500", contentI.(string))
}

func TestRuntContainerStep(t *testing.T) {
	suiteRun(t, &runContainerStepSuite{})
}

type runContainerStepSuite struct {
	ttask.StepSuite

	step    *runContainerStep
	dataDir string
	logDir  string
}

func (s *runContainerStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &runContainerStep{}
	s.Cfg.Services.Fdb.WorkDir = "/var/fdb"
	s.dataDir = "/var/fdb/data"
	s.logDir = "/var/fdb/logs"
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
	s.Runtime.Store(task.RuntimeFdbClusterFileContentKey, "xxxx")
}

func (s *runContainerStepSuite) TestRunContainerStep() {
	s.MockRunner.On("Exec", "mkdir", []string{"-p", s.dataDir}).Return(new(bytes.Buffer), nil)
	s.MockRunner.On("Exec", "mkdir", []string{"-p", s.logDir}).Return(new(bytes.Buffer), nil)
	img, err := image.GetImage("", "fdb")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:       img,
		Name:        &s.Cfg.Services.Fdb.ContainerName,
		HostNetwork: true,
		Detach:      common.Pointer(true),
		Envs: map[string]string{
			"FDB_CLUSTER_FILE_CONTENTS": "xxxx",
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: s.dataDir,
				Target: "/var/fdb/data",
			},
			{
				Source: s.logDir,
				Target: "/var/fdb/logs",
			},
		},
	}).Return(new(bytes.Buffer), nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}

func (s *runContainerStepSuite) TestRunContainerFailed() {
	s.MockRunner.On("Exec", "mkdir", []string{"-p", s.dataDir}).Return(new(bytes.Buffer), nil)
	s.MockRunner.On("Exec", "mkdir", []string{"-p", s.logDir}).Return(new(bytes.Buffer), nil)
	img, err := image.GetImage("", "fdb")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:       img,
		Name:        &s.Cfg.Services.Fdb.ContainerName,
		HostNetwork: true,
		Detach:      common.Pointer(true),
		Envs: map[string]string{
			"FDB_CLUSTER_FILE_CONTENTS": "xxxx",
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: s.dataDir,
				Target: "/var/fdb/data",
			},
			{
				Source: s.logDir,
				Target: "/var/fdb/logs",
			},
		},
	}).Return(nil, errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockRunner.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}

func (s *runContainerStepSuite) TestRunDirFailed() {
	s.MockRunner.On("Exec", "mkdir", []string{"-p", s.dataDir}).Return(new(bytes.Buffer), nil)
	s.MockRunner.On("Exec", "mkdir", []string{"-p", s.logDir}).Return(new(bytes.Buffer), errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockRunner.AssertExpectations(s.T())
}

func TestInitClusterStepSuite(t *testing.T) {
	suiteRun(t, &initClusterStepSuite{})
}

type initClusterStepSuite struct {
	ttask.StepSuite

	step *initClusterStep
}

func (s *initClusterStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &initClusterStep{}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
	s.Runtime.Store(task.RuntimeFdbClusterFileContentKey, "xxxx")
}

func (s *initClusterStepSuite) TestInit() {
	s.MockDocker.On("Exec", s.Runtime.Services.Fdb.ContainerName,
		"fdbcli", []string{"--exec", "'configure new single ssd'"}).
		Return(new(bytes.Buffer), nil)
	s.MockDocker.On("Exec", s.Runtime.Services.Fdb.ContainerName,
		"fdbcli", []string{"--exec", "'status minimal'"}).
		Return(bytes.NewBufferString("The database is available."), nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockDocker.AssertExpectations(s.T())
}

func (s *initClusterStepSuite) TestInitClusterFailed() {
	s.MockDocker.On("Exec", s.Runtime.Services.Fdb.ContainerName,
		"fdbcli", []string{"--exec", "'configure new single ssd'"}).
		Return(nil, errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockDocker.AssertExpectations(s.T())
}

func (s *initClusterStepSuite) TestWaitClusterInitializedFailed() {
	s.MockDocker.On("Exec", s.Runtime.Services.Fdb.ContainerName,
		"fdbcli", []string{"--exec", "'configure new single ssd'"}).
		Return(new(bytes.Buffer), nil)
	s.MockDocker.On("Exec", s.Runtime.Services.Fdb.ContainerName,
		"fdbcli", []string{"--exec", "'status minimal'"}).
		Return(nil, errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockDocker.AssertExpectations(s.T())
}

func TestRmContainerStep(t *testing.T) {
	suiteRun(t, &rmContainerStepSuite{})
}

type rmContainerStepSuite struct {
	ttask.StepSuite

	step    *rmContainerStep
	dataDir string
	logDir  string
}

func (s *rmContainerStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &rmContainerStep{}
	s.Cfg.Services.Fdb.WorkDir = "/var/fdb"
	s.dataDir = "/var/fdb/data"
	s.logDir = "/var/fdb/logs"
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
	s.Runtime.Store(task.RuntimeFdbClusterFileContentKey, "xxxx")
}

func (s *rmContainerStepSuite) TestRmContainerStep() {
	s.MockDocker.On("Rm", s.Cfg.Services.Fdb.ContainerName, true).
		Return(new(bytes.Buffer), nil)
	s.MockRunner.On("Exec", "rm", []string{"-rf", s.dataDir}).Return(new(bytes.Buffer), nil)
	s.MockRunner.On("Exec", "rm", []string{"-rf", s.logDir}).Return(new(bytes.Buffer), nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}

func (s *rmContainerStepSuite) TestRmContainerFailed() {
	s.MockDocker.On("Rm", s.Cfg.Services.Fdb.ContainerName, true).
		Return(new(bytes.Buffer), errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockDocker.AssertExpectations(s.T())
}

func (s *rmContainerStepSuite) TestRmDirFailed() {
	s.MockDocker.On("Rm", s.Cfg.Services.Fdb.ContainerName, true).
		Return(new(bytes.Buffer), nil)
	s.MockRunner.On("Exec", "rm", []string{"-rf", s.dataDir}).
		Return(new(bytes.Buffer), errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockRunner.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}
