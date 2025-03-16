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

package common

import (
	"github.com/davecgh/go-spew/spew"
)

var (
	spewConfig = &spew.ConfigState{
		Indent:         "\t",
		MaxDepth:       3,
		DisableMethods: true,
	}
)

// PrettyDump prints Golang objects in a beautiful way.
func PrettyDump(a ...any) {
	spewConfig.Dump(a...)
}

// PrettySdump prints Golang objects in a beautiful way to string.
func PrettySdump(a ...any) string {
	return spewConfig.Sdump(a...)
}
