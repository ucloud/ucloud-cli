package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestEmitResultJSONWritesStructuredRowToStdout(t *testing.T) {
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Format: cli.OutputJSON})

	ctx.EmitResult(cli.OpResultRow{ResourceID: "eip-abc", Action: "allocate", Status: "Available"})

	s := out.String()
	for _, want := range []string{"eip-abc", "allocate", "Available", `"ResourceID"`, `"Action"`, `"Status"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("EmitResult(JSON) missing %q in %q", want, s)
		}
	}
}

func TestEmitResultTableIsNoOp(t *testing.T) {
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Format: cli.OutputTable})

	ctx.EmitResult(cli.OpResultRow{ResourceID: "eip-abc", Action: "allocate", Status: "Available"})

	if out.Len() != 0 {
		t.Fatalf("EmitResult(table) must be a no-op, got %q", out.String())
	}
}

func TestProgressWriterRoutesByFormat(t *testing.T) {
	var out, err bytes.Buffer

	table := cli.NewContext(cli.Deps{Out: &out, Err: &err, Format: cli.OutputTable})
	if table.ProgressWriter() != table.Out() {
		t.Fatal("table mode: ProgressWriter must route to Out (stdout)")
	}

	js := cli.NewContext(cli.Deps{Out: &out, Err: &err, Format: cli.OutputJSON})
	if js.ProgressWriter() != js.Err() {
		t.Fatal("json mode: ProgressWriter must route to Err (stderr)")
	}

	yml := cli.NewContext(cli.Deps{Out: &out, Err: &err, Format: cli.OutputYAML})
	if yml.ProgressWriter() != yml.Err() {
		t.Fatal("yaml mode: ProgressWriter must route to Err (stderr)")
	}
}

func TestEmitResultJSONEmptyRowsIsEmptyArrayNotNull(t *testing.T) {
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Format: cli.OutputJSON})

	var rows []cli.OpResultRow // nil slice — the all-failed case
	ctx.EmitResult(rows...)

	if got := out.String(); got != "[]\n" {
		t.Fatalf("EmitResult(JSON) empty must be %q, got %q", "[]\n", got)
	}
}

func TestEmitResultYAMLEmptyRowsIsEmptyArray(t *testing.T) {
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Format: cli.OutputYAML})

	var rows []cli.OpResultRow
	ctx.EmitResult(rows...)

	if got := out.String(); got != "[]\n" {
		t.Fatalf("EmitResult(YAML) empty must be %q, got %q", "[]\n", got)
	}
}
