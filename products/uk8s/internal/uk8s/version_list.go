package uk8s

import (
	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

const defaultUK8SKind = "Dedicated"

func newVersionList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewGetUK8SVersionsRequest()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List versions supported by UK8S",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.GetUK8SVersions(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]versionRow, 0, len(resp.Data))
			for _, version := range resp.Data {
				rows = append(rows, versionRow{
					K8sVersion:        version.K8sVersion,
					ContainerdVersion: version.ContainerdVersion,
				})
			}
			ctx.PrintList(rows)
		},
	}

	cmd.Flags().SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Kind = cmd.Flags().String("kind", defaultUK8SKind, "Optional. Cluster kind.")
	command.SetFlagValues(cmd, "kind", defaultUK8SKind)
	return cmd
}
