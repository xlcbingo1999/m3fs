package fdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
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
	s.Runtime.Nodes = make(map[string]config.Node, len(s.Cfg.Nodes))
	for _, node := range s.Cfg.Nodes {
		s.Runtime.Nodes[node.Name] = node
	}
	s.Cfg.Services.Fdb.Nodes = []string{"node1", "node2"}
	s.Cfg.Services.Fdb.Port = 4500
	s.step.Init(s.Runtime, s.MockedEm)
}

func (s *genClusterFileContentStepSuite) TestGenClusterFileContentStep() {
	s.NoError(s.step.Execute(s.Ctx()))

	contentI, ok := s.Runtime.Load("fdb_cluster_file_content")
	s.True(ok)
	s.Equal("node1:node1@1.1.1.1:4500,node2:node2@1.1.1.2:4500", contentI.(string))
}

func TestStartContainerStep(t *testing.T) {
	suiteRun(t, &startContainerStepSuite{})
}

type startContainerStepSuite struct {
	ttask.StepSuite

	step *startContainerStep
}

func (s *startContainerStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &startContainerStep{}
	s.step.Init(s.Runtime, s.MockedEm)
	s.Runtime.Store("fdb_cluster_file_content", "xxxx")
}

func (s *startContainerStepSuite) TestStartContainerStep() {
	dataDir := "/root/3fs/fdb/data"
	logDir := "/root/3fs/fdb/log"
	s.MockOS.On("Exec", "mkdir", []string{"-p", dataDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", logDir}).Return("", nil)
	img, err := image.GetImage("", "fdb")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:       img,
		Name:        common.Pointer("3fs-fdb"),
		HostNetwork: true,
		Envs: map[string]string{
			"FDB_CLUSTER_FILE_CONTENTS": "xxxx",
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: dataDir,
				Target: "/var/fdb/data",
			},
			{
				Source: logDir,
				Target: "/var/fdb/log",
			},
		},
	}).Return(new(bytes.Buffer), nil)

	s.NotNil(s.step)
	s.NoError(s.step.Execute(s.Ctx()))

	s.MockOS.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}
