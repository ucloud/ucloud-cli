package css

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `css` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "css",
		Short: "Manage UES (Elasticsearch/OpenSearch) instances",
		Long:  "Manage UES (Elasticsearch/OpenSearch) instances",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newExpand(ctx))
	cmd.AddCommand(newResize(ctx))
	cmd.AddCommand(newRestart(ctx))
	cmd.AddCommand(newDiskLimit(ctx))
	cmd.AddCommand(newNodeConf(ctx))
	cmd.AddCommand(newAppVersion(ctx))
	return cmd
}
