package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestSchemaJSON(t *testing.T) {
	root := NewCmdRoot()
	AddChildrenForSnapshot(root)

	out, err := cli.RenderSchemaJSON(root)
	if err != nil {
		t.Fatalf("RenderSchemaJSON error: %v", err)
	}

	// Must be valid JSON.
	if !json.Valid([]byte(out)) {
		t.Fatalf("output is not valid JSON:\n%s", out)
	}

	// Must contain the deep mysql db create path.
	if !strings.Contains(out, "mysql db create") {
		t.Fatalf("output missing 'mysql db create':\n%s", out[:min(len(out), 500)])
	}

	// Must contain at least one known flag from mysql db create.
	if !strings.Contains(out, `"charge-type"`) && !strings.Contains(out, `"vpc-id"`) {
		t.Fatalf("output missing expected flags from mysql db create (charge-type or vpc-id):\n%s", out[:min(len(out), 500)])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
