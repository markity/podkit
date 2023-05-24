package container

import (
	"podkit/frontend/cmd/container/exec"
	"podkit/frontend/cmd/container/ls"
	"podkit/frontend/cmd/container/remove"
	"podkit/frontend/cmd/container/restart"
	"podkit/frontend/cmd/container/stop"

	"github.com/spf13/cobra"
)

func init() {
	ContainerCmd.AddCommand(exec.ExecCmd)
	ContainerCmd.AddCommand(remove.RemoveCmd)
	ContainerCmd.AddCommand(restart.RestartCmd)
	ContainerCmd.AddCommand(stop.StopCmd)
	ContainerCmd.AddCommand(ls.LsCmd)
}

var ContainerCmd = &cobra.Command{
	Use:   "container",
	Short: "container command is used to manage containers",
}
