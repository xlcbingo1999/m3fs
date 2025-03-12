package clickhouse

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	ttask "github.com/open3fs/m3fs/tests/task"
)

var suiteRun = suite.Run

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
}

func (s *startContainerStepSuite) TestStartContainerStep() {
	dataDir := "/root/3fs/clickhouse/data"
	logDir := "/root/3fs/clickhouse/log"
	s.MockOS.On("Exec", "mkdir", []string{"-p", dataDir}).Return("", nil)
	s.MockOS.On("Exec", "mkdir", []string{"-p", logDir}).Return("", nil)
	img, err := image.GetImage("", "clickhouse")
	s.NoError(err)
	s.MockDocker.On("Run", &external.RunArgs{
		Image:       img,
		Name:        common.Pointer("3fs-clickhouse"),
		HostNetwork: true,
		Volumes: []*external.VolumeArgs{
			{
				Source: dataDir,
				Target: "/var/lib/clickhouse",
			},
			{
				Source: logDir,
				Target: "/var/log/clickhouse-server",
			},
		},
	}).Return(new(bytes.Buffer), nil)

	s.NotNil(s.step)
	s.NoError(s.step.Execute(s.Ctx()))

	s.MockOS.AssertExpectations(s.T())
	s.MockDocker.AssertExpectations(s.T())
}
