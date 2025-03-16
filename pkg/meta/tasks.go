package meta

import (
	"path"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/task"
	"github.com/open3fs/m3fs/pkg/task/steps"
)

const (
	// ServiceName is the name of the meta service.
	ServiceName = "meta_main"
	serviceType = "META"
)

func getServiceWorkDir(workDir string) string {
	return path.Join(workDir, "meta")
}

// CreateMetaServiceTask is a task for creating 3fs meta services.
type CreateMetaServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *CreateMetaServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("CreateMetaServiceTask")

	workDir := getServiceWorkDir(r.WorkDir)
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
				ServiceWorkDir:       workDir,
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
				workDir,
				serviceType,
			),
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRun3FSContainerStepFunc(
				&steps.Run3FSContainerStepSetup{
					ImgName:       "3fs",
					ContainerName: r.Services.Meta.ContainerName,
					Service:       ServiceName,
					WorkDir:       workDir,
				},
			),
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
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRm3FSContainerStepFunc(
				r.Services.Meta.ContainerName,
				ServiceName,
				getServiceWorkDir(r.WorkDir)),
		},
	})
}
