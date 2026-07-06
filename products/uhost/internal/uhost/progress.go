package uhost

import (
	"fmt"
	"sync"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// failCounter is a concurrency-safe tally of failed create/delete operations, so
// RunE can return a non-zero exit when any item fails (aws/gcloud convention: a
// failed command exits non-zero, not 0).
type failCounter struct {
	mu sync.Mutex
	n  int
}

func (f *failCounter) inc() {
	f.mu.Lock()
	f.n++
	f.mu.Unlock()
}

func (f *failCounter) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.n
}

// resultCollector is a concurrency-safe accumulator of structured operation
// rows, so uhost create/delete (which narrate via the progress block, not
// PrintList) can still emit machine-readable results in --output json/yaml mode
// like the other write commands.
type resultCollector struct {
	mu   sync.Mutex
	rows []cli.OpResultRow
}

func (rc *resultCollector) add(rows ...cli.OpResultRow) {
	rc.mu.Lock()
	rc.rows = append(rc.rows, rows...)
	rc.mu.Unlock()
}

func (rc *resultCollector) all() []cli.OpResultRow {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.rows
}

// reportFail records a failure message: it appends to the progress block (shown
// on a TTY) and, when the block is NOT being animated (non-TTY writer, or the
// aggregate count>5 path), also writes the message to stderr so scripted/piped
// callers still see the error. Mirrors the aws/gcloud convention that command
// errors always reach stderr regardless of whether stdout is a terminal, while
// the spinner stays TTY-only.
func reportFail(ctx *cli.Context, prog *cli.Progress, block *cli.Block, msg string) {
	block.Append(msg)
	if !prog.Animated() {
		fmt.Fprintln(ctx.Err(), msg)
	}
}
