package pathx

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalpathx "github.com/ucloud/ucloud-cli/products/pathx/internal/pathx"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "pathx", Commands: []string{"pathx"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalpathx.NewCommand(ctx)}
}
