package pathx

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUGA builds `ucloud pathx uga`.
func newUGA(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uga",
		Short: "Create,list,update and delete pathx uga instances",
		Long:  "Create,list,update and delete pathx uga instances",
	}
	cmd.AddCommand(newUGAList(ctx))
	cmd.AddCommand(newUGADescribe(ctx))
	cmd.AddCommand(newUGACreate(ctx))
	cmd.AddCommand(newUGADelete(ctx))
	cmd.AddCommand(newUGAAddPort(ctx))
	cmd.AddCommand(newUGARemovePort(ctx))
	return cmd
}
