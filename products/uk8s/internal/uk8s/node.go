package uk8s

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newNode(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{Use: "node", Short: "Manage UK8S nodes"}
	cmd.AddCommand(newNodeAdd(ctx))
	cmd.AddCommand(newNodeDelete(ctx))
	cmd.AddCommand(newNodeList(ctx))
	cmd.AddCommand(newNodeDescribe(ctx))
	return cmd
}
