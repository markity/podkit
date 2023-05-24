package restart

import (
	"podkit/frontend/tools"

	"github.com/spf13/cobra"
)

var RestartCmd = &cobra.Command{
	Use:   "restart CONTAINER_ID",
	Short: "restart a container specified by id, which must be stopped",
	Args:  cobra.MatchAll(cobra.ExactArgs(1)),
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

		println("restart: not yet")
	},
}
