package tools

import (
	"os"
	"syscall"
	"unsafe"
)

func Ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}

func Ptsnum(f *os.File) (uint32, error) {
	var n uint32

	err := Ioctl(f.Fd(), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&n))) //nolint:gosec // Expected unsafe pointer for Syscall call.
	if err != nil {
		return 0, err
	}
	return n, nil
}

func Unlockpt(f *os.File) error {
	var u int32
	// use TIOCSPTLCK with a zero valued arg to clear the slave pty lock
	return Ioctl(f.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&u)))

}
