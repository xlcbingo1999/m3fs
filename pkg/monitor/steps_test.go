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
	s.step.Init(s.Runtime, s.MockedEm)
}

func (s *genMonitorConfigStepSuite) Test() {
	s.NoError(s.step.Execute(s.Ctx()))

	tmpPathValue, ok := s.Runtime.Load("tmp_monitor_collector_main_path")
	s.True(ok)
	tmpPath := tmpPathValue.(string)
	_, err := os.ReadFile(tmpPath)
	s.NoError(err)

	s.Contains(tmpPath, "s3fs-monitor.")
	s.NoError(os.RemoveAll(filepath.Dir(tmpPath)))
}
