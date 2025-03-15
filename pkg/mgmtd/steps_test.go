package mgmtd

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenAdminCliConfigSuite(t *testing.T) {
	suiteRun(t, &genAdminCliConfigStepSuite{})
}

type genAdminCliConfigStepSuite struct {
	ttask.StepSuite

	step *genAdminCliConfigStep
}

func (s *genAdminCliConfigStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.Cfg.Name = "test-cluster"
	s.SetupRuntime()
	s.step = &genAdminCliConfigStep{}
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
}

func (s *genAdminCliConfigStepSuite) Test() {
	s.NoError(s.step.Execute(s.Ctx()))

	dataI, ok := s.Runtime.Load(task.RuntimeAdminCliTomlKey)
	s.True(ok)
	s.Equal([]byte(`cluster_id = "test-cluster"

[fdb]
clusterFile = '/opt/3fs/etc/fdb.cluster'`), dataI.([]byte))
}

func TestInitClusterStepSuite(t *testing.T) {
	suiteRun(t, &initClusterStepSuite{})
}

type initClusterStepSuite struct {
	ttask.StepSuite

	step      *initClusterStep
	configDir string
}

func (s *initClusterStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &initClusterStep{}
	s.Cfg.Services.Mgmtd.WorkDir = "/var/mgmtd"
	s.configDir = "/var/mgmtd/config.d"
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm, config.Node{})
	s.Runtime.Store(task.RuntimeFdbClusterFileContentKey, "xxxx")
}

func (s *initClusterStepSuite) TestInitCluster() {
	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, "3fs")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:      img,
		Name:       &s.Cfg.Services.Mgmtd.ContainerName,
		Entrypoint: common.Pointer("''"),
		Rm:         common.Pointer(true),
		Command: []string{
			"/opt/3fs/bin/admin_cli",
			"-cfg", "/opt/3fs/etc/admin_cli.toml",
			"'init-cluster --mgmtd /opt/3fs/etc/mgmtd_main.toml 1 1048576 16'",
		},
		HostNetwork: true,
		Volumes: []*external.VolumeArgs{
			{
				Source: s.configDir,
				Target: "/opt/3fs/etc",
			},
		},
	}).Return("", nil)

	s.NoError(s.step.Execute(s.Ctx()))

	s.MockRunner.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}
