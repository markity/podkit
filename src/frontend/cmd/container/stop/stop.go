package stop

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	commpacket "podkit/comm_packet"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"
	"strconv"

	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop CONTAINER_ID",
	Short: "stop a container specified by id",
	Args: cobra.MatchAll(func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("one argument is required")
		}
		_, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		return nil
	}),
	Run: func(cmd *cobra.Command, args []string) {
		// 参数已经校验, 这里不做检查
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
				break
			}
		}

		if !exists {
			for _, v := range runningInfo.ContainerStopped {
				if v.ContainerID == id {
					exists = true
					running = false
					break
				}
			}
		}

		if !exists {
			fmt.Printf("container %d does not exist\n", id)
			return
		}

		if !running {
			fmt.Printf("container %d had been stopped already\n", id)
			return
		}

		// 正式与shim进行交互
		conn, err := net.Dial("unix", fmt.Sprintf("/var/lib/podkit/socket/%d.sock", id))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		// 4字节的包大小前缀
		_, err = conn.Write(commpacket.DoPack(4, (&commpacket.ClientCloseContainer{}).MustMarshalToBytes()))
		if err != nil {
			panic(err)
		}

		lengthBytes := make([]byte, 4)
		_, err = io.ReadFull(conn, lengthBytes)
		if err != nil {
			panic(err)
		}

		packetBytes := make([]byte, binary.BigEndian.Uint32(lengthBytes))
		_, err = io.ReadFull(conn, packetBytes)
		if err != nil {
			panic(err)
		}

		// panic检测, 需要它真正的是ClientCloseContainer
		_ = commpacket.ClientParsePacket(packetBytes).(*commpacket.ServerNotifyContainerClosed)

		newRunning := make([]*json_struct.ContainerInfo, 0)
		var replacedContainer *json_struct.ContainerInfo
		for _, v := range runningInfo.ContainerRunning {
			if v.ContainerID != id {
				newRunning = append(newRunning, v)
			} else {
				replacedContainer = v
			}
		}

		newStopped := make([]*json_struct.ContainerInfo, 0)
		newStopped = append(newStopped, replacedContainer)
		newStopped = append(newStopped, runningInfo.ContainerStopped...)
		runningInfo.ContainerRunning = newRunning
		runningInfo.ContainerStopped = newStopped
		runningInfoFile, err := os.OpenFile("/var/lib/podkit/running_info.json", os.O_WRONLY|os.O_TRUNC, 0)
		if err != nil {
			panic(err)
		}
		runningInfoFile.Write(runningInfo.MustMarshalToBytes())

		fmt.Printf("conatiner %d closed successfully\n", id)
	},
}
