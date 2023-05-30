package remove

import (
	"errors"
	"fmt"
	"os"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove CONTAINER_ID",
	Short: "remove a container specified by id, which must be stopped",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("one argument is required")
		}
		_, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := strconv.Atoi(args[0])

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

		exists := false
		running := false
		for _, v := range runningInfo.ContainerRunning {
			if v.ContainerID == id {
				exists = true
				running = true
			}
		}

		if !exists {
			for _, v := range runningInfo.ContainerStopped {
				if v.ContainerID == id {
					exists = true
					running = false
				}
			}
		}

		if !exists {
			fmt.Printf("container %d does not exists\n", id)
			return
		}

		if running {
			fmt.Printf("container %d is running, stop it first\n", id)
			return
		}

		newStpped := make([]*json_struct.ContainerInfo, 0)
		for _, v := range runningInfo.ContainerStopped {
			if v.ContainerID != id {
				newStpped = append(newStpped, v)
			}
		}
		runningInfo.ContainerStopped = newStpped

		err = runningInfo.WriteToFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		err = os.RemoveAll(fmt.Sprintf("/var/lib/podkit/container/%d", id))
		if err != nil {
			panic(err)
		}

		fmt.Printf("removed container %d successfully\n", id)
	},
}
