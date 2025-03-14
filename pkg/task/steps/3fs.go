package steps

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"path"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
)

func getNodeIDKey(service, name string) string {
	return fmt.Sprintf("%s-node-%s-id-", service, name)
}

func getConfigDir(workDir string) string {
	return path.Join(workDir, "config.d")
}

type gen3FSNodeIDStep struct {
	task.BaseStep

	idBegin int
	service string
	nodes   []string
}

func (s *gen3FSNodeIDStep) Execute(context.Context) error {
	nodes := make([]config.Node, len(s.nodes))
	for i, nodeName := range s.nodes {
		nodes[i] = s.Runtime.Nodes[nodeName]
	}

	nodeIDMap := make(map[string]int, len(nodes))
	for i, node := range nodes {
		s.Runtime.Store(getNodeIDKey(s.service, node.Name), s.idBegin+i)
		nodeIDMap[node.Name] = s.idBegin + i
	}
	s.Logger.Debugf("Node ID map: %v", nodeIDMap)

	return nil
}

// NewGen3FSNodeIDStepFunc is the generate 3fs node id step factory func.
func NewGen3FSNodeIDStepFunc(service string, idBegin int, nodes []string) func() task.Step {
	return func() task.Step {
		return &gen3FSNodeIDStep{
			service: service,
			nodes:   nodes,
			idBegin: idBegin,
		}
	}
}

type prepare3FSConfigStep struct {
	task.BaseStep

	service              string
	serviceWorkDir       string
	mainAppTomlTmpl      []byte
	mainLauncherTomlTmpl []byte
	mainTomlTmpl         []byte
	rdmaListenPort       int
	tcpListenPort        int
}

func (s *prepare3FSConfigStep) Execute(ctx context.Context) error {
	localEm := s.Runtime.LocalEm
	err := localEm.FS.MkdirAll(s.serviceWorkDir)
	if err != nil {
		return errors.Trace(err)
	}
	tmpDir, err := localEm.FS.MkdirTemp(s.serviceWorkDir, s.Node.Name+".*")
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := localEm.FS.RemoveAll(tmpDir); err != nil {
			s.Logger.Errorf("Failed to remove temporary directory %s: %v", tmpDir, err)
		}
	}()

	if err = s.genConfigs(tmpDir); err != nil {
		return errors.Trace(err)
	}

	if err = s.genFdbClusterFile(tmpDir); err != nil {
		return errors.Trace(err)
	}

	if err = s.copyFile(tmpDir); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepare3FSConfigStep) copyFile(src string) error {
	dst := getConfigDir(s.serviceWorkDir)
	s.Logger.Infof("Copying %s configs from %s to %s %s", s.service, src, s.Node.Name, dst)
	if err := s.Em.Runner.Scp(src, dst); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepare3FSConfigStep) genConfig(path, tmplName string, tmpl []byte, tmplData any) error {
	s.Logger.Infof("Generating %s to %s", tmplName, path)
	t, err := template.New(tmplName).Parse(string(tmpl))
	if err != nil {
		return errors.Annotatef(err, "parse template of %s", path)
	}
	data := new(bytes.Buffer)
	err = t.Execute(data, tmplData)
	if err != nil {
		return errors.Annotatef(err, "execute template of %s", path)
	}
	s.Logger.Debugf("Config of %s: %s", tmplName, data.String())

	err = s.Runtime.LocalEm.FS.WriteFile(path, data.Bytes(), 0644)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepare3FSConfigStep) genConfigs(tmpDir string) error {
	nodeIDI, _ := s.Runtime.Load(getNodeIDKey(s.service, s.Node.Name))
	nodeID := nodeIDI.(int)
	var mgmtdServerAddresses string

	if valI, ok := s.Runtime.Load(task.RuntimeMgmtdServerAddresseslKey); ok {
		mgmtdServerAddresses = valI.(string)
	}

	mainAppToml := path.Join(tmpDir, fmt.Sprintf("%s_app.toml", s.service))
	mainLauncherToml := path.Join(tmpDir, fmt.Sprintf("%s_launcher.toml", s.service))
	mainToml := path.Join(tmpDir, fmt.Sprintf("%s.toml", s.service))
	adminCliToml := path.Join(tmpDir, "admin_cli.toml")

	appTmplData := map[string]any{
		"NodeID": nodeID,
	}
	s.Logger.Debugf("Template data of %s_app.toml.tmpl: %v", s.service, appTmplData)
	if err := s.genConfig(mainAppToml, fmt.Sprintf("%s_app.toml", s.service),
		s.mainAppTomlTmpl, appTmplData); err != nil {

		return errors.Trace(err)
	}

	launcherTmplData := map[string]any{
		"ClusterID":            s.Runtime.Cfg.Name,
		"MgmtdServerAddresses": mgmtdServerAddresses,
	}
	s.Logger.Debugf("Template data of %s_launcher.toml.tmpl: %v", s.service, launcherTmplData)
	if err := s.genConfig(mainLauncherToml, fmt.Sprintf("%s_launcher.toml", s.service),
		s.mainLauncherTomlTmpl, launcherTmplData); err != nil {

		return errors.Trace(err)
	}

	// TODO: read monitor remote IP from runtime
	mainTmplData := map[string]any{
		"MonitorRemoteIP":      "",
		"RDMAListenPort":       s.rdmaListenPort,
		"TCPListenPort":        s.tcpListenPort,
		"MgmtdServerAddresses": mgmtdServerAddresses,
	}
	s.Logger.Debugf("Template data of %s.toml.tmpl: %v", s.service, mainTmplData)
	if err := s.genConfig(mainToml, fmt.Sprintf("%s.toml", s.service),
		s.mainTomlTmpl, mainTmplData); err != nil {

		return errors.Trace(err)
	}

	adminCliI, _ := s.Runtime.Load(task.RuntimeAdminCliTomlKey)
	adminCliTomlData := adminCliI.([]byte)
	err := s.Runtime.LocalEm.FS.WriteFile(adminCliToml, adminCliTomlData, os.FileMode(0644))
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepare3FSConfigStep) genFdbClusterFile(tmpDir string) error {
	contentI, _ := s.Runtime.Load(task.RuntimeFdbClusterFileContentKey)
	content := contentI.(string)

	clusterFilePath := path.Join(tmpDir, "fdb.cluster")
	s.Logger.Infof("Generating fdb.cluster to %s", clusterFilePath)
	if err := s.Runtime.LocalEm.FS.WriteFile(clusterFilePath, []byte(content), 0644); err != nil {
		return errors.Trace(err)
	}

	return nil
}

// Prepare3FSConfigStepSetup is a struct that holds the configuration of the prepare3FSConfigStep.
type Prepare3FSConfigStepSetup struct {
	Service              string
	ServiceWorkDir       string
	MainAppTomlTmpl      []byte
	MainLauncherTomlTmpl []byte
	MainTomlTmpl         []byte
	RDMAListenPort       int
	TCPListenPort        int
}

// NewPrepare3FSConfigStepFunc is prepare 3fs config step factory func.
func NewPrepare3FSConfigStepFunc(setup *Prepare3FSConfigStepSetup) func() task.Step {
	return func() task.Step {
		return &prepare3FSConfigStep{
			service:              setup.Service,
			serviceWorkDir:       setup.ServiceWorkDir,
			mainAppTomlTmpl:      setup.MainAppTomlTmpl,
			mainLauncherTomlTmpl: setup.MainLauncherTomlTmpl,
			mainTomlTmpl:         setup.MainTomlTmpl,
			rdmaListenPort:       setup.RDMAListenPort,
			tcpListenPort:        setup.TCPListenPort,
		}
	}
}

type run3FSContainerStep struct {
	task.BaseStep

	imgName        string
	containerName  string
	service        string
	serviceWorkDir string
}

func (s *run3FSContainerStep) Execute(ctx context.Context) error {
	s.Logger.Infof("Start %s container", s.service)
	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, s.imgName)
	if err != nil {
		return errors.Trace(err)
	}
	args := &external.RunArgs{
		Image:       img,
		Name:        &s.containerName,
		HostNetwork: true,
		Privileged:  common.Pointer(true),
		Ulimits: map[string]string{
			"nofile": "1048576:1048576",
		},
		Command: []string{
			fmt.Sprintf("/opt/3fs/bin/%s", s.service),
			"--launcher_cfg", fmt.Sprintf("/opt/3fs/etc/%s_launcher.toml", s.service),
			"--app_cfg", fmt.Sprintf("/opt/3fs/etc/%s_app.toml", s.service),
		},
		Detach: common.Pointer(true),
		Volumes: []*external.VolumeArgs{
			{
				Source: getConfigDir(s.serviceWorkDir),
				Target: "/opt/3fs/etc/",
			},
			{
				Source: "/dev",
				Target: "/dev",
			},
		},
	}
	_, err = s.Em.Docker.Run(ctx, args)
	if err != nil {
		return errors.Trace(err)
	}

	s.Logger.Infof("Start %s container success", s.service)
	return nil
}

// NewRun3FSContainerStepFunc is run3FSContainer factory func.
func NewRun3FSContainerStepFunc(imgName, containerName, service, workDir string) func() task.Step {
	return func() task.Step {
		return &run3FSContainerStep{
			imgName:        imgName,
			containerName:  containerName,
			service:        service,
			serviceWorkDir: workDir,
		}
	}
}

type rm3FSContainerStep struct {
	task.BaseStep

	containerName  string
	service        string
	serviceWorkDir string
}

func (s *rm3FSContainerStep) Execute(ctx context.Context) error {
	s.Logger.Infof("Remove %s container %s", s.service, s.containerName)
	_, err := s.Em.Docker.Rm(ctx, s.containerName, true)
	if err != nil {
		return errors.Trace(err)
	}

	configDir := getConfigDir(s.serviceWorkDir)
	s.Logger.Infof("Remove %s container config dir %s", s.serviceWorkDir, configDir)
	_, err = s.Em.Runner.Exec(ctx, "rm", "-rf", configDir)
	if err != nil {
		return errors.Annotatef(err, "rm %s", configDir)
	}

	s.Logger.Infof("Service %s container %s successfully removed", s.service, s.containerName)
	return nil
}

// NewRm3FSContainerStepFunc is rm3FSContainer factory func.
func NewRm3FSContainerStepFunc(containerName, service, serviceWorkDir string) func() task.Step {
	return func() task.Step {
		return &rm3FSContainerStep{
			containerName:  containerName,
			service:        service,
			serviceWorkDir: serviceWorkDir,
		}
	}
}
