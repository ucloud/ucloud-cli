package cli

import "github.com/spf13/cobra"

// Metadata identifies a product and its owners.
// Commands is a slice of command names this product exposes
// (informational only; the actual cobra command tree is built by NewCommand).
type Metadata struct {
	Name     string
	Owners   []string
	Commands []string
	Version  string
}

// Product is a self-contained product module the platform registers.
// NewCommand builds the product's cobra command subtree given a Context.
type Product interface {
	Metadata() Metadata
	NewCommand(ctx *Context) *cobra.Command
}
