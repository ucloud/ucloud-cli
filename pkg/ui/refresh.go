package ui

import (
	"fmt"
	"io"
)

// Refresh rewrites a single progress line on each Do call.
type Refresh struct {
	out   io.Writer
	reset bool
}

func (r *Refresh) Do(text string) {
	if r.reset {
		fmt.Fprint(r.out, ansiCursorLeft+ansiCursorUp(1)+ansiEraseDown)
	} else {
		r.reset = true
	}
	fmt.Fprintln(r.out, text)
}

func NewRefresh(out io.Writer) *Refresh {
	return &Refresh{out: out}
}
