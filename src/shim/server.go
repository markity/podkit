package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	commpacket "podkit/comm_packet"
	"sync"
	"time"
)

var ContainerCloseMark bool = false
var Mutex sync.Mutex

func handleConn(childPid int, c net.Conn, connID int, sendWhenClosedConn chan int, notifyContainerClosingChan chan struct{}) {
	var ClientSent chan interface{} = make(chan interface{})

	var SendToClient = make(chan struct {
		Data           []byte
		CloseAfterSend bool
	})

	var ErrorChan chan error = make(chan error, 2)

	// reader
	go func() {
		lengthBytes := make([]byte, 4)
		_, err := io.ReadFull(c, lengthBytes)
		if err != nil {
			ErrorChan <- err
			return
		}

		length := binary.BigEndian.Uint32(lengthBytes)
		packet := make([]byte, length)
		_, err = io.ReadFull(c, packet)
		if err != nil {
			ErrorChan <- err
			return
		}

		ClientSent <- commpacket.ServerParsePacket(packet)
	}()

	// writer
	go func() {
		for v := range SendToClient {
			_, err := c.Write(v.Data)
			if err != nil {
				ErrorChan <- err
				return
			}

			if v.CloseAfterSend {
				c.Close()
				sendWhenClosedConn <- connID
				return
			}
		}
	}()

	// 读第一个报文, 看它的请求类型
	request := ""
	cmdPath := ""

	select {
	case firstPacket := <-ClientSent:
		switch p := firstPacket.(type) {
		case *commpacket.ClientExecBackground:
			request = "exec back"
			cmdPath = p.CommandPath
		case *commpacket.ClientExecInteractive:
			request = "exec front"
			cmdPath = cmdPath
		case *commpacket.ClientCloseContainer:
			request = "close container"
			Mutex.Lock()
			ContainerCloseMark = true
			Mutex.Unlock()
			SendToClient <- struct {
				Data           []byte
				CloseAfterSend bool
			}{
				Data:           commpacket.DoPack(4, (&commpacket.ServerNotifyContainerClosed{}).MustMarshalToBytes()),
				CloseAfterSend: true,
			}
			return
		}
	case <-notifyContainerClosingChan:
		SendToClient <- struct {
			Data           []byte
			CloseAfterSend bool
		}{
			Data:           commpacket.DoPack(4, (&commpacket.ServerNotifyContainerClosed{}).MustMarshalToBytes()),
			CloseAfterSend: true,
		}
		return
	}

	if request == "exec front" {

	} else {

	}
}

func RunServer(childPid int, id int, sendWhenListenFinished chan struct{}, sendWhenListenClosed chan struct{}) {
	// 下面监听网络连接, 操作容器
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: fmt.Sprintf("/var/lib/podkit/socket/%d.sock", id), Net: "unix"})
	if err != nil {
		panic(err)
	}
	listener.SetUnlinkOnClose(true)
	defer listener.Close()

	sendWhenListenFinished <- struct{}{}

	var connRemain = 0
	var connIDCounter = 0
	var notifyContainerClosingChanels = make(map[int]chan struct{})

	var ConnClosedNotifyChan = make(chan int)

	for {
		listener.SetDeadline(time.Now().Add(time.Second * 1))
		c, err := listener.Accept()

		if err == nil {
			notifyContainerClosing := make(chan struct{}, 1)
			connRemain++
			notifyContainerClosingChanels[connIDCounter] = notifyContainerClosing
			go handleConn(childPid, c, connIDCounter, ConnClosedNotifyChan, notifyContainerClosing)
			connIDCounter++
		}

		// 及时清理notifyContainerClosingChanels, 避免程序越跑越大
		for {
			select {
			case id := <-ConnClosedNotifyChan:
				connRemain--
				delete(notifyContainerClosingChanels, id)
			default:
				// TODO: FIXME
				goto out
			}
		}
	out:

		// 检查mark
		Mutex.Lock()
		shouldClose := ContainerCloseMark
		Mutex.Unlock()

		if shouldClose {
			if connRemain != 0 {
				for _, v := range notifyContainerClosingChanels {
					// 注意loop variable变量
					v <- struct{}{}
				}

				for range ConnClosedNotifyChan {
					connRemain--
					if connRemain == 0 {
						break
					}
				}
			}
			// TODO: FIXME
			goto out2
		}
	}

out2:

	sendWhenListenClosed <- struct{}{}
}
