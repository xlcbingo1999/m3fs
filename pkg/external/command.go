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
