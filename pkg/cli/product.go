package cli

import "github.com/spf13/cobra"

// Metadata identifies a product and its owners.
// Commands declares the top-level command names this product claims. It is the
// basis for golden partitioning: the platform golden (hack/snapshot/testdata)
// prunes exactly these subtrees, and the product's own goldens
// (products/<name>/testdata) cover them. It must match product.yaml (rule-8).
// The actual cobra command trees are built by NewCommand.
type Metadata struct {
	Name     string
	Owners   []string
	Commands []string
	Version  string
}

// Product is a self-contained product module the platform registers.
// NewCommand builds all top-level cobra command subtrees this product owns.
type Product interface {
	Metadata() Metadata
	NewCommand(ctx *Context) []*cobra.Command
}
