package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestContextForwarders(t *testing.T) {
	var buf bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		Out:        &buf,
		Format:     cli.OutputJSON,
		RegionList: func() []string { return []string{"r"} },
	})

	ctx.PrintList([]struct{ Name string }{{Name: "a"}})
	if !strings.Contains(buf.String(), `"Name"`) {
		t.Fatalf("PrintList(JSON) output missing \"Name\": %q", buf.String())
	}

	if got := ctx.PickResourceID("udb-x/n"); got != "udb-x" {
		t.Fatalf("PickResourceID = %q, want udb-x", got)
	}
}
