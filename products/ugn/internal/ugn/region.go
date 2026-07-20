package ugn

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newRegion ucloud ugn region
func newRegion(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "List and manipulate ugn regions",
		Long:  "List and manipulate ugn regions",
	}

	cmd.AddCommand(newRegionList(ctx))

	return cmd
}
