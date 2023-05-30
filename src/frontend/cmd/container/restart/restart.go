package restart

import (
	"errors"
	"fmt"
	"os/exec"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var RestartCmd = &cobra.Command{
	Use:   "restart CONTAINER_ID",
	Short: "restart a container specified by id, which must be stopped",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("num of arguments is wrong")
		}

		_, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.New("CONTAINER_ID must be a number")
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
			fmt.Printf("container %d is already running\n", id)
			return
		}

		println("restarting...")

		shimCmd := exec.Command("podkit_shim", "start", "stage1", fmt.Sprintf("%d", id))
		shimCmd.Run()

		newRunning := make([]*json_struct.ContainerInfo, 0)
		newStopped := make([]*json_struct.ContainerInfo, 0)
		var replaced *json_struct.ContainerInfo
		for _, v := range runningInfo.ContainerStopped {
			if v.ContainerID != id {
				newStopped = append(newStopped, v)
			} else {
				replaced = v
			}
		}

		newRunning = append(newRunning, replaced)
		newRunning = append(newRunning, runningInfo.ContainerRunning...)

		runningInfo.ContainerRunning = newRunning
		runningInfo.ContainerStopped = newStopped

		err = runningInfo.WriteToFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		fmt.Printf("container %d restarted successfully\n", id)
	},
}
