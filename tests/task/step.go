package task

import (
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/task"
	"github.com/open3fs/m3fs/tests/base"
	texternal "github.com/open3fs/m3fs/tests/external"
)

// StepSuite is the base Suite for all step suites.
type StepSuite struct {
	base.Suite

	Cfg        *config.Config
	Runtime    *task.Runtime
	MockEm     *external.Manager
	MockRunner *texternal.MockRunner
	MockDocker *texternal.MockDocker
	// NOTE: external.FSInterface is not implemented for remote runner.
	// MockFS          *texternal.MockFS
	MockLocalEm     *external.Manager
	MockLocalRunner *texternal.MockRunner
	MockLocalFS     *texternal.MockFS
	MockLocalDocker *texternal.MockDocker
}

// SetupTest runs before each test in the step suite.
func (s *StepSuite) SetupTest() {
	s.Suite.SetupTest()

	s.Cfg = config.NewConfigWithDefaults()
	s.Cfg.Name = "test-cluster"
	s.Cfg.WorkDir = "/root/3fs"

	s.MockRunner = new(texternal.MockRunner)
	s.MockDocker = new(texternal.MockDocker)
	s.MockEm = &external.Manager{
		Runner: s.MockRunner,
		Docker: s.MockDocker,
	}

	s.MockLocalDocker = new(texternal.MockDocker)
	s.MockLocalRunner = new(texternal.MockRunner)
	s.MockLocalFS = new(texternal.MockFS)
	s.MockLocalEm = &external.Manager{
		Runner: s.MockLocalRunner,
		FS:     s.MockLocalFS,
		Docker: s.MockLocalDocker,
	}

	s.SetupRuntime()
}

// SetupRuntime setup runtime with the test config.
func (s *StepSuite) SetupRuntime() {
	s.Runtime = &task.Runtime{
		Cfg:      s.Cfg,
		WorkDir:  s.Cfg.WorkDir,
		Services: &s.Cfg.Services,
		LocalEm:  s.MockLocalEm,
	}
	s.Runtime.Nodes = make(map[string]config.Node, len(s.Cfg.Nodes))
	for _, node := range s.Cfg.Nodes {
		s.Runtime.Nodes[node.Name] = node
	}
	s.Runtime.Services = &s.Cfg.Services
}
