package monitor

import (
	"context"
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/open3fs/m3fs/pkg/errors"
	"github.com/open3fs/m3fs/pkg/task"
)

var (
	//go:embed templates/*
	templatesFs embed.FS

	// MonitorCollectorMainTmpl is the template content of monitor_collector_main.toml
	MonitorCollectorMainTmpl []byte
)

func init() {
	var err error
	MonitorCollectorMainTmpl, err = templatesFs.ReadFile("templates/monitor_collector_main.tmpl")
	if err != nil {
		panic(err)
	}
}

type genMonitorConfigStep struct {
	task.BaseStep
}

func (s *genMonitorConfigStep) Execute(context.Context) error {
	fileName := "monitor_collector_main.toml"
	tmpl, err := template.New(fileName).Parse(string(MonitorCollectorMainTmpl))
	if err != nil {
		return errors.Annotate(err, "parse monitor_collector_main.toml template")
	}

	tempDir, err := os.MkdirTemp(os.TempDir(), "s3fs-monitor.")
	if err != nil {
		return errors.Trace(err)
	}
	configPath := filepath.Join(tempDir, fileName)
	file, err := os.Create(configPath)
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.Logger.Warnf("Failed to close file %+v", err)
		}
	}()
	err = tmpl.Execute(file, map[string]string{
		"Host":     "",
		"Password": "password",
		"Port":     "8999",
		"User":     "default",
	})
	if err != nil {
		return errors.Annotate(err, "write monitor_collector_main.toml")
	}
	s.Runtime.Store("tmp_monitor_collector_main_path", configPath)

	return nil
}
