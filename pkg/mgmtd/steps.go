package mgmtd

import (
	"bytes"
	"context"
	"embed"
	"path"
	"text/template"

	"github.com/open3fs/m3fs/pkg/config"
	"github.com/open3fs/m3fs/pkg/errors"
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
}

func getNodeID(name string) string {
	return "mgmtd-node-id-" + name
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
		s.Runtime.Store(getNodeID(node.Name), i+1)
		nodeIDMap[node.Name] = i + 1
	}
	s.Logger.Debugf("Node ID map: %v", nodeIDMap)

	return nil
}

type prepareMgmtdConfigStep struct {
	task.BaseStep
}

func (s *prepareMgmtdConfigStep) Execute(context.Context) error {
	err := s.Em.Local.MkdirAll(s.Runtime.Services.Mgmtd.WorkDir)
	if err != nil {
		return errors.Trace(err)
	}
	tmpDir, err := s.Em.Local.MkdirTemp(s.Runtime.Services.Mgmtd.WorkDir, s.Node.Name+".*")
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := s.Em.Local.RemoveAll(tmpDir); err != nil {
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

func (s *prepareMgmtdConfigStep) copyFile(src, dst string) error {
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

	err = s.Em.Local.WriteFile(path, data.Bytes(), 0644)
	if err != nil {
		return errors.Trace(err)
	}

	return nil
}

func (s *prepareMgmtdConfigStep) genConfigs(tmpDir string) error {
	mgmtd := s.Runtime.Services.Mgmtd
	nodeIDI, _ := s.Runtime.Load(getNodeID(s.Node.Name))
	nodeID := nodeIDI.(int)

	mgmtdMainAppToml := path.Join(tmpDir, "mgmtd_main_app.toml")
	mgmtdMainLauncherToml := path.Join(tmpDir, "mgmtd_main_launcher.toml")
	mgmtdMainToml := path.Join(tmpDir, "mgmtd_main.toml")

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
	if err := s.Em.Local.WriteFile(clusterFilePath, []byte(content), 0644); err != nil {
		return errors.Trace(err)
	}

	return nil
}
