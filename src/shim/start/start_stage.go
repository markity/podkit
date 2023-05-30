package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"podkit/frontend/tools"
	"syscall"

	"golang.org/x/sys/unix"
)

// stage1: 创建自己的守护进程, 进入stage2, 务必让stage2进入listen之后再结束程序
// 因此这里用了pipe做通讯
func startStage1() {
	pipeReader, pipeWriter := io.Pipe()
	cmd := exec.Command("podkit_shim", "start", "stage2", fmt.Sprint(ContainerID))
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
func startStage2() {
	syscall.Umask(0)

	// 需要等待stage3完成挂载, 创建设备节点, 创建软链接等工作之后再进入监听状态
	// 因此这里用了pipe
	cmd := exec.Command("podkit_shim", "start", "stage3", fmt.Sprint(ContainerID))
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter
	cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWPID,
	}

	// 子进程现在在新的namespace下, 运行的stage3
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	initProcPid := cmd.Process.Pid

	// 等待子进程完成文件挂载等工作
	bs := make([]byte, 1)
	_, err = pipeReader.Read(bs)
	if err != nil {
		panic(err)
	}

	// 告知主线程目前的监听状态
	listenFinished := make(chan struct{})
	listenClosed := make(chan struct{})
	go RunServer(initProcPid, listenFinished, listenClosed)

	// Server协程开始监听后就能告知父进程, 让stage1退出了, 然后start指令完成
	<-listenFinished
	_, err = syscall.Write(1, []byte{1})
	if err != nil {
		panic(err)
	}

	// 等待listener退出
	<-listenClosed

}

// 挂载文件, exec成为orphan_reaper
func startStage3() {
	syscall.Sethostname([]byte(fmt.Sprintf("container%d", ContainerID)))
	prefix := fmt.Sprintf("/var/lib/podkit/container/%d", ContainerID)

	// 把resolve.conf挂载到容器中
	err := syscall.Mount("/etc/resolv.conf", fmt.Sprintf("%s/etc/resolv.conf", prefix), "bind", syscall.MS_BIND|syscall.MS_RDONLY, "")
	if err != nil {
		panic(err)
	}

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

	err = unix.Mknod(fmt.Sprintf("%s/dev/null", prefix), 0666|syscall.S_IFCHR, int(tools.MakeDev(1, 3)))
	if err != nil {
		panic(err)
	}
	err = unix.Mknod(fmt.Sprintf("%s/dev/full", prefix), 0666|syscall.S_IFCHR, int(tools.MakeDev(1, 7)))
	if err != nil {
		panic(err)
	}
	err = unix.Mknod(fmt.Sprintf("%s/dev/console", prefix), 0640|syscall.S_IFCHR, int(tools.MakeDev(136, 0)))
	if err != nil {
		panic(err)
	}
	err = unix.Mknod(fmt.Sprintf("%s/dev/random", prefix), 0666|syscall.S_IFCHR, int(tools.MakeDev(1, 8)))
	if err != nil {
		panic(err)
	}
	err = unix.Mknod(fmt.Sprintf("%s/dev/urandom", prefix), 0666|syscall.S_IFCHR, int(tools.MakeDev(1, 9)))
	if err != nil {
		panic(err)
	}
	err = unix.Mknod(fmt.Sprintf("%s/dev/tty", prefix), 0666|syscall.S_IFCHR, int(tools.MakeDev(5, 0)))
	if err != nil {
		panic(err)
	}
	err = unix.Mknod(fmt.Sprintf("%s/dev/zero", prefix), 0666|syscall.S_IFCHR, int(tools.MakeDev(1, 5)))
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

	err = syscall.Exec("/bin/podkit_orphan_reaper", []string{"init"}, nil)
	if err != nil {
		panic(err)
	}
}
