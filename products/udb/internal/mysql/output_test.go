package mysql

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newTestCtx(format cli.OutputFormat) (*cli.Context, *bytes.Buffer, *bytes.Buffer) {
	var out, errb bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		In:     strings.NewReader(""),
		Out:    &out,
		Err:    &errb,
		Format: format,
	})
	return ctx, &out, &errb
}

// In json/yaml mode emitResult writes the structured rows to stdout and nothing
// to stderr.
func TestEmitResultJSON(t *testing.T) {
	ctx, out, errb := newTestCtx(cli.OutputJSON)
	emitResult(ctx, OpResultRow{ResourceID: "udbha-x", Action: "create", Status: "Initializing"})

	if errb.Len() != 0 {
		t.Fatalf("err buffer should be empty, got %q", errb.String())
	}
	var rows []OpResultRow
	if err := json.Unmarshal(out.Bytes(), &rows); err != nil {
		t.Fatalf("stdout is not a valid json array: %v; raw=%q", err, out.String())
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d: %+v", len(rows), rows)
	}
	got := rows[0]
	if got.ResourceID != "udbha-x" || got.Action != "create" || got.Status != "Initializing" {
		t.Fatalf("unexpected row: %+v", got)
	}
}

// In table mode emitResult is a no-op: the human narration on stdout is the
// result, so no structured table is added.
func TestEmitResultTableIsNoop(t *testing.T) {
	ctx, out, _ := newTestCtx(cli.OutputTable)
	emitResult(ctx, OpResultRow{ResourceID: "udbha-x", Action: "create", Status: "Initializing"})
	if out.Len() != 0 {
		t.Fatalf("table mode emitResult must be a no-op, got %q", out.String())
	}
}

// progressWriter routes human narration to stdout in table mode and to stderr
// in machine (json/yaml) modes.
func TestProgressWriterRouting(t *testing.T) {
	ctxT, _, _ := newTestCtx(cli.OutputTable)
	if progressWriter(ctxT) != ctxT.Out() {
		t.Fatal("table mode: progressWriter should return Out()")
	}

	ctxJ, _, _ := newTestCtx(cli.OutputJSON)
	if progressWriter(ctxJ) != ctxJ.Err() {
		t.Fatal("json mode: progressWriter should return Err()")
	}

	ctxY, _, _ := newTestCtx(cli.OutputYAML)
	if progressWriter(ctxY) != ctxY.Err() {
		t.Fatal("yaml mode: progressWriter should return Err()")
	}
}
