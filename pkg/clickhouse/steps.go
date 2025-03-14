package clickhouse

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/open3fs/m3fs/pkg/common"
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

	// ClickhouseSQLTmpl is the template content of 3fs-monitor.sql
	ClickhouseSQLTmpl []byte
)

func init() {
	var err error
	ClickhouseConfigTmpl, err = templatesFs.ReadFile("templates/config.tmpl")
	if err != nil {
		panic(err)
	}
	ClickhouseSQLTmpl, err = templatesFs.ReadFile("templates/sql.tmpl")
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
		"TCPPort": strconv.Itoa(s.Runtime.Services.Clickhouse.TCPPort),
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
	sqlTmpl, err := template.New(sqlFileName).Parse(string(ClickhouseSQLTmpl))
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
	if _, err := s.Em.Runner.Exec(ctx, "mkdir", "-p", dataDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", dataDir)
	}
	logDir := path.Join(workDir, "log")
	if _, err := s.Em.Runner.Exec(ctx, "mkdir", "-p", logDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", logDir)
	}
	configDir := path.Join(workDir, "config.d")
	if _, err := s.Em.Runner.Exec(ctx, "mkdir", "-p", configDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", configDir)
	}
	localConfigDir, _ := s.Runtime.Load("clickhouse_temp_config_dir")
	localConfigFile := path.Join(localConfigDir.(string), "config.xml")
	remoteConfigFile := path.Join(configDir, "config.xml")
	if err := s.Em.Runner.Scp(localConfigFile, remoteConfigFile); err != nil {
		return errors.Annotatef(err, "scp config.xml")
	}

	sqlDir := path.Join(workDir, "sql")
	if _, err := s.Em.Runner.Exec(ctx, "mkdir", "-p", sqlDir); err != nil {
		return errors.Annotatef(err, "mkdir %s", sqlDir)
	}
	localSQLFile := path.Join(localConfigDir.(string), "3fs-monitor.sql")
	remoteSQLFile := path.Join(sqlDir, "3fs-monitor.sql")
	if err := s.Em.Runner.Scp(localSQLFile, remoteSQLFile); err != nil {
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
		Detach:      common.Pointer(true),
		Envs: map[string]string{
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
	time.Sleep(time.Second * 5)

	s.Logger.Infof("Started container %s", s.Runtime.Services.Clickhouse.ContainerName)
	return nil
}

type initClusterStep struct {
	task.BaseStep
}

func (s *initClusterStep) Execute(ctx context.Context) error {
	err := s.initCluster(ctx)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (s *initClusterStep) initCluster(ctx context.Context) error {
	s.Logger.Infof("Initializing clickhouse cluster")
	_, err := s.Em.Docker.Exec(ctx, s.Runtime.Services.Clickhouse.ContainerName,
		"bash", "-c", fmt.Sprintf(`"clickhouse-client --port %d -n < /tmp/sql/3fs-monitor.sql"`,
			s.Runtime.Services.Clickhouse.TCPPort))
	if err != nil {
		return errors.Annotate(err, "initialize fdb cluster")
	}
	s.Logger.Infof("Initialized clickhouse cluster")
	return nil
}

type rmContainerStep struct {
	task.BaseStep
}

func (s *rmContainerStep) Execute(ctx context.Context) error {
	containerName := s.Runtime.Services.Clickhouse.ContainerName
	s.Logger.Infof("Removing clickhouse container %s", containerName)
	_, err := s.Em.Docker.Rm(ctx, containerName, true)
	if err != nil {
		return errors.Trace(err)
	}

	dataDir := path.Join(s.Runtime.Services.Clickhouse.WorkDir, "data")
	s.Logger.Infof("Remove clickhouse container data dir %s", dataDir)
	_, err = s.Em.Runner.Exec(ctx, "rm", "-rf", dataDir)
	if err != nil {
		return errors.Annotatef(err, "rm %s", dataDir)
	}
	logDir := path.Join(s.Runtime.Services.Clickhouse.WorkDir, "log")
	s.Logger.Infof("Remove clickhouse container log dir %s", logDir)
	_, err = s.Em.Runner.Exec(ctx, "rm", "-rf", logDir)
	if err != nil {
		return errors.Annotatef(err, "rm %s", logDir)
	}
	configDir := path.Join(s.Runtime.Services.Clickhouse.WorkDir, "config.d")
	s.Logger.Infof("Remove clickhouse container config dir %s", configDir)
	_, err = s.Em.Runner.Exec(ctx, "rm", "-rf", configDir)
	if err != nil {
		return errors.Annotatef(err, "rm %s", configDir)
	}
	sqlDir := path.Join(s.Runtime.Services.Clickhouse.WorkDir, "sql")
	s.Logger.Infof("Remove clickhouse container sql init dir %s", sqlDir)
	_, err = s.Em.Runner.Exec(ctx, "rm", "-rf", sqlDir)
	if err != nil {
		return errors.Annotatef(err, "rm %s", sqlDir)
	}

	s.Logger.Infof("ClickHouse container %s successfully removed", containerName)
	return nil
}
