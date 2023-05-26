package start

import (
	"fmt"
	"os"
	"os/exec"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"
	"syscall"

	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start IMAGE_NAME",
	Short: "start a container and print its uuid",
	Args:  cobra.MatchAll(cobra.ExactArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		syscall.Umask(0)

		imageName := args[0]
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

		imageInfo := json_struct.ImageInfoStruct{}
		err = imageInfo.ParseFromFile("/var/lib/podkit/images/image_info.json")
		if err != nil {
			panic(err)
		}

		exists := false

		for k := range imageInfo.ImageTarFilename {
			if k == imageName {
				exists = true
				break
			}
		}

		if !exists {
			fmt.Printf("%v does not exists\n", imageName)
			return
		}

		runningInfo := json_struct.RunningInfoStruct{}
		err = runningInfo.ParseFromFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		currentContainerID := runningInfo.ContainerIDCount

		err = os.Mkdir(fmt.Sprintf("/var/lib/podkit/container/%d", currentContainerID), 0755)
		if err != nil {
			panic(err)
		}

		// 解压文件
		imageFilePath := fmt.Sprintf("/var/lib/podkit/images/%s", imageInfo.ImageTarFilename[imageName])
		fmt.Printf("Extracting %v\n", imageFilePath)

		// TODO: 用golang提供的函数解压
		exec.Command("tar", "-xvf", imageFilePath, "-C", fmt.Sprintf("/var/lib/podkit/container/%d", currentContainerID)).Run()

		// 开启shim程序, 等待stage1执行完毕, stage1执行完毕后socket文件已经创建且进入监听状态
		shimCmd := exec.Command("podkit_shim", "start", "stage1", fmt.Sprintf("%d", currentContainerID))
		shimCmd.Run()

		// 更新running_info.json
		runningInfo.ContainerRunning = append(runningInfo.ContainerRunning, &json_struct.ContainerInfo{
			ContainerID:        currentContainerID,
			ContainerImageName: imageName,
		})

		runningInfo.ContainerIDCount++

		bs := runningInfo.MustMarshalToBytes()

		runningInfoFile, err := os.OpenFile("/var/lib/podkit/running_info.json", os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			panic(err)
		}
		defer runningInfoFile.Close()

		_, err = runningInfoFile.Write(bs)
		if err != nil {
			panic(err)
		}

		fmt.Printf("succeed: container id is %d\n", currentContainerID)
	},
}
