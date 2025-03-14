package meta

import (
	"embed"
)

var (
	//go:embed templates/*.tmpl
	templatesFs embed.FS

	// MetaMainAppTomlTmpl is the template content of meta_main_app.toml
	MetaMainAppTomlTmpl []byte
	// MetaMainLauncherTomlTmpl is the template content of meta_main_launcher.toml
	MetaMainLauncherTomlTmpl []byte
	// MetaMainTomlTmpl is the template content of meta_main.toml
	MetaMainTomlTmpl []byte
)

func init() {
	var err error
	MetaMainAppTomlTmpl, err = templatesFs.ReadFile("templates/meta_main_app.toml.tmpl")
	if err != nil {
		panic(err)
	}

	MetaMainLauncherTomlTmpl, err = templatesFs.ReadFile("templates/meta_main_launcher.toml.tmpl")
	if err != nil {
		panic(err)
	}

	MetaMainTomlTmpl, err = templatesFs.ReadFile("templates/meta_main.toml.tmpl")
	if err != nil {
		panic(err)
	}
}
