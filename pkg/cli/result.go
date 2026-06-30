package cli

import "io"

// OpResultRow is the platform-standard structured result of a write command
// (create/delete/start/stop/resize/...). It is emitted only in machine
// (json/yaml) modes via EmitResult; in table mode the human narration on stdout
// is the result. Field names are the JSON keys (no json tags), matching the
// CLI's existing row convention. Products use this instead of each defining
// their own OpResultRow (see batch-1 plan D-A; promoted from products/udb).
type OpResultRow struct {
	ResourceID string
	Action     string
	Status     string
}

// ProgressWriter returns the writer for human-facing narration and progress:
//
//   - Table mode: stdout, so the interactive experience is unchanged.
//   - JSON/YAML mode: stderr, so stdout carries only the structured result and
//     stays machine-parseable (gcloud convention: progress on stderr, result on
//     stdout).
func (c *Context) ProgressWriter() io.Writer {
	if c.format == OutputTable {
		return c.out
	}
	return c.err
}

// EmitResult prints structured operation-result rows to stdout, but only in
// machine (json/yaml) modes. In table mode it is a no-op: the human narration
// already written to stdout is the result, so no extra table is added.
func (c *Context) EmitResult(rows ...OpResultRow) {
	if c.format == OutputTable {
		return
	}
	c.PrintList(rows)
}
