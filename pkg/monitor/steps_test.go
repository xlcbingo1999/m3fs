package monitor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

func TestGenMonitorConfigStep(t *testing.T) {
	suiteRun(t, &genMonitorConfigStepSuite{})
}

type genMonitorConfigStepSuite struct {
	ttask.StepSuite

	step *genMonitorConfigStep
}

func (s *genMonitorConfigStepSuite) SetupTest() {
	s.StepSuite.SetupTest()

	s.step = &genMonitorConfigStep{}
	s.SetupRuntime()
	s.step.Init(s.Runtime, s.MockEm)
}

func (s *genMonitorConfigStepSuite) Test() {
	s.NoError(s.step.Execute(s.Ctx()))

	tmpDirValue, ok := s.Runtime.Load("monitor_temp_config_dir")
	s.True(ok)
	tmpDir := tmpDirValue.(string)
	s.Contains(tmpDir, "3fs-monitor.")
	_, err := os.ReadFile(filepath.Join(tmpDir, "monitor_collector_main.toml"))
	s.NoError(err)
	s.NoError(os.RemoveAll(tmpDir))
}
