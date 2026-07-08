package mysql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalmysql "github.com/ucloud/ucloud-cli/products/mysql/internal/mysql"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "mysql", Commands: []string{"mysql"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalmysql.NewCommand(ctx)}
}
