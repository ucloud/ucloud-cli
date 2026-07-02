package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	got := RenderPlatform(root, productSkipSet())
	// The version string (root's Short: "UCloud CLI vX.Y.Z") is an intended,
	// separately-managed value — not part of the command-tree structure we guard.
	// Normalize it to a stable placeholder so the golden is version-insensitive
	// (base.Version may be a const "0.3.3", "dev", or an ldflags-injected tag).
	if base.Version != "" {
		got = strings.ReplaceAll(got, "v"+base.Version, "v{VERSION}")
	}

	compareOrWrite(t, goldenPath, got, "WRITE_CMDTREE_GOLDEN")
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
	got := RenderCompletionPlatform(root, productSkipSet())

	compareOrWrite(t, completionGoldenPath, got, "WRITE_COMPLETION_GOLDEN")
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

// productSkipSet returns the top-level command names claimed by registered
// products — exactly the subtrees the platform golden prunes.
func productSkipSet() map[string]bool {
	skip := map[string]bool{}
	for _, p := range cmd.ProductsForSnapshot() {
		for _, c := range p.Metadata().Commands {
			skip[c] = true
		}
	}
	return skip
}

// renderProduct renders the product-claimed top-level subtrees (sorted by
// command name) from the fully-built root, so CommandPath keeps the
// "ucloud " prefix and lines stay byte-identical to the pre-split golden.
func renderProduct(t *testing.T, root *cobra.Command, commands []string, render func(*cobra.Command) string) string {
	t.Helper()
	names := append([]string(nil), commands...)
	sort.Strings(names)
	var b strings.Builder
	for _, name := range names {
		var target *cobra.Command
		for _, ch := range root.Commands() {
			if ch.Name() == name {
				target = ch
				break
			}
		}
		if target == nil {
			t.Fatalf("product command %q not found under root — product.yaml/Metadata out of sync?", name)
		}
		b.WriteString(render(target))
	}
	return b.String()
}

// compareOrWrite implements the golden write/compare protocol shared by all
// snapshot tests.
func compareOrWrite(t *testing.T, path, got, writeEnv string) {
	t.Helper()
	if os.Getenv(writeEnv) == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("wrote %s (%d bytes)", path, len(got))
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v — run %s=1 go test ./hack/snapshot to generate", err, writeEnv)
	}
	if got != string(data) {
		t.Fatalf("golden mismatch for %s (refresh: %s=1 go test ./hack/snapshot).\ngot:\n%s\nwant:\n%s", path, writeEnv, got, string(data))
	}
}

// lineMultisetDiff compares a and b as line multisets and returns up to 10
// offending lines, each prefixed with its signed residual count (positive:
// surplus in a; negative: surplus in b). An empty result means the multisets
// are equal.
func lineMultisetDiff(a, b string) []string {
	count := map[string]int{}
	for _, l := range strings.Split(a, "\n") {
		count[l]++
	}
	for _, l := range strings.Split(b, "\n") {
		count[l]--
	}
	var diff []string
	for l, n := range count {
		if n != 0 {
			diff = append(diff, fmt.Sprintf("%+d %s", n, l))
		}
	}
	sort.Strings(diff)
	if len(diff) > 10 {
		diff = diff[:10]
	}
	return diff
}

// TestProductGoldens verifies each product's command subtree against the
// golden the product team owns. Refresh one product:
//
//	WRITE_CMDTREE_GOLDEN=1 go test ./hack/snapshot -run 'TestProductGoldens/<name>'
func TestProductGoldens(t *testing.T) {
	root := cmd.NewCmdRoot()
	cmd.AddChildrenForSnapshot(root)
	for _, p := range cmd.ProductsForSnapshot() {
		meta := p.Metadata()
		t.Run(meta.Name, func(t *testing.T) {
			got := renderProduct(t, root, meta.Commands, Render)
			path := filepath.Join("..", "..", "products", meta.Name, "testdata", "cmdtree.golden")
			compareOrWrite(t, path, got, "WRITE_CMDTREE_GOLDEN")
		})
	}
}

// TestProductCompletionGoldens is the completion-candidate counterpart of
// TestProductGoldens. Refresh:
//
//	WRITE_COMPLETION_GOLDEN=1 go test ./hack/snapshot -run 'TestProductCompletionGoldens/<name>'
func TestProductCompletionGoldens(t *testing.T) {
	root := cmd.NewCmdRoot()
	cmd.AddChildrenForSnapshot(root)
	base.BizClient = nil
	base.ClientConfig = nil
	base.AuthCredential = nil
	for _, p := range cmd.ProductsForSnapshot() {
		meta := p.Metadata()
		t.Run(meta.Name, func(t *testing.T) {
			got := renderProduct(t, root, meta.Commands, RenderCompletion)
			path := filepath.Join("..", "..", "products", meta.Name, "testdata", "completion.golden")
			compareOrWrite(t, path, got, "WRITE_COMPLETION_GOLDEN")
		})
	}
}

// TestGoldenPartition guards against silent coverage loss: the full-tree
// render must equal platform render + all product renders as a line multiset.
// A pruning bug that dropped a non-product subtree would fail here — this is
// the permanent replacement for the one-time migration equivalence check.
func TestGoldenPartition(t *testing.T) {
	root := cmd.NewCmdRoot()
	cmd.AddChildrenForSnapshot(root)
	full := Render(root)
	parts := RenderPlatform(root, productSkipSet())
	for _, p := range cmd.ProductsForSnapshot() {
		parts += renderProduct(t, root, p.Metadata().Commands, Render)
	}
	if d := lineMultisetDiff(full, parts); len(d) > 0 {
		t.Fatalf("golden partition lost or duplicated lines: full render != platform + products: %v", d)
	}

	root2 := cmd.NewCmdRoot()
	cmd.AddChildrenForSnapshot(root2)
	base.BizClient = nil
	base.ClientConfig = nil
	base.AuthCredential = nil
	fullC := RenderCompletion(root2)
	partsC := RenderCompletionPlatform(root2, productSkipSet())
	for _, p := range cmd.ProductsForSnapshot() {
		partsC += renderProduct(t, root2, p.Metadata().Commands, RenderCompletion)
	}
	if d := lineMultisetDiff(fullC, partsC); len(d) > 0 {
		t.Fatalf("completion partition lost or duplicated lines: %v", d)
	}
}
