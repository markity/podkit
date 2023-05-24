package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// stage1: 创建自己的守护进程, 进入stage2, 务必让stage2进入listen之后再结束程序
// 因此这里用了pipe做通讯
func startStage1(id int) {
	pipeReader, pipeWriter := io.Pipe()
	cmd := exec.Command("podkit_shim", "start", "stage2", fmt.Sprintf("%d", id))
	cmd.Stdout = pipeWriter
	_, err := syscall.Setsid()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	// 等待stage2进入监听状态之后才能结束这个程序
	bs := make([]byte, 1)
	_, err = pipeReader.Read(bs)
	if err != nil {
		panic(err)
	}
}

// stage2: 需要fork一个子进程作为容器的init进程(podkit_orphan_reaper), 同时监听过来的网络连接, 向容器插入进程
func startStage2(id int) {
	syscall.Umask(0)

	// 需要等待stage3完成挂载, 创建设备节点, 创建软链接等工作之后再进入监听状态
	// 因此这里用了pipe
	cmd := exec.Command("podkit_shim", "start", "stage3", fmt.Sprintf("%d", id))
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}

	// 子进程现在在新的namespace下, 运行的stage3
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	childPid := cmd.Process.Pid

	// 等待子进程完成文件挂载等工作
	bs := make([]byte, 1)
	_, err = pipeReader.Read(bs)
	if err != nil {
		panic(err)
	}

	listenFinished := make(chan struct{})
	listenClosed := make(chan struct{})
	go RunServer(childPid, id, listenFinished, listenClosed)

	<-listenFinished
	_, err = syscall.Write(1, []byte{1})
	if err != nil {
		panic(err)
	}

	// 等待listener退出
	<-listenClosed

	// 取消挂载
	prefix := fmt.Sprintf("/var/lib/podkit/container/%d", id)
	err = syscall.Unmount(fmt.Sprintf("%s/dev/pts", prefix), 0)
	if err != nil {
		panic(err)
	}
	err = syscall.Unmount(fmt.Sprintf("%s/dev/mqueue", prefix), 0)
	if err != nil {
		panic(err)
	}
	err = syscall.Unmount(fmt.Sprintf("%s/dev/shm", prefix), 0)
	if err != nil {
		panic(err)
	}
	err = syscall.Unmount(fmt.Sprintf("%s/dev", prefix), 0)
	if err != nil {
		panic(err)
	}
	err = syscall.Unmount(fmt.Sprintf("%s/tmp", prefix), 0)
	if err != nil {
		panic(err)
	}
	err = syscall.Unmount(fmt.Sprintf("%s/sys", prefix), 0)
	if err != nil {
		panic(err)
	}
	err = syscall.Unmount(fmt.Sprintf("%s/proc", prefix), 0)
	if err != nil {
		panic(err)
	}

	// 杀死init进程
	err = syscall.Kill(childPid, syscall.SIGKILL)
	if err != nil {
		panic(err)
	}
}

// 挂载文件, exec成为orphan_reaper
func startStage3(id int) {
	prefix := fmt.Sprintf("/var/lib/podkit/container/%d", id)

	// 挂载proc sys tmp dev
	err := syscall.Mount("proc", fmt.Sprintf("%s/proc", prefix), "proc", 0, "")
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
	// 通知父进程挂载完毕
	_, err = os.Stdout.Write([]byte{1})
	if err != nil {
		panic(err)
	}

	// 现在可以安全关闭所有引用的fd
	// TODO: 不能syscall.Close
	// n, err := sysconf.Sysconf(sysconf.SC_OPEN_MAX)
	// if err != nil {
	// 	f.Write([]byte("2\n"))
	// 	panic(err)
	// }

	err = syscall.Exec("/bin/podkit_orphan_reaper", nil, nil)
	if err != nil {
		panic(err)
	}
}
