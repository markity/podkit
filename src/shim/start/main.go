package main

import (
	"errors"
	"os"
	"strconv"
)

// podkit_shim

// podkit_shim start stagen 容器id(int类型) ip地址
// podkit_shim exec back 容器id 命令
// podkit_shim exec front 容器id slave_num 命令
func main() {
	// 做一个简单的分辨, 防止用户误用这个命令, 有的用户很好奇啥都想执行一下
	if len(os.Args) < 4 || (os.Args[1] != "start" && os.Args[1] != "exec") || os.Getegid() != 0 {
		return
	}

	mode := os.Args[1]

	switch mode {
	case "start":
		stage := os.Args[2]
		containerID, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
		ipaddr := os.Args[4]
		switch stage {
		case "stage1":
			startStage1(containerID, ipaddr)
		case "stage2":
			startStage2(containerID, ipaddr)
		case "stage3":
			startStage3(containerID, ipaddr)
		default:
			panic(errors.New("unexpected error"))
		}
	case "exec":
		var background bool
		switch os.Args[2] {
		// podkit exec back container_id cmd
		case "back":
			background = true
		// podkit exec front container_id pty_slave_num cmd
		case "front":
			background = false
		default:
			panic(errors.New("unexpected error"))
		}

		containerID, err := strconv.Atoi(os.Args[3])
		if err != nil {
			panic(errors.New("unexpected error"))
		}

		// podkit_shim exec back container_id
		if background {
			cmdPath := os.Args[4]
			execBackground(containerID, cmdPath)
		} else {
			slaveNumString := os.Args[4]
			slaveNum, err := strconv.Atoi(slaveNumString)
			if err != nil {
				panic(errors.New("unexpected error"))
			}
			cmdPath := os.Args[5]
			execFrontground(containerID, cmdPath, slaveNum)
		}
	}

}
