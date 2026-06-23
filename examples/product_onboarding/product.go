// Package onboarding is the canonical greenfield worked example for the
// ucloud-cli platform onboarding contract.
//
// It is NOT a real product: it is deliberately placed under examples/ (outside
// products/) so the platform's gen-products/check-product tooling ignores it,
// and it is never registered into the CLI (no products.yaml entry). It exists
// for two reasons:
//
//  1. Documentation. A new product author copies this directory as a starting
//     point. It shows the standard 2-level command shape (<product> <verb>) and
//     exercises every platform API a product is expected to use.
//
//  2. Compile gate. Because it calls EVERY public pkg/cli + pkg/command API,
//     any drift in those signatures breaks `go build ./...`, so CI catches
//     platform-API regressions before they reach real products.
//
// The Run funcs build real SDK requests and type-check against the live SDK,
// but the example is never executed; it only needs to compile.
//
// Shape conventions demonstrated here (the onboarding contract):
//   - Standard verbs only: list, describe, create, delete, start, stop, restart.
//   - A flat 2-level tree: `<product> <verb>` (no 3-level db/conf/backup groups).
//   - Resource id flag named after the product: `--example-id`.
//   - Required flags: MarkFlagRequired + a "Required." description prefix.
//   - Optional flags: an "Optional." description prefix.
//   - Long-running verbs (create/start/stop/restart) offer `--async` and
//     otherwise wait via ctx.Poller(...).Spoll(...).
//   - Destructive verbs (delete) offer `--yes/-y` and gate on ctx.Confirm(...).
package onboarding

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// productName is the single source of truth for the product's command name and
// its resource-id flag (`--<productName>-id`). A real product hard-codes this.
const productName = "example"

// resourceIDFlag is the resource-id flag, named after the product per the
// onboarding contract.
const resourceIDFlag = productName + "-id" // "example-id"

// Product implements cli.Product. The platform calls New() to obtain it, then
// Metadata() to learn who owns it, then NewCommand(ctx) to mount its subtree.
type Product struct{}

// New returns the product instance. The platform's generated registration code
// calls this constructor; here it is exercised only by the example's own tests
// and by NewCommand below.
func New() cli.Product { return &Product{} }

// Metadata identifies the product and its owners. Commands is informational
// (the real tree is built by NewCommand); Version is filled at build time for a
// real product.
func (p *Product) Metadata() cli.Metadata {
	return cli.Metadata{
		Name:     productName,
		Owners:   []string{"platform-onboarding@ucloud.cn"},
		Commands: []string{"list", "describe", "create", "delete", "start", "stop", "restart"},
		Version:  "0.0.0",
	}
}

// NewCommand builds the product's cobra subtree. This is the only wiring a
// product owns: construct the root command and AddCommand one cobra.Command per
// verb. Each verb constructor receives ctx so it can build authed SDK clients
// (via cli.NewServiceClient) and bind common flags.
func (p *Product) NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   productName,
		Short: "Greenfield example product (onboarding worked example)",
		Long: "Greenfield example product demonstrating the ucloud-cli platform " +
			"onboarding contract. Not a real product; exists as the onboarding " +
			"worked example and the platform-API compile gate.",
	}

	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newStart(ctx))
	cmd.AddCommand(newStop(ctx))
	cmd.AddCommand(newRestart(ctx))

	return cmd
}

// Compile-time assurance that Product satisfies the platform interface. If
// cli.Product changes shape, this line (and New's return type) fail to build.
var _ cli.Product = (*Product)(nil)
