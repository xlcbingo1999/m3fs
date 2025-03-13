package mgmtd

import (
	"bytes"
	"context"
	"embed"
	"path"
	"text/template"

	"github.com/open3fs/m3fs/pkg/common"
	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
)

var (
	//go:embed templates/*.tmpl
	templatesFs embed.FS

	// MgmtdMainAppTomlTmpl is the template content of mgmtd_main_app.toml
	MgmtdMainAppTomlTmpl []byte
	// MgmtdMainLauncherTomlTmpl is the template content of mgmtd_main_launcher.toml
	MgmtdMainLauncherTomlTmpl []byte
	// MgmtdMainTomlTmpl is the template content of mgmtd_main.toml
	MgmtdMainTomlTmpl []byte
	// AdminCliTomlTmpl is the template content of admin_cli.toml
	AdminCliTomlTmpl []byte
)

func init() {
	var err error
	MgmtdMainAppTomlTmpl, err = templatesFs.ReadFile("templates/mgmtd_main_app.toml.tmpl")
	if err != nil {
		panic(err)
	}

	MgmtdMainLauncherTomlTmpl, err = templatesFs.ReadFile("templates/mgmtd_main_launcher.toml.tmpl")
	if err != nil {
		panic(err)
	}

	MgmtdMainTomlTmpl, err = templatesFs.ReadFile("templates/mgmtd_main.toml.tmpl")
	if err != nil {
		panic(err)
	}

	AdminCliTomlTmpl, err = templatesFs.ReadFile("templates/admin_cli.toml.tmpl")
	if err != nil {
		panic(err)
	}
}

func getNodeIDKey(name string) string {
	return "mgmtd-node-id-" + name
}

func getConfigDir(workDir string) string {
	return path.Join(workDir, "config.d")
}

type genNodeIDStep struct {
	task.BaseStep
}

func (s *genNodeIDStep) Execute(context.Context) error {
	mgmtd := s.Runtime.Services.Mgmtd
	nodes := make([]config.Node, len(mgmtd.Nodes))
	for i, mgmtdNode := range mgmtd.Nodes {
		nodes[i] = s.Runtime.Nodes[mgmtdNode]
	}

	nodeIDMap := make(map[string]int, len(nodes))
	for i, node := range nodes {
		s.Runtime.Store(getNodeIDKey(node.Name), i+1)
		nodeIDMap[node.Name] = i + 1
	}
	s.Logger.Debugf("Node ID map: %v", nodeIDMap)

	return nil
}

type prepareMgmtdConfigStep struct {
	task.BaseStep
}

func (s *prepareMgmtdConfigStep) Execute(ctx context.Context) error {
	localEm := s.Runtime.LocalEm
	err := localEm.FS.MkdirAll(s.Runtime.Services.Mgmtd.WorkDir)
	if err != nil {
		return errors.Trace(err)
	}
	tmpDir, err := localEm.FS.MkdirTemp(s.Runtime.Services.Mgmtd.WorkDir, s.Node.Name+".*")
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

	if err = s.copyFile(tmpDir, s.Runtime.Services.Mgmtd.WorkDir); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepareMgmtdConfigStep) copyFile(src, workDir string) error {
	dst := getConfigDir(workDir)
	s.Logger.Infof("Copying mgmtd configs from %s to %s %s", src, s.Node.Name, dst)
	if err := s.Em.Runner.Scp(src, dst); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepareMgmtdConfigStep) genConfig(path, tmplName string, tmpl []byte, tmplData any) error {
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

func (s *prepareMgmtdConfigStep) genConfigs(tmpDir string) error {
	mgmtd := s.Runtime.Services.Mgmtd
	nodeIDI, _ := s.Runtime.Load(getNodeIDKey(s.Node.Name))
	nodeID := nodeIDI.(int)

	mgmtdMainAppToml := path.Join(tmpDir, "mgmtd_main_app.toml")
	mgmtdMainLauncherToml := path.Join(tmpDir, "mgmtd_main_launcher.toml")
	mgmtdMainToml := path.Join(tmpDir, "mgmtd_main.toml")
	adminCliToml := path.Join(tmpDir, "admin_cli.toml")

	appTmplData := map[string]any{
		"NodeID": nodeID,
	}
	s.Logger.Debugf("Mgmtd main app config template data: %v", appTmplData)
	if err := s.genConfig(mgmtdMainAppToml, "mgmtd_main_app.toml",
		MgmtdMainAppTomlTmpl, appTmplData); err != nil {

		return errors.Trace(err)
	}

	launcherTmplData := map[string]any{
		"ClusterID": s.Runtime.Cfg.Name,
	}
	s.Logger.Debugf("Mgmtd main launcher config template data: %v", launcherTmplData)
	if err := s.genConfig(mgmtdMainLauncherToml, "mgmtd_main_launcher.toml",
		MgmtdMainLauncherTomlTmpl, launcherTmplData); err != nil {

		return errors.Trace(err)
	}

	s.Logger.Debugf("Admin cli config template data: %v", launcherTmplData)
	if err := s.genConfig(adminCliToml, "admin_cli.toml",
		AdminCliTomlTmpl, launcherTmplData); err != nil {

		return errors.Trace(err)
	}

	// TODO: read monitor remote IP from runtime
	mainTmplData := map[string]any{
		"MonitorRemoteIP": "",
		"RDMAListenPort":  mgmtd.RDMAListenPort,
		"TCPListenPort":   mgmtd.TCPListenPort,
	}
	s.Logger.Debugf("Mgmtd main config template data: %v", mainTmplData)
	if err := s.genConfig(mgmtdMainToml, "mgmtd_main.toml",
		MgmtdMainTomlTmpl, mainTmplData); err != nil {

		return errors.Trace(err)
	}

	return nil
}

func (s *prepareMgmtdConfigStep) genFdbClusterFile(tmpDir string) error {
	contentI, _ := s.Runtime.Load(task.RuntimeFdbClusterFileContentKey)
	content := contentI.(string)

	clusterFilePath := path.Join(tmpDir, "fdb.cluster")
	s.Logger.Infof("Generating fdb.cluster to %s", clusterFilePath)
	if err := s.Runtime.LocalEm.FS.WriteFile(clusterFilePath, []byte(content), 0644); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func getContainerVolumes(workDir string) []*external.VolumeArgs {
	configDir := getConfigDir(workDir)
	return []*external.VolumeArgs{
		{
			Source: path.Join(configDir, "mgmtd_main_app.toml"),
			Target: "/opt/3fs/etc/mgmtd_main_app.toml",
		},
		{
			Source: path.Join(configDir, "mgmtd_main_launcher.toml"),
			Target: "/opt/3fs/etc/mgmtd_main_launcher.toml",
		},
		{
			Source: path.Join(configDir, "mgmtd_main.toml"),
			Target: "/opt/3fs/etc/mgmtd_main.toml",
		},
		{
			Source: path.Join(configDir, "admin_cli.toml"),
			Target: "/opt/3fs/etc/admin_cli.toml",
		},
		{
			Source: path.Join(configDir, "fdb.cluster"),
			Target: "/opt/3fs/etc/fdb.cluster",
		},
	}
}

type initClusterStep struct {
	task.BaseStep
}

func (s *initClusterStep) Execute(ctx context.Context) error {
	mgmtd := s.Runtime.Services.Mgmtd
	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, "3fs")
	if err != nil {
		return errors.Trace(err)
	}
	args := &external.RunArgs{
		Image:      img,
		Name:       &mgmtd.ContainerName,
		Entrypoint: common.Pointer("''"),
		Rm:         common.Pointer(true),
		Command: []string{
			"/opt/3fs/bin/admin_cli",
			"-cfg", "/opt/3fs/etc/admin_cli.toml",
			"'init-cluster --mgmtd /opt/3fs/etc/mgmtd_main.toml 1 1048576 16'"},
		HostNetwork: true,
		Volumes:     getContainerVolumes(mgmtd.WorkDir),
	}
	_, err = s.Em.Docker.Run(ctx, args)
	if err != nil {
		return errors.Trace(err)
	}

	s.Logger.Infof("Cluster initialization success")
	return nil
}

type runContainerStep struct {
	task.BaseStep
}

func (s *runContainerStep) Execute(ctx context.Context) error {
	mgmtd := s.Runtime.Services.Mgmtd
	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, "3fs")
	if err != nil {
		return errors.Trace(err)
	}
	args := &external.RunArgs{
		Image:       img,
		Name:        &mgmtd.ContainerName,
		HostNetwork: true,
		Privileged:  common.Pointer(true),
		Ulimits: map[string]string{
			"nofile": "1048576:1048576",
		},
		Command: []string{
			"/opt/3fs/bin/mgmtd_main",
			"--launcher_cfg", "/opt/3fs/etc/mgmtd_main_launcher.toml",
			"--app_cfg", "/opt/3fs/etc/mgmtd_main_app.toml",
		},
		Detach:  common.Pointer(true),
		Volumes: getContainerVolumes(mgmtd.WorkDir),
	}
	_, err = s.Em.Docker.Run(ctx, args)
	if err != nil {
		return errors.Trace(err)
	}

	s.Logger.Infof("Start mgmtd container success")
	return nil
}

type rmContainerStep struct {
	task.BaseStep
}

func (s *rmContainerStep) Execute(ctx context.Context) error {
	containerName := s.Runtime.Services.Mgmtd.ContainerName
	s.Logger.Infof("Remove mgmtd container %s", containerName)
	_, err := s.Em.Docker.Rm(ctx, containerName, true)
	if err != nil {
		return errors.Trace(err)
	}

	configDir := getConfigDir(s.Runtime.Services.Mgmtd.WorkDir)
	s.Logger.Infof("Remove mgmtd container config dir %s", configDir)
	_, err = s.Em.Runner.Exec(ctx, "rm", "-rf", configDir)
	if err != nil {
		return errors.Annotatef(err, "rm %s", configDir)
	}

	s.Logger.Infof("Mgmtd container %s successfully removed", containerName)
	return nil
}
