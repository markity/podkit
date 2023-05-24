package start

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"

	"github.com/spf13/cobra"
)

var StartCmd = &cobra.Command{
	Use:   "start CONTAINER_NAME",
	Short: "start a container and print its uuid",
	Args:  cobra.MatchAll(cobra.ExactArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		containerName := args[0]
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
			if k == containerName {
				exists = true
				break
			}
		}

		if !exists {
			fmt.Printf("%v does not exists\n", containerName)
			return
		}

		runningInfo := json_struct.RunningInfoStruct{}
		err = runningInfo.ParseFromFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		currentID := runningInfo.ContainerIDCount

		err = os.Mkdir(fmt.Sprintf("/var/lib/podkit/container/%d", currentID), 0755)
		if err != nil {
			panic(err)
		}

		// 解压文件
		imageFilePath := fmt.Sprintf("/var/lib/podkit/images/%s", imageInfo.ImageTarFilename[containerName])
		fmt.Printf("Extracting %v\n", imageFilePath)
		tarFile, err := os.Open(imageFilePath)
		if err != nil {
			panic(err)
		}
		defer tarFile.Close()
		tarReader := tar.NewReader(tarFile)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				panic(err)
			}

			switch header.Typeflag {
			case tar.TypeDir:
				err := os.Mkdir(fmt.Sprintf("/var/lib/podkit/container/%d/%s", currentID, header.Name), 0755)
				if err != nil {
					panic(err)
				}
			case tar.TypeReg:
			case tar.TypeSymlink:
			case tar.TypeLink:
				f, err := os.Create(fmt.Sprintf("/var/lib/podkit/container/%d/%s", currentID, header.Name))
				if err != nil {
					panic(err)
				}
				defer f.Close()

				_, err = io.Copy(f, tarReader)
				if err != nil {
					panic(err)
				}
			default:
				fmt.Println(header.Typeflag)
				panic(errors.New("cannot recognize image tar file"))
			}
		}

		// 开启shim程序, 等待stage1执行完毕, stage1执行完毕后socket文件已经创建且进入监听状态
		shimCmd := exec.Command("podkit_shim", "stage1", fmt.Sprintf("%d", currentID))
		shimCmd.Run()

		// 更新running_info.json
		runningInfo.ContainerRunning = append(runningInfo.ContainerRunning, &json_struct.ContainerInfo{
			ContainerID: currentID,
		})

		runningInfo.ContainerIDCount++

		bs, err := runningInfo.MarshalToBytes()
		if err != nil {
			panic(err)
		}

		runningInfoFile, err := os.OpenFile("/var/lib/podkit/running_info.json", os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			panic(err)
		}
		defer runningInfoFile.Close()

		_, err = runningInfoFile.Write(bs)
		if err != nil {
			panic(err)
		}

		fmt.Printf("succeed: container id is %d\n", currentID)
	},
}
