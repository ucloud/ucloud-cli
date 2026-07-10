package uk8s

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newNodeGroup(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{Use: "nodegroup", Short: "Manage UK8S node groups"}
	cmd.AddCommand(newNodeGroupAdd(ctx))
	cmd.AddCommand(newNodeGroupDelete(ctx))
	cmd.AddCommand(newNodeGroupList(ctx))
	return cmd
}
