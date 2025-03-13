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
	s.step.Init(s.Runtime, s.MockedEm)
}

func (s *genClusterFileContentStepSuite) TestGenClusterFileContentStep() {
	s.NoError(s.step.Execute(s.Ctx()))

	contentI, ok := s.Runtime.Load("fdb_cluster_file_content")
	s.True(ok)
	s.Equal("test-cluster:test-cluster@1.1.1.1:4500,1.1.1.2:4500", contentI.(string))
}

func TestStartContainerStep(t *testing.T) {
	suiteRun(t, &startContainerStepSuite{})
}

type startContainerStepSuite struct {
	ttask.StepSuite

	step    *startContainerStep
	dataDir string
	logDir  string
}

func (s *startContainerStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &startContainerStep{}
	s.Cfg.Services.Fdb.WorkDir = "/var/fdb"
	s.dataDir = "/var/fdb/data"
	s.logDir = "/var/fdb/log"
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockedEm)
	s.Runtime.Store("fdb_cluster_file_content", "xxxx")
}

func (s *startContainerStepSuite) TestStartContainerStep() {
	s.MockOS.On("Exec", "mkdir", []string{"-p", s.dataDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", s.logDir}).Return("", nil)
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
				Target: "/var/fdb/log",
			},
		},
	}).Return(new(bytes.Buffer), nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockOS.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}

func (s *startContainerStepSuite) TestStartContainerFailed() {
	s.MockOS.On("Exec", "mkdir", []string{"-p", s.dataDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", s.logDir}).Return("", nil)
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
				Target: "/var/fdb/log",
			},
		},
	}).Return(nil, errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockOS.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}

func (s *startContainerStepSuite) TestCreateDirFailed() {
	s.MockOS.On("Exec", "mkdir", []string{"-p", s.dataDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", s.logDir}).Return("", errors.New("dummy error"))

	s.Error(s.step.Execute(s.Ctx()), "dummy error")

	s.MockOS.AssertExpectations(s.T())
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
	s.Cfg.Services.Fdb.WorkDir = "/var/fdb"
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockedEm)
	s.Runtime.Store("fdb_cluster_file_content", "xxxx")
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
