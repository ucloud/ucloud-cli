package cli_test

import (
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

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
