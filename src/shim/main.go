package main

import (
	"os"
	"strconv"
)

// podkit_shim

// podkit_shim start stagen 容器id(int类型)
// podkit_shim exec back/front 容器id 命令 uts_ns mnt_ns
func main() {
	if len(os.Args) < 2 {
		return
	}
	if os.Geteuid() != 0 {
		return
	}

	mode := os.Args[1]

	switch mode {
	case "start":
		if len(os.Args) != 4 {
			return
		}
		stage := os.Args[2]
		id, err := strconv.Atoi(os.Args[3])
		if err != nil {
			return
		}
		switch stage {
		case "stage1":
			startStage1(id)
		case "stage2":
			startStage2(id)
		case "stage3":
			startStage3(id)
		default:
			return
		}
	case "exec":
	}

}
