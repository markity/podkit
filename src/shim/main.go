package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/tklauser/go-sysconf"
)

// podkit_shim

// podkit_shim stagen 容器id(int类型)
func main() {
	if len(os.Args) != 3 {
		return
	}
	if os.Geteuid() != 0 {
		return
	}

	stage := os.Args[1]
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return
	}

	switch stage {
	case "stage1":
		stage1(id)
	case "stage2":
		stage2(id)
	case "stage3":
		stage3(id)
	}
}

// stage1: 创建自己的守护进程, 进入stage2
func stage1(id int) {
	cmd := exec.Command("podkit_shim", "stage2", fmt.Sprintf("%d", id))
	_, err := syscall.Setsid()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
}

// stage2: 需要fork一个子进程作为容器的init进程(podkit_orphan_reaper), 同时监听过来的网络连接, 向容器插入进程
// 需要的参数: 容器的id
func stage2(id int) {
	syscall.Umask(0)

	cmd := exec.Command("podkit_shim", "stage3", fmt.Sprintf("%d", id))
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}

	// 子进程现在在新的namespace下, 运行的stage3
	cmd.Start()

	// 等待子进程完成文件挂载等工作
	bs := make([]byte, 1)
	_, err := pipeReader.Read(bs)
	if err != nil {
		panic(err)
	}

	// 下面监听网络连接, 操作容器
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: fmt.Sprintf("/var/lib/podkit/socket/%d.sock", id), Net: "unix"})
	if err != nil {
		panic(err)
	}

	for {
		c, err := listener.Accept()
		if err != nil {
			continue
		}

		go func() {
		}()
	}
}

// 挂载文件, exec成为orphan_reaper
func stage3(id int) {
	n, err := sysconf.Sysconf(sysconf.SC_OPEN_MAX)
	if err != nil {
		panic(err)
	}
	syscall.Close(0)
	for i := 2; i < int(n); i++ {
		syscall.Close(i)
	}

	prefix := fmt.Sprintf("/var/lib/podkit/container/%d", id)

	// 挂载proc sys tmp dev
	err = syscall.Mount("proc", fmt.Sprintf("%s/proc", prefix), "proc", 0, "")
	if err != nil {
		panic(err)
	}
	err = syscall.Mount("sysfs", fmt.Sprintf("%s/sys", prefix), "sysfs", 0, "")
	if err != nil {
		panic(err)
	}
	err = syscall.Mount("tmpfs", fmt.Sprintf("%s/tmp", prefix), "tmpfs", 0, "")
	if err != nil {
		panic(err)
	}
	err = syscall.Mount("tmpfs", fmt.Sprintf("%s/dev", prefix), "tmpfs", 0, "")
	if err != nil {
		panic(err)
	}

	// 创建 /dev/pts /dev/shm /dev/mqueue /dev/pts这些设备文件夹
	err = os.Mkdir(fmt.Sprintf("%s/dev/shm", prefix), 0700)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(fmt.Sprintf("%s/dev/mqueue", prefix), 0700)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(fmt.Sprintf("%s/dev/pts", prefix), 0700)
	if err != nil {
		panic(err)
	}

	// 挂载上面创建的文件夹
	err = syscall.Mount("shm", fmt.Sprintf("%s/dev/shm", prefix), "tmpfs", 0, "")
	if err != nil {
		panic(err)
	}
	err = syscall.Mount("mqueue", fmt.Sprintf("%s/dev/mqueue", prefix), "mqueue", 0, "")
	if err != nil {
		panic(err)
	}
	err = syscall.Mount("devpts", fmt.Sprintf("%s/dev/pts", prefix), "devpts", 0, "")
	if err != nil {
		panic(err)
	}

	os.Stdout.Write([]byte("ok"))
	syscall.Close(1)

	err = syscall.Exec("/bin/podkit_orphan_reaper", nil, nil)
	if err != nil {
		panic(err)
	}
}
