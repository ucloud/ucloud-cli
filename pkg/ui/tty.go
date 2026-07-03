package ui

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IsTTY reports whether w is a terminal.
// Returns false for any writer that is not an *os.File backed by a real TTY.
func IsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	return ok && isatty.IsTerminal(f.Fd())
}

// IsReaderTTY reports whether r is an interactive terminal. Only *os.File can
// be a TTY; anything else (bytes.Buffer in tests, pipes) is non-interactive.
// Mirrors base.IsStdinTTY's Cygwin/mintty handling.
func IsReaderTTY(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
