package mgmtd

import (
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/task"
	"github.com/open3fs/m3fs/pkg/task/steps"
)

// ServiceName is the name of the mgmtd service.
const ServiceName = "mgmtd_main"

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
			NewStep: steps.NewGen3FSNodeIDStepFunc(ServiceName, 1, r.Cfg.Services.Mgmtd.Nodes),
		},
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(genAdminCliConfigStep) },
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewPrepare3FSConfigStepFunc(&steps.Prepare3FSConfigStepSetup{
				Service:              ServiceName,
				ServiceWorkDir:       r.Services.Mgmtd.WorkDir,
				MainAppTomlTmpl:      MgmtdMainAppTomlTmpl,
				MainLauncherTomlTmpl: MgmtdMainLauncherTomlTmpl,
				MainTomlTmpl:         MgmtdMainTomlTmpl,
				RDMAListenPort:       r.Services.Mgmtd.RDMAListenPort,
				TCPListenPort:        r.Services.Mgmtd.TCPListenPort,
			}),
		},
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(initClusterStep) },
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRun3FSContainerStepFunc(
				&steps.Run3FSContainerStepSetup{
					ImgName:       "3fs",
					ContainerName: r.Services.Mgmtd.ContainerName,
					Service:       ServiceName,
					WorkDir:       r.Services.Mgmtd.WorkDir,
				}),
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
			Nodes: nodes,
			NewStep: steps.NewRm3FSContainerStepFunc(
				r.Services.Mgmtd.ContainerName,
				ServiceName,
				r.Services.Mgmtd.WorkDir),
		},
	})
}

// InitUserAndChainTask is a task for initializing user and chain.
type InitUserAndChainTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *InitUserAndChainTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("InitUserAndChainTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Mgmtd.Nodes))
	for i, node := range r.Cfg.Services.Mgmtd.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: func() task.Step { return new(initUserAndChainStep) },
		},
	})
}
