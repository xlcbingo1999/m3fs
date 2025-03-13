package clickhouse

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenClickhouseConfigStep(t *testing.T) {
	suiteRun(t, &genClickhouseConfigStepSuite{})
}

type genClickhouseConfigStepSuite struct {
	ttask.StepSuite

	step *genClickhouseConfigStep
}

func (s *genClickhouseConfigStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &genClickhouseConfigStep{}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm)
}

func (s *genClickhouseConfigStepSuite) Test() {
	s.NoError(s.step.Execute(s.Ctx()))

	tmpDirValue, ok := s.Runtime.Load("clickhouse_temp_config_dir")
	s.True(ok)
	tmpDir := tmpDirValue.(string)
	s.Contains(tmpDir, "3fs-clickhouse.")
	_, err := os.ReadFile(filepath.Join(tmpDir, "config.xml"))
	s.NoError(err)
	_, err = os.ReadFile(filepath.Join(tmpDir, "3fs-monitor.sql"))
	s.NoError(err)
	s.NoError(os.RemoveAll(tmpDir))
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
	s.step.Init(s.Runtime, s.MockEm)
	s.Runtime.Store("clickhouse_temp_config_dir", "/tmp/3f-clickhouse.xxx")
}

func (s *startContainerStepSuite) TestStartContainerStep() {
	dataDir := "/root/3fs/clickhouse/data"
	logDir := "/root/3fs/clickhouse/log"
	configDir := "/root/3fs/clickhouse/config.d"
	sqlDir := "/root/3fs/clickhouse/sql"
	s.MockOS.On("Exec", "mkdir", []string{"-p", dataDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", logDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", configDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", sqlDir}).Return("", nil)
	s.MockRunner.On("Scp", "/tmp/3f-clickhouse.xxx/config.xml",
		"/root/3fs/clickhouse/config.d/config.xml").Return(nil)
	s.MockRunner.On("Scp", "/tmp/3f-clickhouse.xxx/3fs-monitor.sql",
		"/root/3fs/clickhouse/sql/3fs-monitor.sql").Return(nil)
	img, err := image.GetImage("", "clickhouse")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:       img,
		Name:        common.Pointer("3fs-clickhouse"),
		HostNetwork: true,
		Envs: map[string]string{
			"CLICKHOUSE_DB":       "3fs",
			"CLICKHOUSE_USER":     "default",
			"CLICKHOUSE_PASSWORD": "password",
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: dataDir,
				Target: "/var/lib/clickhouse",
			},
			{
				Source: logDir,
				Target: "/var/log/clickhouse-server",
			},
			{
				Source: configDir,
				Target: "/etc/clickhouse-server/config.d",
			},
			{
				Source: sqlDir,
				Target: "/tmp/sql",
			},
		},
	}).Return(new(bytes.Buffer), nil)

	s.NotNil(s.step)
	s.NoError(s.step.Execute(s.Ctx()))

	s.MockOS.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}
