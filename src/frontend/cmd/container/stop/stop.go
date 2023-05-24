package stop

import (
	"podkit/frontend/tools"

	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop CONTAINER_ID",
	Short: "stop a container specified by id",
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

		println("stop: not yet")
	},
}
