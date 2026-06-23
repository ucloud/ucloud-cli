package mysql

import (
	"io"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// progressWriter returns the writer for human-facing narration (the
// "udb[...] is ..." lines and the poller spinner).
//
//   - Table mode: stdout, so the interactive/human experience is byte-for-byte
//     unchanged from the legacy behaviour.
//   - JSON/YAML mode: stderr, so stdout carries only the structured result and
//     stays machine-parseable (matches gcloud: progress on stderr, result on
//     stdout).
func progressWriter(ctx *cli.Context) io.Writer {
	if ctx.Format() == cli.OutputTable {
		return ctx.Out()
	}
	return ctx.Err()
}

// emitResult prints the structured operation result rows to stdout, but only in
// machine (json/yaml) modes. In table mode it is a no-op: the human narration
// already written to stdout is the result, so no extra table is added.
func emitResult(ctx *cli.Context, rows ...OpResultRow) {
	if ctx.Format() == cli.OutputTable {
		return
	}
	ctx.PrintList(rows)
}
