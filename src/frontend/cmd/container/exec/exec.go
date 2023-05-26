package exec

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

	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	ExecCmd.PersistentFlags().BoolP("interactive", "i", false, "open pty device and enter interactive mode, default running in deamon mode")
}

var ExecCmd = &cobra.Command{
	Use:   "exec CONTAINER_ID COMMAND_PATH",
	Short: "exec command in container specified by id",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
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
		_ = args[1]

		execCmd := args[1]

		flock := tools.FlockManager{}
		err := flock.Init("/var/lib/podkit/lock")
		if err != nil {
			panic(err)
		}
		err = flock.Lock()
		if err != nil {
			panic(err)
		}

		interactive, err := cmd.Flags().GetBool("interactive")
		if err != nil {
			panic(err)
		}

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

		if !running {
			fmt.Printf("container %d stopped, start it first\n", id)
			return
		}

		conn, err := net.Dial("unix", fmt.Sprintf("/var/lib/podkit/socket/%d.sock", id))
		if err != nil {
			panic(err)
		}

		// 与stage2的监听进程建立连接
		if interactive {
			rows, cols, err := pty.Getsize(os.Stdin)
			if err != nil {
				panic(err)
			}

			_, err = conn.Write(tools.DoPackWith4Bytes((&commpacket.ClientRequestExecInteractive{Rows: rows, Cols: cols, Command: execCmd}).MustMarshalToBytes()))
			if err != nil {
				panic(err)
			}
			// 读取第一个包, 查看是否有这个命令

			packetBytes, err := tools.ReadPacketWith4BytesLengthHeader(conn)
			if err != nil {
				panic(err)
			}

			if !commpacket.ClientParsePacket(packetBytes).(*commpacket.ServerInteractiveCommandResp).CommandExists {
				fmt.Println("command does not exists, check it again")
				flock.Release()
				return
			}

			// 此处要释放文件锁, 让其它podkit程序得以运行
			flock.Release()

			oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
			if err != nil {
				panic(err)
			}

			var errorsChannel = make(chan error, 2)

			// conn reader
			readFromConn := make(chan interface{})
			go func() {
				for {
					lengthBytes := make([]byte, 4)
					_, err := io.ReadFull(conn, lengthBytes)
					if err != nil {
						<-readFromConn
						return
					}

					packetBytes := make([]byte, binary.BigEndian.Uint32(lengthBytes))
					_, err = io.ReadFull(conn, packetBytes)
					if err != nil {
						errorsChannel <- err
						return
					}

					readFromConn <- commpacket.ClientParsePacket(packetBytes)
				}
			}()

			// stdin reader
			readFromStdin := make(chan []byte)
			go func() {
				for {
					bs := make([]byte, 512)
					// 理论是这里永远不会出错
					n, err := os.Stdin.Read(bs)
					if err != nil {
						errorsChannel <- err
						return
					}

					newBs := make([]byte, n)
					copy(newBs, bs)
					readFromStdin <- newBs
				}
			}()

			for {
				select {
				case iface := <-readFromConn:
					switch packet := iface.(type) {
					case *commpacket.ServerNotifyInteractiveExecContainerClosed:
						term.Restore(int(os.Stdin.Fd()), oldState)
						conn.Close()
						fmt.Println("container is closed")
						return
					case *commpacket.ServerInteractiveCommandExited:
						term.Restore(int(os.Stdin.Fd()), oldState)
						conn.Close()
						fmt.Println("command exited")
						return
					case *commpacket.ServerSendPtyOutput:
						os.Stdout.Write([]byte(packet.Data))
					}
				case stdinBs := <-readFromStdin:
					_, err := conn.Write(tools.DoPackWith4Bytes((&commpacket.ClientSendPtyInput{Data: string(stdinBs)}).MustMarshalToBytes()))
					if err != nil {
						panic(err)
					}
				case err := <-errorsChannel:
					panic(err)
				}
			}
		} else {
			_, err := conn.Write((tools.DoPackWith4Bytes((&commpacket.ClientRequestExecBackground{Command: execCmd}).MustMarshalToBytes())))
			if err != nil {
				panic(err)
			}

			packetBytes, err := tools.ReadPacketWith4BytesLengthHeader(conn)
			if err != nil {
				panic(err)
			}

			packet := commpacket.ClientParsePacket(packetBytes)
			switch resp := packet.(type) {
			case *commpacket.ServerExecBackgroundResp:
				if resp.CommandExists {
					fmt.Println("ok, command now is running")
				} else {
					fmt.Println("failed, command not found")
				}
				flock.Release()
				return
			default:
				panic(errors.New("unexpected error"))
			}
		}
	},
}
