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
	ok, err := ui.Confirm(nil, &buf, true, false, "x")
	if err != nil || !ok {
		t.Fatalf("Confirm with yes=true should return (true,nil), got (%v,%v)", ok, err)
	}

	// "y" input should return true
	ok, err = ui.Confirm(strings.NewReader("y\n"), &buf, false, true, "ok?")
	if err != nil || !ok {
		t.Fatalf("Confirm with 'y' input should return (true,nil), got (%v,%v)", ok, err)
	}

	// "yes" input should return true
	buf.Reset()
	ok, err = ui.Confirm(strings.NewReader("yes\n"), &buf, false, true, "ok?")
	if err != nil || !ok {
		t.Fatalf("Confirm with 'yes' input should return (true,nil), got (%v,%v)", ok, err)
	}

	// "n" input should return false
	buf.Reset()
	ok, err = ui.Confirm(strings.NewReader("n\n"), &buf, false, true, "ok?")
	if err != nil || ok {
		t.Fatalf("Confirm with 'n' input should return (false,nil), got (%v,%v)", ok, err)
	}
}

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := ui.PrintJSON(map[string]int{"a": 1}, &buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\"a\": 1") {
		t.Fatalf("unexpected: %q", buf.String())
	}
}
