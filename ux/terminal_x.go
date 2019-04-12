// +build linux darwin freebsd netbsd openbsd solaris dragonfly

package ux

import (
	"errors"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

var (
	echoLockMutex    sync.Mutex
	origTermStatePtr *unix.Termios
	tty              *os.File
	istty            bool
)

func init() {
	echoLockMutex.Lock()
	defer echoLockMutex.Unlock()

	var err error
	tty, err = os.Open("/dev/tty")
	istty = true
	if err != nil {
		tty = os.Stdin
		istty = false
	}
}

// terminalSize returns width and rows of the terminal.
func terminalSize() (int, int, error) {
	if !istty {
		return 0, 0, errors.New("Not Supported")
	}
	echoLockMutex.Lock()
	defer echoLockMutex.Unlock()

	fd := int(tty.Fd())

	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}

	return int(ws.Col), int(ws.Row), nil
}
