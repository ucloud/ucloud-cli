package redis

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalredis "github.com/ucloud/ucloud-cli/products/redis/internal/redis"
)

type product struct{}

func New() cli.Product { return product{} }

func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "redis", Commands: []string{"redis"}}
}

func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalredis.NewCommand(ctx)}
}
