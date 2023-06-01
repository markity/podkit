package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/tklauser/go-sysconf"
	"golang.org/x/sys/unix"
)

// 被设置了path
func execBackground(commandPath string) {
	ipcNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/ipc", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	mntNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/mnt", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	pidNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/pid", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	utsNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/uts", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	err = unix.Setns(ipcNS, 0)
	if err != nil {
		panic(err)
	}
	err = unix.Setns(mntNS, 0)
	if err != nil {
		panic(err)
	}
	err = unix.Setns(pidNS, 0)
	if err != nil {
		panic(err)
	}
	err = unix.Setns(utsNS, 0)
	if err != nil {
		panic(err)
	}

	openMax, err := sysconf.Sysconf(sysconf.SC_OPEN_MAX)
	if err != nil {
		panic(err)
	}
	for i := int64(0); i < openMax; i++ {
		syscall.CloseOnExec(int(i))
	}

	err = syscall.Chroot(fmt.Sprintf("/var/lib/podkit/container/%d", ContainerID))
	if err != nil {
		panic(err)
	}

	s, err := exec.LookPath(commandPath)
	if err != nil {
		os.Stdout.Write([]byte{1})
		return
	}

	os.Stdout.Write([]byte{0})

	os.Stdin.Close()
	os.Stdout.Close()
	os.Stderr.Close()

	// 开的文件描述符为null
	syscall.Open("/dev/null", os.O_RDWR, 0)
	syscall.Dup(0)
	syscall.Dup(0)

	err = syscall.Exec(s, nil, []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"})
	if err != nil {
		panic(err)
	}
}

// 被设置了path
func execFrontground(commandPath string, ptyNum int) {
	ipcNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/ipc", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	mntNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/mnt", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	pidNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/pid", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	utsNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/uts", ContainerID), syscall.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	err = unix.Setns(ipcNS, 0)
	if err != nil {
		panic(err)
	}
	err = unix.Setns(mntNS, 0)
	if err != nil {
		panic(err)
	}
	err = unix.Setns(pidNS, 0)
	if err != nil {
		panic(err)
	}
	err = unix.Setns(utsNS, 0)
	if err != nil {
		panic(err)
	}

	err = syscall.Chroot(fmt.Sprintf("/var/lib/podkit/container/%d", ContainerID))
	if err != nil {
		panic(err)
	}

	err = os.Chdir("/")
	if err != nil {
		panic(err)
	}

	cmdPath, err := exec.LookPath(commandPath)
	if err != nil {
		_, err := os.Stdout.Write([]byte{1})
		if err != nil {
			panic(err)
		}
		return
	}

	_, err = os.Stdout.Write([]byte{0})
	if err != nil {
		panic(err)
	}

	os.Stdout.Close()
	os.Stderr.Close()

	fd, err := syscall.Dup(0)
	if fd != 1 || err != nil {
		panic(err)
	}

	fd, err = syscall.Dup(0)
	if fd != 2 || err != nil {
		panic(err)
	}

	syscall.Exec(cmdPath, nil, []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"})
}
