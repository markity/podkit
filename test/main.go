package main

import "net"

func main() {
	c, err := net.ListenUnix("unix", &net.UnixAddr{Name: "./unix.socket", Net: "unix"})
	if err != nil {
		panic(err)
	}

	for {
		c.Accept()
	}
}
