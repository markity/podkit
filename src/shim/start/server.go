package main

import (
	"C"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	commpacket "podkit/comm_packet"
	"podkit/frontend/tools"
	"runtime"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)
import (
	"unsafe"

	"github.com/creack/pty"
)

type interactiveRunningContext struct {
	PtyMaster                 *os.File
	PtySlave                  *os.File
	NotifyWhenContainerClosed chan struct{}
	NotifyWhenCommandExited   chan struct{}

	ConnClosedNotify chan struct{}
}

func RunServer(sendWhenListenFinished chan struct{}, sendWhenListenClosed chan struct{}) {
	runtime.LockOSThread()

	// 下面监听网络连接, 操作容器
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: fmt.Sprintf("/var/lib/podkit/socket/%d.sock", ContainerID), Net: "unix"})
	if err != nil {
		panic(err)
	}
	listener.SetUnlinkOnClose(true)
	defer listener.Close()

	// 等待开始Listen的时候才能给start命令得以结束
	sendWhenListenFinished <- struct{}{}

	// 标志位, 使程序进入"正在关闭"的状态
	closing := false
	interactiveContext := make(map[int]*interactiveRunningContext)
	mu := sync.Mutex{} // 保护上面的一切数据

	connClosedNotifySentNotify := make(chan struct{})

	// 开启shim-reaper, 负责收割所有执行完毕的侵入容器的子进程
	go func() {
		waitPIDChan := make(chan int)
		go func() {
			for {
				// init进程永远在运行, 因此这里出错是致命的
				wpid, err := syscall.Wait4(-1, nil, 0, nil)
				waitPIDChan <- wpid
				if err != nil {
					panic(err)
				}
			}
		}()

		for {
			select {
			case wpid := <-waitPIDChan:
				mu.Lock()
				if ctx, ok := interactiveContext[wpid]; ok {
					ctx.NotifyWhenCommandExited <- struct{}{}
					ctx.PtySlave.Close()
					ctx.PtyMaster.Close()
					<-ctx.ConnClosedNotify
					delete(interactiveContext, wpid)
				}

				if closing {
					for k, v := range interactiveContext {
						syscall.Kill(k, syscall.SIGKILL)
						v.NotifyWhenContainerClosed <- struct{}{}
						<-v.ConnClosedNotify
						v.PtySlave.Close()
						v.PtyMaster.Close()
					}
					mu.Unlock()

					connClosedNotifySentNotify <- struct{}{}
					return
				}
				mu.Unlock()
				// 100 ms的检查时间
			case <-time.After(time.Microsecond * 100):
				mu.Lock()

				if closing {
					for k, v := range interactiveContext {
						syscall.Kill(k, syscall.SIGKILL)
						v.NotifyWhenContainerClosed <- struct{}{}
						<-v.ConnClosedNotify
						v.PtySlave.Close()
						v.PtyMaster.Close()
					}
					mu.Unlock()

					connClosedNotifySentNotify <- struct{}{}
					return
				}
				mu.Unlock()
			}
		}
	}()

	for {
		c, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		packetBytes, err := tools.ReadPacketWith4BytesLengthHeader(c)
		if err != nil {
			panic(err)
		}

		switch packet := commpacket.ServerParsePacket(packetBytes).(type) {
		case *commpacket.PacketClientCloseContainerRequest:
			mu.Lock()
			closing = true
			mu.Unlock()
			<-connClosedNotifySentNotify
			_, err := c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerContainerClosedOK{}).MustMarshalToBytes()))
			if err != nil {
				panic(err)
			}
			goto out
		case *commpacket.PacketClientExecBackgroundRequest:
			pidNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/pid", ContainerID), os.O_RDONLY, 0)
			if err != nil {
				panic(err)
			}

			err = unix.Setns(pidNS, 0)
			if err != nil {
				panic(err)
			}

			pipeReader, pipeWriter := io.Pipe()
			//cmd := exec.Command("podkit_shim", "exec", "back", fmt.Sprintf("%d", ContainerID), packet.Command)
			cmd := exec.Command("podkit_shim_exec_back", fmt.Sprintf("%d", ContainerID), packet.Command)
			cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
			// stdout用来通知是否有这个命令
			cmd.Stdout = pipeWriter
			err = cmd.Start()
			if err != nil {
				panic(err)
			}

			result := make([]byte, 1)
			_, err = io.ReadFull(pipeReader, result)
			if err != nil {
				panic(err)
			}

			if result[0] == 0 {
				c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerExecBackgroundResponse{CommandExists: true}).MustMarshalToBytes()))
			} else {
				c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerExecBackgroundResponse{CommandExists: false}).MustMarshalToBytes()))
			}
			c.Close()
			continue
		case *commpacket.PacketClientExecInteractiveRequest:
			ptyMasterFile, err := os.OpenFile(fmt.Sprintf("/var/lib/podkit/container/%d/dev/pts/ptmx", ContainerID), os.O_RDWR, 0)
			if err != nil {
				panic(err)
			}

			ptySlaveFd, err := tools.Ptsnum(ptyMasterFile)
			if err != nil {
				panic(err)
			}

			err = tools.Unlockpt(ptyMasterFile)
			if err != nil {
				panic(err)
			}

			ptySlaveFile, err := os.OpenFile(fmt.Sprintf("/var/lib/podkit/container/%d/dev/pts/%d", ContainerID, ptySlaveFd), os.O_RDWR|syscall.O_NOCTTY, 0)
			if err != nil {
				panic(err)
			}

			pidNS, err := syscall.Open(fmt.Sprintf("/var/lib/podkit/container/%d/proc/1/ns/pid", ContainerID), os.O_RDONLY, 0)
			if err != nil {
				panic(err)
			}

			err = unix.Setns(pidNS, 0)
			if err != nil {
				panic(err)
			}

			pipeReader, pipeWriter := io.Pipe()
			cmd := exec.Command("podkit_shim_exec_front", fmt.Sprint(ContainerID), packet.Command)
			cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
			cmd.Stdin = ptySlaveFile
			cmd.Stdout = pipeWriter
			err = cmd.Start()
			if err != nil {
				panic(err)
			}

			newProcPid := cmd.Process.Pid

			// 这里是向子进程确认是否有这个命令
			result := make([]byte, 1)
			_, err = io.ReadFull(pipeReader, result)
			if err != nil {
				panic(err)
			}

			if result[0] == 1 {
				// 没有这个命令直接返回
				c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerExecInteractiveResponse{CommandExists: false}).MustMarshalToBytes()))
				c.Close()
				ptySlaveFile.Close()
				ptyMasterFile.Close()
				continue
			}

			// 如果有这个命令, 那么开启read, write pty转发
			c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerExecInteractiveResponse{CommandExists: true}).MustMarshalToBytes()))

			ws := pty.Winsize{Rows: uint16(packet.Rows), Cols: uint16(packet.Cols)}

			// 设置winsize
			err = tools.Ioctl(ptyMasterFile.Fd(), pty.TIOCSWINSZ, uintptr(unsafe.Pointer(&ws)))
			if err != nil {
				panic(err)
			}

			notifyWhenConnClosed := make(chan struct{}, 1)
			notifyWhenCommandExited := make(chan struct{}, 1)
			notifyWhenContainerClosed := make(chan struct{}, 1)
			mu.Lock()
			// 这个管道里面包含了巨多东西, 都要读完才能证明所有interactive协程都退出了
			interactiveContext[newProcPid] = &interactiveRunningContext{
				PtyMaster:                 ptyMasterFile,
				PtySlave:                  ptySlaveFile,
				NotifyWhenContainerClosed: notifyWhenContainerClosed,
				NotifyWhenCommandExited:   notifyWhenCommandExited,
				ConnClosedNotify:          notifyWhenConnClosed,
			}

			mu.Unlock()

			go handleInteractiveConn(c, ptyMasterFile, notifyWhenCommandExited, notifyWhenContainerClosed, notifyWhenConnClosed)
		default:
			panic(err)
		}

	}
out:
	sendWhenListenClosed <- struct{}{}
}

func handleInteractiveConn(c net.Conn, ptyMasterFile *os.File, notifyWhenCommandExited chan struct{}, notifyWhenContainerClosed chan struct{}, notifyWhenConnClosed chan struct{}) {
	readFromClientChan := make(chan interface{})
	readFromPtyMaster := make(chan []byte)
	errorChan := make(chan error, 2)

	// conn reader
	go func() {
		for {
			packetBytes, err := tools.ReadPacketWith4BytesLengthHeader(c)
			if err != nil {
				errorChan <- err
				return
			}

			readFromClientChan <- commpacket.ServerParsePacket(packetBytes)
		}
	}()

	// pty reader
	go func() {
		for {
			bs := make([]byte, 512)
			n, err := ptyMasterFile.Read(bs)
			if err != nil {
				errorChan <- err
				return
			}

			readFromPtyMaster <- bs[:n]
		}
	}()

	for {
		select {
		case <-notifyWhenCommandExited:
			c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerNotifyExecInteractiveExited{}).MustMarshalToBytes()))
			c.Close()
			notifyWhenConnClosed <- struct{}{}
			return
		case <-notifyWhenContainerClosed:
			c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerNotifyExecInteractiveContainerClosed{}).MustMarshalToBytes()))
			c.Close()
			notifyWhenConnClosed <- struct{}{}
			return
		case bs := <-readFromPtyMaster:
			c.Write(tools.DoPackWith4Bytes((&commpacket.PacketServerSendPtyOutput{Data: string(bs)}).MustMarshalToBytes()))
		case iface := <-readFromClientChan:
			switch packet := iface.(type) {
			case *commpacket.PacketClientSendPtyInput:
				ptyMasterFile.Write([]byte(iface.(*commpacket.PacketClientSendPtyInput).Data))
			case *commpacket.PacketClientNotifyWinch:
				ws := pty.Winsize{Rows: uint16(packet.Rows), Cols: uint16(packet.Cols)}
				err := tools.Ioctl(ptyMasterFile.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&ws)))
				if err != nil {
					panic(err)
				}
			}
		case err := <-errorChan:
			panic(err)
		}
	}
}
