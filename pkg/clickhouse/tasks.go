package clickhouse

import (
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/task"
)

// CreateClickhouseClusterTask is a task for creating a new clickhouse cluster.
type CreateClickhouseClusterTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *CreateClickhouseClusterTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("CreateClickhouseClusterTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Clickhouse.Nodes))
	for i, node := range r.Cfg.Services.Clickhouse.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   nodes,
			NewStep: func() task.Step { return new(startContainerStep) },
		},
	})
}
