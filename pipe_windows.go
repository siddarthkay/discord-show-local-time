//go:build windows

package main

import (
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

// pipeConn implements net.Conn for Windows named pipes
type pipeConn struct {
	handle windows.Handle
	addr   pipeAddr

	readDeadline  pipeDeadline
	writeDeadline pipeDeadline
}

type pipeAddr struct {
	path string
}

func (p pipeAddr) Network() string { return "pipe" }
func (p pipeAddr) String() string  { return p.path }

type pipeDeadline struct {
	mu       sync.Mutex
	timer    *time.Timer
	deadline time.Time
}

func dialPipe(pipePath string) (net.Conn, error) {
	// Convert Go string to UTF-16 for Windows API
	pathPtr, err := windows.UTF16PtrFromString(pipePath)
	if err != nil {
		return nil, err
	}

	// Open the named pipe with synchronous I/O
	// Discord IPC is write-only (fire-and-forget), so synchronous I/O is sufficient
	handle, err := windows.CreateFile(
		pathPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return nil, &net.OpError{
			Op:   "dial",
			Net:  "pipe",
			Addr: pipeAddr{pipePath},
			Err:  err,
		}
	}

	return &pipeConn{
		handle: handle,
		addr:   pipeAddr{pipePath},
	}, nil
}

func (c *pipeConn) Read(b []byte) (int, error) {
	var n uint32
	err := windows.ReadFile(c.handle, b, &n, nil)
	if err != nil {
		if err == windows.ERROR_BROKEN_PIPE {
			return int(n), io.EOF
		}
		return int(n), err
	}
	return int(n), nil
}

func (c *pipeConn) Write(b []byte) (int, error) {
	var n uint32
	err := windows.WriteFile(c.handle, b, &n, nil)
	if err != nil {
		return int(n), err
	}
	return int(n), nil
}

func (c *pipeConn) Close() error {
	return windows.CloseHandle(c.handle)
}

func (c *pipeConn) LocalAddr() net.Addr {
	return c.addr
}

func (c *pipeConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *pipeConn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

func (c *pipeConn) SetReadDeadline(t time.Time) error {
	c.readDeadline.set(t)
	return nil
}

func (c *pipeConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline.set(t)
	return nil
}

func (d *pipeDeadline) set(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.deadline = t
}
