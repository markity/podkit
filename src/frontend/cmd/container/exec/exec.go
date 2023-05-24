package exec

import (
	"podkit/frontend/tools"

	"github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
	Use:   "exec CONTAINER_ID COMMAND_PATH",
	Short: "exec command in container specified by id",
	Args:  cobra.MatchAll(cobra.ExactArgs(2)),
	Run: func(cmd *cobra.Command, args []string) {
		flock := tools.FlockManager{}
		err := flock.Init("/var/lib/podkit/lock")
		if err != nil {
			panic(err)
		}
		err = flock.Lock()
		if err != nil {
			panic(err)
		}
		defer flock.Release()

		println("exec: not yet")
	},
}
