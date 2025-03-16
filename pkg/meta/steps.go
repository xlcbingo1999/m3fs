// Copyright 2025 Open3FS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
