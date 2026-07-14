package ukafka

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand builds the `ukafka` root command and mounts the subcommands.
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ukafka",
		Short: "Manage UKafka instances",
		Long:  "Manage UKafka instances",
	}
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newDescribe(ctx))
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newNodeConf(ctx))
	cmd.AddCommand(newAppVersion(ctx))
	cmd.AddCommand(newAddNode(ctx))
	cmd.AddCommand(newDescribeConsumer(ctx))
	cmd.AddCommand(newCheckTopic(ctx))
	cmd.AddCommand(newListConsumers(ctx))
	cmd.AddCommand(newListTopics(ctx))
	cmd.AddCommand(newModifyType(ctx))
	cmd.AddCommand(newResizeDisk(ctx))
	return cmd
}
