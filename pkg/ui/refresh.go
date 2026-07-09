package ui

import (
	"fmt"
	"io"

	"github.com/ucloud/ucloud-cli/ansi"
)

// Refresh rewrites a single progress line on each Do call.
type Refresh struct {
	out   io.Writer
	reset bool
}

func (r *Refresh) Do(text string) {
	if r.reset {
		fmt.Fprint(r.out, ansi.CursorLeft+ansi.CursorUp(1)+ansi.EraseDown)
	} else {
		r.reset = true
	}
	fmt.Fprintln(r.out, text)
}

func NewRefresh(out io.Writer) *Refresh {
	return &Refresh{out: out}
}
