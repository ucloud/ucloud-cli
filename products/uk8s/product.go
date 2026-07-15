package uk8s

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internaluk8s "github.com/ucloud/ucloud-cli/products/uk8s/internal/uk8s"
)

type product struct{}

// New returns the uk8s product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uk8s", Commands: []string{"uk8s"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internaluk8s.NewCommand(ctx)}
}