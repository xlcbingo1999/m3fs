package storage

import (
	"embed"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/task"
	"github.com/open3fs/m3fs/pkg/task/steps"
)

var (
	//go:embed templates/*.tmpl
	templatesFs embed.FS

	// StorageMainAppTomlTmpl is the template content of storage_main_app.toml
	StorageMainAppTomlTmpl []byte
	// StorageMainLauncherTomlTmpl is the template content of storage_main_launcher.toml
	StorageMainLauncherTomlTmpl []byte
	// StorageMainTomlTmpl is the template content of storage_main.toml
	StorageMainTomlTmpl []byte
	// DiskToolScript is the template content of disk_tool.sh
	DiskToolScript []byte
)

func init() {
	var err error
	StorageMainAppTomlTmpl, err = templatesFs.ReadFile("templates/storage_main_app.toml.tmpl")
	if err != nil {
		panic(err)
	}

	StorageMainLauncherTomlTmpl, err = templatesFs.ReadFile("templates/storage_main_launcher.toml.tmpl")
	if err != nil {
		panic(err)
	}

	StorageMainTomlTmpl, err = templatesFs.ReadFile("templates/storage_main.toml.tmpl")
	if err != nil {
		panic(err)
	}

	DiskToolScript, err = templatesFs.ReadFile("templates/disk_tool.sh.tmpl")
	if err != nil {
		panic(err)
	}
}

func makeTargetPaths(diskNum int) string {
	targets := make([]string, diskNum)
	for i := 0; i < diskNum; i++ {
		targets[i] = fmt.Sprintf(`"%s"`,
			path.Join("/mnt", "3fsdata", "data"+strconv.Itoa(i), "3fs"))
	}

	return fmt.Sprintf("[%s]", strings.Join(targets, ","))
}

const (
	// ServiceName is the name of the storage service.
	ServiceName = "storage_main"
	serviceType = "STORAGE"
)

// CreateStorageServiceTask is a task for creating 3fs storage services.
type CreateStorageServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *CreateStorageServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("CreateStorageServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Storage.Nodes))
	storage := r.Cfg.Services.Storage
	for i, node := range storage.Nodes {
		nodes[i] = r.Nodes[node]
	}
	t.SetSteps([]task.StepConfig{
		{
			Nodes:   []config.Node{nodes[0]},
			NewStep: steps.NewGen3FSNodeIDStepFunc(ServiceName, 10001, storage.Nodes),
		},
		{
			Nodes: nodes,
			NewStep: steps.NewRemoteRunScriptStepFunc(
				storage.WorkDir,
				"disk_tool.sh",
				DiskToolScript,
				[]string{
					storage.WorkDir,
					strconv.Itoa(storage.DiskNumPerNode),
					string(storage.DiskType),
					"prepare",
				}),
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewPrepare3FSConfigStepFunc(&steps.Prepare3FSConfigStepSetup{
				Service:              ServiceName,
				ServiceWorkDir:       storage.WorkDir,
				MainAppTomlTmpl:      StorageMainAppTomlTmpl,
				MainLauncherTomlTmpl: StorageMainLauncherTomlTmpl,
				MainTomlTmpl:         StorageMainTomlTmpl,
				RDMAListenPort:       storage.RDMAListenPort,
				TCPListenPort:        storage.TCPListenPort,
				ExtraMainTomlData: map[string]any{
					"TargetPaths": makeTargetPaths(storage.DiskNumPerNode),
				},
			}),
		},
		{
			Nodes: []config.Node{nodes[0]},
			NewStep: steps.NewUpload3FSMainConfigStepFunc(
				"3fs",
				storage.ContainerName,
				ServiceName,
				storage.WorkDir,
				serviceType,
			),
		},
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRun3FSContainerStepFunc(
				&steps.Run3FSContainerStepSetup{
					ImgName:       "3fs",
					ContainerName: storage.ContainerName,
					Service:       ServiceName,
					WorkDir:       storage.WorkDir,
					ExtraVolumes: []*external.VolumeArgs{
						{
							Source: path.Join(storage.WorkDir, "3fsdata"),
							Target: "/mnt/3fsdata",
						},
					},
				}),
		},
	})
}

// DeleteStorageServiceTask is a task for deleting a storage services.
type DeleteStorageServiceTask struct {
	task.BaseTask
}

// Init initializes the task.
func (t *DeleteStorageServiceTask) Init(r *task.Runtime) {
	t.BaseTask.Init(r)
	t.BaseTask.SetName("DeleteStorageServiceTask")
	nodes := make([]config.Node, len(r.Cfg.Services.Storage.Nodes))
	for i, node := range r.Cfg.Services.Storage.Nodes {
		nodes[i] = r.Nodes[node]
	}
	storage := r.Services.Storage
	t.SetSteps([]task.StepConfig{
		{
			Nodes:    nodes,
			Parallel: true,
			NewStep: steps.NewRm3FSContainerStepFunc(
				r.Services.Storage.ContainerName,
				ServiceName,
				r.Services.Storage.WorkDir),
		},
		{
			Nodes: nodes,
			NewStep: steps.NewRemoteRunScriptStepFunc(
				storage.WorkDir,
				"disk_tool.sh",
				DiskToolScript,
				[]string{
					storage.WorkDir,
					strconv.Itoa(storage.DiskNumPerNode),
					string(storage.DiskType),
					"clear",
				}),
		},
	})
}
