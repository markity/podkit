package tools

import (
	"os"
	"runtime"
	"sync"

	"golang.org/x/sys/unix"
)

func MustDoInNS(nsPath string, f func()) {
	w := sync.WaitGroup{}
	w.Add(1)
	go func() {
		runtime.LockOSThread()
		oldNSFile, err := os.OpenFile("/proc/self/ns/net", os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		defer oldNSFile.Close()
		nsFile, err := os.OpenFile(nsPath, os.O_RDONLY, 0)
		if err != nil {
			panic(err)
		}
		defer nsFile.Close()
		err = unix.Setns(int(nsFile.Fd()), 0)
		if err != nil {
			panic(err)
		}
		f()

		err = unix.Setns(int(oldNSFile.Fd()), 0)
		if err != nil {
			panic(err)
		}
		runtime.UnlockOSThread()
		w.Done()
	}()

	w.Wait()
}
