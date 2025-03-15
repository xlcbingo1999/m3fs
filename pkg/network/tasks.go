package network

import (
	"github.com/open3fs/m3fs/pkg/task"
)

// PrepareNetworkTask is a task for preparing network for a new node.
type PrepareNetworkTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *PrepareNetworkTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("PrepareNetworkTask")
	nodes := r.Cfg.Nodes

	t.SetSteps([]task.StepConfig{
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(genIbdev2netdevScriptStep) },
		},
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(loadRdmaRxeModuleStep) },
		},
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(createRdmaRxeLinkStep) },
		},
	})
}
