package main

// podkit

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"podkit/frontend/cmd/container"
	"podkit/frontend/cmd/image"
	"podkit/frontend/json_struct"
	"podkit/frontend/tools"
	"syscall"

	"github.com/coreos/go-iptables/iptables"
	"github.com/lorenzosaino/go-sysctl"
	"github.com/milosgajdos/tenus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:               "podkit",
	Short:             "Podkit is an easy linux container kit",
	Long:              "Podkit helps you better understand the mechanism of docker\nIt provides main functions to get you understand how docker works",
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Root permission is required")
		return
	}

	lock := tools.FlockManager{}
	if err := lock.Init("/var/lib/podkit/lock"); err != nil {
		panic(err)
	}
	err := lock.Lock()
	if err != nil {
		panic(err)
	}

	// 检查是否reboot了, 如果reboot了应当把running_info.json里面所有的running标记为stopped
	lock_check_reboot := tools.FlockManager{}
	if err := lock_check_reboot.Init("/var/lib/podkit/lock_check_reboot"); err != nil {
		panic(err)
	}
	ok, err := lock_check_reboot.TryLock()
	if err != nil {
		panic(err)
	}
	if ok {
		pipeReader, pipeWriter := io.Pipe()
		lock_check_reboot.Release()
		cmd := exec.Command("podkit_lock_keeper")
		cmd.Stdout = pipeWriter
		cmd.Start()
		readBytes := make([]byte, 1)
		// 等待锁占用完成
		_, err := pipeReader.Read(readBytes)
		if err != nil {
			panic(err)
		}
		// 把所有的running变成stopped
		runningInfo := json_struct.RunningInfoStruct{}
		err = runningInfo.ParseFromFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		newRunning := make([]*json_struct.ContainerInfo, 0)
		newStopped := make([]*json_struct.ContainerInfo, 0)
		newStopped = append(newStopped, runningInfo.ContainerStopped...)
		for _, v := range runningInfo.ContainerRunning {
			prefix := fmt.Sprintf("/var/lib/podkit/container/%d", v.ContainerID)
			syscall.Unmount(fmt.Sprintf("%s/etc/resolv.conf", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/dev/pts", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/dev/mqueue", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/dev/shm", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/dev", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/tmp", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/sys", prefix), 0)
			syscall.Unmount(fmt.Sprintf("%s/proc", prefix), 0)
			os.Remove(fmt.Sprintf("/var/lib/podkit/socket/%d.sock", v.ContainerID))
			newStopped = append(newRunning, v)
		}

		runningInfo.ContainerStopped = newStopped
		runningInfo.ContainerRunning = newRunning

		err = runningInfo.WriteToFile("/var/lib/podkit/running_info.json")
		if err != nil {
			panic(err)
		}

		// 创建网桥
		br0, err := tenus.NewBridgeWithName("pkbr0")
		if err != nil {
			panic(err)
		}

		err = br0.SetLinkIp(net.IPv4(172, 16, 0, 1), &net.IPNet{IP: net.IPv4(172, 16, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)})
		if err != nil {
			panic(err)
		}

		err = br0.SetLinkUp()
		if err != nil {
			panic(err)
		}

		table, err := iptables.New()
		if err != nil {
			panic(err)
		}

		err = table.Append("nat", "POSTROUTING", "-s", "172.16.0.0/16", "-j", "MASQUERADE")
		if err != nil {
			panic(err)
		}

		err = sysctl.Set("net.ipv4.ip_forward", "1")
		if err != nil {
			panic(err)
		}
	}

	lock.Release()

	// podkit image ls
	// podkit image start ubuntu22.04
	RootCmd.AddCommand(image.ImageCmd)
	RootCmd.AddCommand(container.ContainerCmd)
	RootCmd.Execute()
}
