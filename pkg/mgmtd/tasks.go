package mgmtd

import (
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/task"
)

// CreateMgmtdServiceTask is a task for creating 3fs mgmtd services.
type CreateMgmtdServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *CreateMgmtdServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("CreateMgmtdServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Mgmtd.Nodes))
	for i, node := range r.Cfg.Services.Mgmtd.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(genNodeIDStep) },
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep:  func() task.Step { return new(prepareMgmtdConfigStep) },
		},
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(initClusterStep) },
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep:  func() task.Step { return new(runContainerStep) },
		},
	})
}

// DeleteMgmtdServiceTask is a task for deleting a mgmtd services.
type DeleteMgmtdServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *DeleteMgmtdServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("DeleteMgmtdServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Mgmtd.Nodes))
	for i, node := range r.Cfg.Services.Mgmtd.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(rmContainerStep) },
		},
	})
}
