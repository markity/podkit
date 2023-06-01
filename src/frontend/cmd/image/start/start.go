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
	Short: "start a container and print its id",
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

		ipPoll := tools.NewAddrPool("172.16.0.0/16")
		allContainers := []*json_struct.ContainerInfo{}
		allContainers = append(allContainers, runningInfo.ContainerRunning...)
		allContainers = append(allContainers, runningInfo.ContainerStopped...)
		ipPoll.Next()
		// 如果没有任何容器开启, 那么ipFetched
		ipFetched := ipPoll.Next().String()
		// 从0.2开始, 因为0.1是分配给u网桥的
		for _, v := range allContainers {
			ip := ipPoll.Next()
			// 检查是否把ip用完了
			if ip.String() == "172.16.255.255" {
				fmt.Println("all ips are used up, please remove some containers to create a new container")
				flock.Release()
				return
			}
			if ip.String() != v.IP {
				ipFetched = ip.String()
				break
			}
		}

		err = os.Mkdir(fmt.Sprintf("/var/lib/podkit/container/%d", currentContainerID), 0755)
		if err != nil {
			panic(err)
		}

		// 解压文件
		imageFilePath := fmt.Sprintf("/var/lib/podkit/images/%s", imageInfo.ImageTarFilename[imageName])
		fmt.Printf("Extracting %v\n", imageFilePath)

		// TODO: 用golang提供的函数解压
		err = exec.Command("tar", "-xvf", imageFilePath, "-C", fmt.Sprintf("/var/lib/podkit/container/%d", currentContainerID)).Run()
		if err != nil {
			panic(err)
		}

		// 开启shim程序, 等待stage1执行完毕, stage1执行完毕后socket文件已经创建且进入监听状态
		shimCmd := exec.Command("podkit_shim", "start", "stage1", fmt.Sprintf("%d", currentContainerID), ipFetched)
		err = shimCmd.Run()
		if err != nil {
			panic(err)
		}

		// 更新running_info.json
		runningInfo.ContainerRunning = append(runningInfo.ContainerRunning, &json_struct.ContainerInfo{
			ContainerID:        currentContainerID,
			ContainerImageName: imageName,
			IP:                 ipFetched,
		})

		runningInfo.ContainerIDCount++

		err = runningInfo.WriteToFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		fmt.Printf("succeed: container id is %d\n", currentContainerID)
	},
}
