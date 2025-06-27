//go:build windows

package main

import (
	"github.com/Microsoft/go-winio"
	"net"
)

func dialPipe(pipePath string) (net.Conn, error) {
	// because simple things are a pain in Windows
	return winio.DialPipe(pipePath, nil)
}
