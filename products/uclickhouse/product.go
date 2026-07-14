package uclickhouse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalclickhouse "github.com/ucloud/ucloud-cli/products/uclickhouse/internal/clickhouse"
)

type product struct{}

// New returns the uclickhouse product (registered via hack/gen-products).
func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uclickhouse", Commands: []string{"uclickhouse"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalclickhouse.NewCommand(ctx)}
}
