package upfs

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalupfs "github.com/ucloud/ucloud-cli/products/upfs/internal/upfs"
)

type product struct{}

// New returns the upfs product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "upfs", Commands: []string{"upfs"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalupfs.NewCommand(ctx)}
}
