package fdb

import (
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/task"
)

// CreateFdbClusterTask is a task for creating a new FoundationDB cluster.
type CreateFdbClusterTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *CreateFdbClusterTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("CreateFdbClusterTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Fdb.Nodes))
	for i, node := range r.Cfg.Services.Fdb.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(genClusterFileContentStep) },
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep:  func() task.Step { return new(runContainerStep) },
		},
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(initClusterStep) },
		},
	})
}

// DeleteFdbClusterTask is a task for deleting a FoundationDB cluster.
type DeleteFdbClusterTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *DeleteFdbClusterTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("DeleteFdbClusterTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Fdb.Nodes))
	for i, node := range r.Cfg.Services.Fdb.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(rmContainerStep) },
		},
	})
}
