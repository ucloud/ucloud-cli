package uk8s

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	uk8ssdk "github.com/ucloud/ucloud-sdk-go/services/uk8s"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newGetConfig(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uk8ssdk.NewClient)
	req := client.NewGetClusterConfigRequest()
	var external bool

	cmd := &cobra.Command{
		Use:   "get-config",
		Short: "Print a UK8S cluster kubeconfig",
		Long:  "Print the internal kubeconfig, or the external kubeconfig with --external.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			*req.ClusterId = ctx.PickResourceID(*req.ClusterId)
			resp, err := client.GetClusterConfig(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			config := resp.KubeConfig
			if external {
				config = resp.ExternalKubeConfig
			}
			if strings.TrimSpace(config) == "" {
				kind := "internal"
				if external {
					kind = "external"
				}
				ctx.HandleError(fmt.Errorf("%s kubeconfig is not available for cluster %q", kind, *req.ClusterId))
				return
			}
			fmt.Fprint(ctx.Out(), config)
			if !strings.HasSuffix(config, "\n") {
				fmt.Fprintln(ctx.Out())
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.ClusterId = flags.String("cluster-id", "", "Required. Cluster ID whose kubeconfig will be printed.")
	flags.BoolVar(&external, "external", false, "Optional. Print the external kubeconfig.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	cmd.MarkFlagRequired("cluster-id")
	command.SetCompletion(cmd, "cluster-id", func() []string {
		return listClusterIDs(ctx, []string{CLUSTER_RUNNING}, derefStr(req.Region), derefStr(req.ProjectId))
	})
	return cmd
}
