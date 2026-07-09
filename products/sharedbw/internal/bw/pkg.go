package bw

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPkg returns ucloud bw pkg.
func newPkg(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: "List, create and delete bandwidth package instances",
		Long:  "List, create and delete bandwidth package instances",
	}
	cmd.AddCommand(newPkgCreate(ctx))
	cmd.AddCommand(newPkgList(ctx))
	cmd.AddCommand(newPkgDelete(ctx))
	return cmd
}
