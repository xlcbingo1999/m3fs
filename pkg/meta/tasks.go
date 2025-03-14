package meta

import (
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/task"
	"github.com/open3fs/m3fs/pkg/task/steps"
)

const (
	// ServiceName is the name of the meta service.
	ServiceName = "meta_main"
	serviceType = "META"
)

// CreateMetaServiceTask is a task for creating 3fs meta services.
type CreateMetaServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *CreateMetaServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("CreateMetaServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Meta.Nodes))
	for i, node := range r.Cfg.Services.Meta.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: steps.NewGen3FSNodeIDStepFunc(ServiceName, 100, r.Cfg.Services.Meta.Nodes),
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewPrepare3FSConfigStepFunc(&steps.Prepare3FSConfigStepSetup{
				Service:              ServiceName,
				ServiceWorkDir:       r.Services.Meta.WorkDir,
				MainAppTomlTmpl:      MetaMainAppTomlTmpl,
				MainLauncherTomlTmpl: MetaMainLauncherTomlTmpl,
				MainTomlTmpl:         MetaMainTomlTmpl,
				RDMAListenPort:       r.Services.Meta.RDMAListenPort,
				TCPListenPort:        r.Services.Meta.TCPListenPort,
			}),
		},
		{
			Nodes: []config.Node{nodes[0]},
			NewStep: steps.NewUpload3FSMainConfigStepFunc(
				"3fs",
				r.Services.Meta.ContainerName,
				ServiceName,
				r.Services.Meta.WorkDir,
				serviceType,
			),
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRun3FSContainerStepFunc("3fs",
				r.Services.Meta.ContainerName,
				ServiceName,
				r.Services.Meta.WorkDir),
		},
	})
}

// DeleteMetaServiceTask is a task for deleting a meta services.
type DeleteMetaServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *DeleteMetaServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("DeleteMetaServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Meta.Nodes))
	for i, node := range r.Cfg.Services.Meta.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes: nodes,
			NewStep: steps.NewRm3FSContainerStepFunc(
				r.Services.Meta.ContainerName,
				ServiceName,
				r.Services.Meta.WorkDir),
		},
	})
}
