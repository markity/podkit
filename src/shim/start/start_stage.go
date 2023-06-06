package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"podkit/frontend/tools"
	shim_tools "podkit/shim/tools"
	"syscall"

	"github.com/docker/libcontainer/netlink"
	"github.com/milosgajdos/tenus"
	vlink "github.com/vishvananda/netlink"

	"golang.org/x/sys/unix"
)

// stage1: 创建自己的守护进程, 进入stage2, 务必让stage2进入listen之后再结束程序
// 因此这里用了pipe做通讯
func startStage1(containerID int, ipaddr string) {
	pipeReader, pipeWriter := io.Pipe()
	cmd := exec.Command("podkit_shim", "start", "stage2", fmt.Sprint(containerID), ipaddr)
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

// stage2: 需要fork一个子进程作为容器的init进程(podkit_orphan_reaper), 同时-iru监听过来的网络连接, 向容器插入进程
func startStage2(containerID int, ipaddr string) {
	syscall.Umask(0)

	// 需要等待stage3完成挂载, 创建设备节点, 创建软链接等工作之后再进入监听状态
	// 因此这里用了pipe
	cmd := exec.Command("podkit_shim", "start", "stage3", fmt.Sprint(containerID), ipaddr)
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter
	cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET,
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

	// 配置net ns的网络

	// 在容器中loopback网卡
	containerInitProcNetNSPath := fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/net", containerID)
	shim_tools.MustDoInNS(containerInitProcNetNSPath, func() {
		link, err := vlink.LinkByName("lo")
		if err != nil {
			panic(err)
		}

		err = vlink.LinkSetUp(link)
		if err != nil {
			panic(err)
		}
	})

	bridge, err := tenus.BridgeFromName("pkbr0")
	if err != nil {
		panic(err)
	}

	// 创建veth网卡
	vethName := fmt.Sprintf("pk%d", containerID)
	vethPeerName := fmt.Sprintf("pkpeer%d", containerID)
	err = netlink.NetworkCreateVethPair(vethName, vethPeerName, 1000)
	if err != nil {
		panic(err)
	}

	vethIface, err := net.InterfaceByName(vethName)
	if err != nil {
		panic(err)
	}
	vethPeerIface, err := net.InterfaceByName(vethPeerName)
	if err != nil {
		panic(err)
	}

	// 随机产生mac
	err = netlink.SetMacAddress(vethName, tools.RandMacAddr())
	if err != nil {
		panic(err)
	}

	// 将veth移入namespace
	netNSFile, err := os.OpenFile(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/net", containerID), os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	err = netlink.NetworkSetNsFd(vethIface, int(netNSFile.Fd()))
	if err != nil {
		panic(err)
	}
	netNSFile.Close()

	// 设置新名字并启动
	shim_tools.MustDoInNS(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/net", containerID), func() {
		err = netlink.ChangeName(vethIface, "eth0")
		if err != nil {
			panic(err)
		}
		// 设置veth的ip
		err = netlink.NetworkLinkAddIp(vethIface, net.ParseIP(ipaddr).To4(), &net.IPNet{IP: net.IPv4(172, 1, 0, 0), Mask: net.IPv4Mask(255, 255, 0, 0)})
		if err != nil {
			panic(err)
		}
		err = netlink.NetworkLinkUp(vethIface)
		if err != nil {
			panic(err)
		}

		err = netlink.AddRoute("0.0.0.0/0", "0.0.0.0", "172.16.0.1", "eth0")
		if err != nil {
			panic(err)
		}
	})

	// 将vethpeer添加进网桥
	err = bridge.AddSlaveIfc(vethPeerIface)
	if err != nil {
		panic(err)
	}

	// 告知主线程目前的监听状态
	listenFinished := make(chan struct{})
	listenClosed := make(chan struct{})
	go RunServer(containerID, initProcPid, listenFinished, listenClosed)

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
// stage3已经创建了新的ns
func startStage3(containerID int, ipaddr string) {
	syscall.Sethostname([]byte(fmt.Sprintf("container%d", containerID)))
	prefix := fmt.Sprintf("/var/lib/podkit/container/%d", containerID)

	// 把resolv.conf挂载到容器中
	f, err := os.Create(fmt.Sprintf("%s/etc/resolv.conf", prefix))
	if err != nil {
		panic(err)
	}
	f.Close()
	err = syscall.Mount("/etc/resolv.conf", fmt.Sprintf("%s/etc/resolv.conf", prefix), "bind", syscall.MS_BIND|syscall.MS_RDONLY, "")
	if err != nil {
		panic(err)
	}

	// 有的rootfs没有这两个文件夹, 需要手动创建
	err = os.Mkdir(fmt.Sprintf("%s/proc", prefix), syscall.S_IWUSR|syscall.S_IRUSR|syscall.S_IXUSR|syscall.S_IRGRP|syscall.S_IXGRP|syscall.S_IXOTH)
	if err != nil && err.(*os.PathError).Err != syscall.EEXIST {
		panic(err)
	}
	err = os.Mkdir(fmt.Sprintf("%s/sys", prefix), syscall.S_IWUSR|syscall.S_IRUSR|syscall.S_IXUSR|syscall.S_IRGRP|syscall.S_IXGRP|syscall.S_IXOTH)
	if err != nil && err.(*os.PathError).Err != syscall.EEXIST {
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

	err = os.Chmod(fmt.Sprintf("%s/dev", prefix), 0751)
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
	// err = unix.Mknod(fmt.Sprintf("%s/dev/console", prefix), 0640|syscall.S_IFCHR, int(tools.MakeDev(136, 0)))
	// if err != nil {
	// 	panic(err)
	// }
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

	err = syscall.Chdir(fmt.Sprintf("%s/dev", prefix))
	if err != nil {
		panic(err)
	}
	err = syscall.Symlink("pts/ptmx", "ptmx")
	if err != nil {
		panic(err)
	}

	err = syscall.Symlink("/proc/self/fd", "fd")
	if err != nil {
		panic(err)
	}

	err = syscall.Symlink("/proc/self/fd/0", "stdin")
	if err != nil {
		panic(err)
	}
	err = syscall.Symlink("/proc/self/fd/1", "stdout")
	if err != nil {
		panic(err)
	}
	err = syscall.Symlink("/proc/self/fd/2", "stderr")
	if err != nil {
		panic(err)
	}
	err = syscall.Symlink("/proc/kcore", "core")
	if err != nil {
		panic(err)
	}

	// 隔离文件
	err = os.Chdir(prefix)
	if err != nil {
		panic(err)
	}
	err = os.Chdir("/")
	if err != nil {
		panic(err)
	}

	// 通知父进程挂载完毕
	_, err = os.Stdout.Write([]byte{1})
	if err != nil {
		panic(err)
	}

	err = syscall.Exec("/bin/podkit_orphan_reaper", []string{"podkit_orphan_reaper"}, nil)
	if err != nil {
		panic(err)
	}
}
