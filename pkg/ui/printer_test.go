package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/ui"
)

type row struct{ Name, Status string }

func TestPrinterFormats(t *testing.T) {
	for _, c := range []struct {
		f    ui.Format
		want string
	}{
		{ui.Table, "Name"}, {ui.JSON, `"Name"`}, {ui.YAML, "name:"},
	} {
		var b bytes.Buffer
		ui.Printer{Out: &b, Format: c.f}.PrintList([]row{{"mydb", "Running"}})
		if !strings.Contains(b.String(), c.want) {
			t.Fatalf("fmt %v missing %q: %s", c.f, c.want, b.String())
		}
	}
}

func TestConfirm(t *testing.T) {
	var buf bytes.Buffer

	// yes=true should always return true without reading input
	if !ui.Confirm(nil, &buf, true, "x") {
		t.Fatal("Confirm with yes=true should return true")
	}

	// "y" input should return true
	if !ui.Confirm(strings.NewReader("y\n"), &buf, false, "ok?") {
		t.Fatal("Confirm with 'y' input should return true")
	}

	// "yes" input should return true
	buf.Reset()
	if !ui.Confirm(strings.NewReader("yes\n"), &buf, false, "ok?") {
		t.Fatal("Confirm with 'yes' input should return true")
	}

	// "n" input should return false
	buf.Reset()
	if ui.Confirm(strings.NewReader("n\n"), &buf, false, "ok?") {
		t.Fatal("Confirm with 'n' input should return false")
	}
}
