package clickhouse

import (
	"context"
	"embed"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/external"
	"github.com/open3fs/m3fs/pkg/image"
	"github.com/open3fs/m3fs/pkg/task"
)

var (
	//go:embed templates/*
	templatesFs embed.FS

	// ClickhouseConfigTmpl is the template content of config.xml
	ClickhouseConfigTmpl []byte

	// ClickhouseSqlTmpl is the template content of 3fs-monitor.sql
	ClickhouseSqlTmpl []byte
)

func init() {
	var err error
	ClickhouseConfigTmpl, err = templatesFs.ReadFile("templates/config.tmpl")
	if err != nil {
		panic(err)
	}
	ClickhouseSqlTmpl, err = templatesFs.ReadFile("templates/sql.tmpl")
	if err != nil {
		panic(err)
	}
}

type genClickhouseConfigStep struct {
	task.BaseStep
}

func (s *genClickhouseConfigStep) Execute(context.Context) error {
	tempDir, err := os.MkdirTemp(os.TempDir(), "3fs-clickhouse.")
	if err != nil {
		return errors.Trace(err)
	}
	s.Runtime.Store("clickhouse_temp_config_dir", tempDir)

	configFileName := "config.xml"
	configFile, err := os.Create(filepath.Join(tempDir, configFileName))
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := configFile.Close(); err != nil {
			s.Logger.Warnf("Failed to close file %+v", err)
		}
	}()
	configTmpl, err := template.New(configFileName).Parse(string(ClickhouseConfigTmpl))
	if err != nil {
		return errors.Annotate(err, "parse config.xml template")
	}
	err = configTmpl.Execute(configFile, map[string]string{
		"Port": strconv.Itoa(s.Runtime.Services.Clickhouse.TcpPort),
	})
	if err != nil {
		return errors.Annotate(err, "write config.xml")
	}

	sqlFileName := "3fs-monitor.sql"
	sqlFile, err := os.Create(filepath.Join(tempDir, sqlFileName))
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := sqlFile.Close(); err != nil {
			s.Logger.Warnf("Failed to close file %+v", err)
		}
	}()
	sqlTmpl, err := template.New(sqlFileName).Parse(string(ClickhouseSqlTmpl))
	if err != nil {
		return errors.Annotate(err, "parse 3fs-monitor.sql template")
	}
	err = sqlTmpl.Execute(sqlFile, map[string]string{
		"Db": s.Runtime.Services.Clickhouse.Db,
	})
	if err != nil {
		return errors.Annotate(err, "write 3fs-monitor.sql")
	}

	return nil
}

type startContainerStep struct {
	task.BaseStep
}

func (s *startContainerStep) Execute(ctx context.Context) error {
	workDir := s.Runtime.Services.Clickhouse.WorkDir
	dataDir := path.Join(workDir, "data")
	if _, err := s.Em.OS.Exec(ctx, "mkdir", "-p", dataDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", dataDir)
	}
	logDir := path.Join(workDir, "log")
	if _, err := s.Em.OS.Exec(ctx, "mkdir", "-p", logDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", logDir)
	}
	configDir := path.Join(workDir, "config.d")
	if _, err := s.Em.OS.Exec(ctx, "mkdir", "-p", configDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", configDir)
	}
	localConfigDir, _ := s.Runtime.Load("clickhouse_temp_config_dir")
	localConfigFile := path.Join(localConfigDir.(string), "config.xml")
	remoteConfigFile := path.Join(configDir, "config.xml")
	if err := s.Em.Runner.Scp(localConfigFile, remoteConfigFile); err != nil {
		return errors.Annotatef(err, "scp config.xml")
	}

	sqlDir := path.Join(workDir, "sql")
	if _, err := s.Em.OS.Exec(ctx, "mkdir", "-p", sqlDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", sqlDir)
	}
	localSqlFile := path.Join(localConfigDir.(string), "3fs-monitor.sql")
	remoteSqlFile := path.Join(sqlDir, "3fs-monitor.sql")
	if err := s.Em.Runner.Scp(localSqlFile, remoteSqlFile); err != nil {
		return errors.Annotatef(err, "scp 3fs-monitor.sql")
	}

	img, err := image.GetImage(s.Runtime.Cfg.Registry.CustomRegistry, "clickhouse")
	if err != nil {
		return errors.Trace(err)
	}
	args := &external.RunArgs{
		Image:       img,
		Name:        &s.Runtime.Services.Clickhouse.ContainerName,
		HostNetwork: true,
		Envs: map[string]string{
			"CLICKHOUSE_DB":       s.Runtime.Services.Clickhouse.Db,
			"CLICKHOUSE_USER":     s.Runtime.Services.Clickhouse.User,
			"CLICKHOUSE_PASSWORD": s.Runtime.Services.Clickhouse.Password,
		},
		Volumes: []*external.VolumeArgs{
			{
				Source: dataDir,
				Target: "/var/lib/clickhouse",
			},
			{
				Source: logDir,
				Target: "/var/log/clickhouse-server",
			},
			{
				Source: configDir,
				Target: "/etc/clickhouse-server/config.d",
			},
			{
				Source: sqlDir,
				Target: "/tmp/sql",
			},
		},
	}
	_, err = s.Em.Docker.Run(ctx, args)
	if err != nil {
		return errors.Trace(err)
	}

	s.Logger.Infof("Started container %s", s.Runtime.Services.Clickhouse.ContainerName)
	return nil
}
