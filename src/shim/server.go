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

func tmp(i interface{}) {

}

func handleConn(c net.Conn, connID int, sendWhenClosedConn chan int, notifyContainerClosingChan chan struct{}) {
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

	var interactive bool
	tmp(interactive)

	// 读第一个报文, 看它的请求类型

	for {
		select {
		case firstPacket := <-ClientSent:
			switch p := firstPacket.(type) {
			case *commpacket.ClientExecBackground:
				tmp(p)
				interactive = false
			case *commpacket.ClientExecInteractive:
				interactive = true
			}
		case <-notifyContainerClosingChan:
			SendToClient <- struct {
				Data           []byte
				CloseAfterSend bool
			}{
				Data:           (&commpacket.ServerNotifyContainerClosed{}).MustMarshalToBytes(),
				CloseAfterSend: true,
			}
			break
		}
	}

}

func RunServer(id int, sendWhenListenFinished chan struct{}, sendWhenListenClosed chan struct{}) {
	// 下面监听网络连接, 操作容器
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: fmt.Sprintf("/var/lib/podkit/socket/%d.sock", id), Net: "unix"})
	if err != nil {
		panic(err)
	}
	listener.SetUnlinkOnClose(true)

	sendWhenListenFinished <- struct{}{}

	var connRemain = 0
	var connIDCounter = 0
	var notifyContainerClosingChanels = make(map[int]chan struct{})

	var ConnClosedNotifyChan = make(chan int)

	for {
		listener.SetDeadline(time.Now().Add(time.Second * 1))
		c, _ := listener.Accept()

		// 及时清理notifyContainerClosingChanels, 避免程序越跑越大
		for {
			select {
			case id := <-ConnClosedNotifyChan:
				connRemain--
				delete(notifyContainerClosingChanels, id)
			default:
				break
			}
		}

		// 检查mark
		Mutex.Lock()
		shouldClose := ContainerCloseMark
		Mutex.Unlock()

		if shouldClose {
			for _, v := range notifyContainerClosingChanels {
				go func() {
					v <- struct{}{}
				}()
			}

			for range ConnClosedNotifyChan {
				connRemain--
				if connRemain == 0 {
					break
				}
			}
		} else {
			notifyContainerClosing := make(chan struct{})
			connRemain++
			go handleConn(c, connIDCounter, ConnClosedNotifyChan, notifyContainerClosing)
			connIDCounter++
		}
	}

	sendWhenListenFinished <- struct{}{}
}
