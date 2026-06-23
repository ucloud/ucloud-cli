package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/cmd"
)

const goldenPath = "testdata/cmdtree.golden"

func TestWriteBaseline(t *testing.T) {
	root := cmd.NewCmdRoot()
	cmd.AddChildrenForSnapshot(root)
	got := Render(root)
	// The version string (root's Short: "UCloud CLI vX.Y.Z") is an intended,
	// separately-managed value — not part of the command-tree structure we guard.
	// Normalize it to a stable placeholder so the golden is version-insensitive
	// (base.Version may be a const "0.3.3", "dev", or an ldflags-injected tag).
	if base.Version != "" {
		got = strings.ReplaceAll(got, "v"+base.Version, "v{VERSION}")
	}

	if os.Getenv("WRITE_CMDTREE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("wrote %s (%d bytes)", goldenPath, len(got))
		return
	}

	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v — run WRITE_CMDTREE_GOLDEN=1 go test ./hack/snapshot/ -run TestWriteBaseline to generate", err)
	}
	if got != string(data) {
		t.Fatalf("command tree does not match golden.\nDiff (got vs golden):\ngot:\n%s\nwant:\n%s", got, string(data))
	}
}

const completionGoldenPath = "testdata/completion.golden"

func TestWriteCompletionBaseline(t *testing.T) {
	root := cmd.NewCmdRoot()
	cmd.AddChildrenForSnapshot(root)
	// Nil out the network-backing globals so dynamic completions panic-on-invoke
	// (SetFlagValues closures are immune; SetCompletion closures dereference them).
	//
	// Platform (cmd) dynamic completions dereference base.BizClient directly, so
	// nil-ing it makes them panic. Product (products/udb) dynamic completions go
	// through cli.NewServiceClient → ctor(base.ClientConfig, base.BuildCredential())
	// → SDK call, which ignores base.BizClient; nil-ing base.ClientConfig makes the
	// SDK request build panic instead of issuing a real (non-deterministic, slow)
	// network call. AuthCredential is nil'd alongside for symmetry. This must run
	// AFTER AddChildrenForSnapshot, since some constructors build requests at
	// construction time and need the non-nil stubs.
	base.BizClient = nil
	base.ClientConfig = nil
	base.AuthCredential = nil
	got := RenderCompletion(root)

	if os.Getenv("WRITE_COMPLETION_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(completionGoldenPath), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(completionGoldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("wrote %s (%d bytes)", completionGoldenPath, len(got))
		return
	}

	data, err := os.ReadFile(completionGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v — run WRITE_COMPLETION_GOLDEN=1 go test ./hack/snapshot/ -run TestWriteCompletionBaseline to generate", err)
	}
	if got != string(data) {
		t.Fatalf("completion candidates do not match golden.\ngot:\n%s\nwant:\n%s", got, string(data))
	}
}

func TestRenderStructure(t *testing.T) {
	root := &cobra.Command{Use: "ucloud"}
	sub := &cobra.Command{Use: "demo", Short: "d"}
	sub.Flags().String("name", "def", "Required. name")
	sub.MarkFlagRequired("name")
	root.AddCommand(sub)
	got := Render(root)
	for _, w := range []string{"ucloud demo", "use=demo", "short=d", "flag=name", "default=def", "required=true"} {
		if !strings.Contains(got, w) {
			t.Fatalf("missing %q\n%s", w, got)
		}
	}
}
