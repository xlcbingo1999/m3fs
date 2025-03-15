package fsclient

import (
	"embed"
	"path"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/task"
	"github.com/open3fs/m3fs/pkg/task/steps"
)

var (
	//go:embed templates/*.tmpl
	templatesFs embed.FS

	// ClientFuseMainLauncherTomlTmpl is the template content of hf3fs_fuse_main_launcher.toml
	ClientFuseMainLauncherTomlTmpl []byte
	// ClientMainTomlTmpl is the template content of hf3fs_fuse_main.toml
	ClientMainTomlTmpl []byte
)

func init() {
	var err error
	ClientFuseMainLauncherTomlTmpl, err = templatesFs.ReadFile("templates/hf3fs_fuse_main_launcher.toml.tmpl")
	if err != nil {
		panic(err)
	}

	ClientMainTomlTmpl, err = templatesFs.ReadFile("templates/hf3fs_fuse_main.toml.tmpl")
	if err != nil {
		panic(err)
	}
}

const (
	// ServiceName is the name of the 3fs client service.
	ServiceName = "hf3fs_fuse_main"
	serviceType = "FUSE"
)

func getServiceWorkDir(workDir string) string {
	return path.Join(workDir, "3fs_client")
}

// Create3FSClientServiceTask is a task for creating 3fs client services.
type Create3FSClientServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *Create3FSClientServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("Create3FSClicnetServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Client.Nodes))
	client := r.Cfg.Services.Client
	for i, node := range client.Nodes {
		nodes[i] = r.Nodes[node]
	}
	runContainerVolumes := []*external.VolumeArgs{}
	if client.HostMountpoint != "" {
		runContainerVolumes = append(runContainerVolumes, &external.VolumeArgs{
			Source: client.HostMountpoint,
			Target: "/mnt/3fs",
			Rshare: common.Pointer(true),
		})
	}
	token, _ := r.LoadString(task.RuntimeUserTokenKey)
	workDir := getServiceWorkDir(r.WorkDir)
	t.SetSteps([]task.StepConfig{
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewPrepare3FSConfigStepFunc(&steps.Prepare3FSConfigStepSetup{
				Service:              ServiceName,
				ServiceWorkDir:       workDir,
				MainAppTomlTmpl:      []byte(""),
				MainLauncherTomlTmpl: ClientFuseMainLauncherTomlTmpl,
				MainTomlTmpl:         ClientMainTomlTmpl,
				Extra3FSConfigFiles: []*steps.Extra3FSConfigFile{
					{
						FileName: "token.txt",
						Data:     []byte(token),
					},
				},
			}),
		},
		{
			Nodes: []config.Node{nodes[0]},
			NewStep: steps.NewUpload3FSMainConfigStepFunc(
				"3fs",
				client.ContainerName,
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
					ContainerName: client.ContainerName,
					Service:       ServiceName,
					WorkDir:       workDir,
					ExtraVolumes:  runContainerVolumes,
				}),
		},
	})
}

// Delete3FSClientServiceTask is a task for deleting a 3fs client services.
type Delete3FSClientServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *Delete3FSClientServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("Delete3FSClientServiceTask")
	client := r.Services.Client
	nodes := make([]config.Node, len(client.Nodes))
	for i, node := range client.Nodes {
		nodes[i] = r.Nodes[node]
	}
	workDir := getServiceWorkDir(r.WorkDir)
	t.SetSteps([]task.StepConfig{
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRm3FSContainerStepFunc(
				client.ContainerName,
				ServiceName,
				workDir),
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep:  func() task.Step { return new(umountHostMountponitStep) },
		},
	})
}
