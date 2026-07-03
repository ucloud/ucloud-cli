package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/ui"
)

func TestConfirmYesShortCircuits(t *testing.T) {
	ok, err := ui.Confirm(strings.NewReader(""), &bytes.Buffer{}, true, false, "delete?")
	if err != nil || !ok {
		t.Fatalf("yes=true must be (true,nil), got (%v,%v)", ok, err)
	}
}

func TestConfirmNonInteractiveNoYesErrors(t *testing.T) {
	ok, err := ui.Confirm(strings.NewReader(""), &bytes.Buffer{}, false, false, "delete?")
	if err == nil {
		t.Fatal("non-interactive without --yes must return an error")
	}
	if ok {
		t.Fatal("non-interactive must not confirm")
	}
}

func TestConfirmInteractiveYes(t *testing.T) {
	ok, err := ui.Confirm(strings.NewReader("y\n"), &bytes.Buffer{}, false, true, "delete?")
	if err != nil || !ok {
		t.Fatalf(`interactive "y" must be (true,nil), got (%v,%v)`, ok, err)
	}
}

func TestConfirmInteractiveNo(t *testing.T) {
	ok, err := ui.Confirm(strings.NewReader("n\n"), &bytes.Buffer{}, false, true, "delete?")
	if err != nil || ok {
		t.Fatalf(`interactive "n" must be (false,nil), got (%v,%v)`, ok, err)
	}
}
