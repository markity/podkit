package main

import (
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

func main() {
	unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(1), 0, 0, 0)

	for {
		var s syscall.WaitStatus
		var r syscall.Rusage
		_, err := syscall.Wait4(-1, &s, 0, &r)
		if err == syscall.ECHILD {
			time.Sleep(time.Second)
			continue
		}
		if err != nil {
			panic(err)
		}
	}
}
