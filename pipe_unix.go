//go:build !windows

package main

import (
	"net"
)

func dialPipe(pipePath string) (net.Conn, error) {
	return net.Dial("unix", pipePath)
}
