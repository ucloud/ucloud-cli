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
