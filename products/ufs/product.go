package ufs

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalufs "github.com/ucloud/ucloud-cli/products/ufs/internal/ufs"
)

type product struct{}

// New returns the ufs product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "ufs", Commands: []string{"ufs"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalufs.NewCommand(ctx)}
}
