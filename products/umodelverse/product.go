package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalumodelverse "github.com/ucloud/ucloud-cli/products/umodelverse/internal/umodelverse"
)

type product struct{}

// New returns the umodelverse product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "umodelverse", Commands: []string{"umodelverse"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalumodelverse.NewCommand(ctx)}
}
