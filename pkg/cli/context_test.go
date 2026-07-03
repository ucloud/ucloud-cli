package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestFailedFalseUntilHandleError(t *testing.T) {
	var out, errw bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Err: &errw, Format: cli.OutputJSON})

	if ctx.Failed() {
		t.Fatal("fresh context must not be Failed()")
	}
	ctx.HandleError(fmt.Errorf("boom"))
	if !ctx.Failed() {
		t.Fatal("after HandleError, Failed() must be true")
	}
}

func TestPickResourceID(t *testing.T) {
	if cli.PickResourceID("udb-x/n") != "udb-x" {
		t.Fatal("bad")
	}
}

func TestOutputFormatDefault(t *testing.T) {
	// OutputTable must be the iota-zero value
	var f cli.OutputFormat
	if f != cli.OutputTable {
		t.Fatal("zero-value OutputFormat should be OutputTable")
	}

	// NewContext with no Format set should report OutputTable via Format()
	ctx := cli.NewContext(cli.Deps{})
	if ctx.Format() != cli.OutputTable {
		t.Fatal("NewContext with zero Deps should have Format() == OutputTable")
	}
}
