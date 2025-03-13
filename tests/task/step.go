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

	Cfg         *config.Config
	Runtime     *task.Runtime
	MockEm      *external.Manager
	MockRunner  *texternal.MockRunner
	MockOS      *texternal.MockOS
	MockDocker  *texternal.MockDocker
	MockedLocal *texternal.MockLocal
}

// SetupTest runs before each test in the step suite.
func (s *StepSuite) SetupTest() {
	s.Suite.SetupTest()

	testConfig := `name: test-cluster
networktype: RDMA
services:
  fdb:
    containerName: 3fs-fdb
    workDir: "/root/3fs/fdb"
    port: 4500
  clickhouse:
    containerName: 3fs-clickhouse
    workDir: "/root/3fs/clickhouse"
    db: 3fs
    user: default
    password: password
    tcpPort: 9000
  mgmtd:
    containerName: 3fs-mgmtd
  meta:
    containerName: 3fs-meta
  storage:
    containerName: 3fs-storage
    disktype: "NVMe"
  client:
    containerName: 3fs-fuseclient
registry:
  customRegistry: ""
  `
	s.Cfg = new(config.Config)
	s.YamlUnmarshal([]byte(testConfig), s.Cfg)

	s.MockRunner = new(texternal.MockRunner)
	s.MockOS = new(texternal.MockOS)
	s.MockDocker = new(texternal.MockDocker)
	s.MockedLocal = new(texternal.MockLocal)
	s.MockEm = &external.Manager{
		Runner: s.MockRunner,
		OS:     s.MockOS,
		Docker: s.MockDocker,
		Local:  s.MockedLocal,
	}

	s.SetupRuntime()
}

// SetupRuntime setup runtime with the test config.
func (s *StepSuite) SetupRuntime() {
	s.Runtime = &task.Runtime{
		Cfg:      s.Cfg,
		Services: &s.Cfg.Services,
	}
	s.Runtime.Nodes = make(map[string]config.Node, len(s.Cfg.Nodes))
	for _, node := range s.Cfg.Nodes {
		s.Runtime.Nodes[node.Name] = node
	}
	s.Runtime.Services = &s.Cfg.Services
}
