package task

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/external"
)

// defines keys of runtime cache.
const (
	RuntimeFdbClusterFileContentKey = "fdb_cluster_file_content"
)

// Runtime contains task run info
type Runtime struct {
	sync.Map
	Cfg      *config.Config
	Nodes    map[string]config.Node
	Services *config.Services
	WorkDir  string
	LocalEm  *external.Manager
}

// Runner is a task runner.
type Runner struct {
	tasks []Interface
	cfg   *config.Config
	init  bool
}

// Init initializes all tasks.
func (r *Runner) Init() {
	runtime := &Runtime{Cfg: r.cfg, WorkDir: r.cfg.WorkDir}
	runtime.Nodes = make(map[string]config.Node, len(r.cfg.Nodes))
	for _, node := range r.cfg.Nodes {
		runtime.Nodes[node.Name] = node
	}
	runtime.Services = &r.cfg.Services
	em := external.NewManager(external.NewLocalRunner(&external.LocalRunnerCfg{
		Logger:         logrus.StandardLogger(),
		MaxExitTimeout: r.cfg.CmdMaxExitTimout,
	}))
	runtime.LocalEm = em

	for _, task := range r.tasks {
		task.Init(runtime)
	}
	r.init = true
}

// Register registers tasks.
func (r *Runner) Register(task ...Interface) error {
	if r.init {
		return errors.New("runner has been initialized")
	}
	r.tasks = append(r.tasks, task...)
	return nil
}

// Run runs all tasks.
func (r *Runner) Run(ctx context.Context) error {
	for _, task := range r.tasks {
		logrus.Infof("Running task %s", task.Name())
		if err := task.Run(ctx); err != nil {
			return errors.Annotatef(err, "run task %s", task.Name())
		}
	}

	return nil
}

// NewRunner creates a new task runner.
func NewRunner(cfg *config.Config, tasks ...Interface) *Runner {
	return &Runner{
		tasks: tasks,
		cfg:   cfg,
	}
}
