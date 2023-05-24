package ls

import (
	"fmt"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"

	"github.com/spf13/cobra"
)

var LsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list all containers",
	Args:  cobra.MatchAll(cobra.ExactArgs(0)),
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

		runningInfo := json_struct.RunningInfoStruct{}
		err = runningInfo.ParseFromFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		fmt.Println("---stopped:")
		for _, v := range runningInfo.ContainerStopped {
			fmt.Printf("%d %s\n", v.ContainerID, v.ContainerImageName)
		}
		fmt.Println("---running:")
		for _, v := range runningInfo.ContainerRunning {
			fmt.Printf("%d %s\n", v.ContainerID, v.ContainerImageName)
		}
	},
}
