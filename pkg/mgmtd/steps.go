package mgmtd

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/open3fs/m3fs/pkg/common"
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

type genAdminCliConfigStep struct {
	task.BaseStep
}

func (s *genAdminCliConfigStep) Execute(ctx context.Context) error {
	adminCliData := map[string]any{
		"ClusterID": s.Runtime.Cfg.Name,
	}
	s.Logger.Debugf("Admin cli config template data: %v", adminCliData)
	t, err := template.New("admin_cli.toml").Parse(string(AdminCliTomlTmpl))
	if err != nil {
		return errors.Annotatef(err, "parse template of admin_cli.toml.tmpl")
	}
	data := new(bytes.Buffer)
	err = t.Execute(data, adminCliData)
	if err != nil {
		return errors.Annotate(err, "execute template of admin_cli.toml.tmpl")
	}
	s.Logger.Debugf("Config of admin cli: %s", data.String())
	s.Runtime.Store(task.RuntimeAdminCliTomlKey, data.Bytes())

	return nil
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
		Volumes: []*external.VolumeArgs{
			{
				Source: path.Join(mgmtd.WorkDir, "config.d"),
				Target: "/opt/3fs/etc",
			},
		},
	}
	_, err = s.Em.Docker.Run(ctx, args)
	if err != nil {
		return errors.Trace(err)
	}

	address := make([]string, len(s.Runtime.Services.Mgmtd.Nodes))
	port := strconv.Itoa(s.Runtime.Services.Mgmtd.RDMAListenPort)
	for i, nodeName := range s.Runtime.Services.Mgmtd.Nodes {
		node := s.Runtime.Nodes[nodeName]
		address[i] = fmt.Sprintf(`"RDMA://%s"`, net.JoinHostPort(node.Host, port))
	}
	s.Runtime.Store(task.RuntimeMgmtdServerAddressesKey,
		fmt.Sprintf(`[%s]`, strings.Join(address, ",")))

	s.Logger.Infof("Cluster initialization success")
	return nil
}
