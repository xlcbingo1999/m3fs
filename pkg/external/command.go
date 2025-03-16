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

package external

import (
	"context"
	"fmt"
	"strings"
)

// Command define command
type Command struct {
	runner RunnerInterface

	cmdName string
	args    []string
}

// Command gets the command
func (cmd *Command) Command() string {
	if len(cmd.args) > 0 {
		return cmd.cmdName + " " + strings.Join(cmd.args, " ")
	}
	return cmd.cmdName
}

// AppendArgs append new args to current args
func (cmd *Command) AppendArgs(args ...any) {
	for _, arg := range args {
		cmd.args = append(cmd.args, fmt.Sprintf("%v", arg))
	}
}

// Exec execute the command
func (cmd *Command) Exec(ctx context.Context) (out string, err error) {
	if cmd.cmdName == "" {
		return "", fmt.Errorf("No command")
	}
	return cmd.runner.NonSudoExec(ctx, cmd.cmdName, cmd.args...)
}

// SudoExec execute the command
func (cmd *Command) SudoExec(ctx context.Context) (out string, err error) {
	if cmd.cmdName == "" {
		return "", fmt.Errorf("No command")
	}
	return cmd.runner.Exec(ctx, cmd.cmdName, cmd.args...)
}

func (cmd *Command) String() string {
	return fmt.Sprintf("cmd: %s", cmd.Command())
}

// NewCommand inits a new command
func NewCommand(cmdName string, args ...any) *Command {
	cmd := &Command{
		cmdName: cmdName,
	}
	cmd.AppendArgs(args...)

	return cmd
}
