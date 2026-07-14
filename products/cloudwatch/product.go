package cloudwatch

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	internalcloudwatch "github.com/ucloud/ucloud-cli/products/cloudwatch/internal/cloudwatch"
)

type Product struct{}

func New() cli.Product { return &Product{} }

func (*Product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "cloudwatch", Commands: []string{"cloudwatch"}}
}

func (*Product) NewCommand(ctx *cli.Context) []*cobra.Command {
	return []*cobra.Command{internalcloudwatch.NewCommand(ctx)}
}

var _ cli.Product = (*Product)(nil)
