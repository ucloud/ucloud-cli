package uhadoop

import (
	"fmt"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newRestartService(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewRestartUHadoopServiceRequest()
	var nodeIds []string
	var nodeRoles []string
	cmd := &cobra.Command{
		Use:   "restart-service",
		Short: "Restart/start/stop a UHadoop cluster service",
		Long:  `Restart, start, or stop a service on a UHadoop cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			action := "restart"
			if req.OnlyStart != nil && *req.OnlyStart {
				action = "start"
			}
			if req.OnlyStop != nil && *req.OnlyStop {
				action = "stop"
			}
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to %s service %s on cluster %s?", action, *req.ServiceName, *req.InstanceId))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			req.NodeId = nodeIds
			req.NodeRole = nodeRoles
			resp, err := client.RestartUHadoopService(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(w, "uhadoop[%s] service %s %s, state: %s\n", *req.InstanceId, *req.ServiceName, action, resp.State)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.InstanceId, Action: action, Status: resp.State})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.InstanceId = flags.String("instance-id", "", "Required. Cluster instance ID")
	req.ServiceName = flags.String("service-name", "", "Required. Service name")
	req.ApplicationVersion = flags.String("application-version", "", "Optional. Application version")
	req.OnlyStart = flags.Bool("only-start", false, "Optional. Only start the service")
	req.OnlyStop = flags.Bool("only-stop", false, "Optional. Only stop the service")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")
	flags.StringSliceVar(&nodeIds, "node-id", nil, "Optional. Node IDs")
	flags.StringSliceVar(&nodeRoles, "node-role", nil, "Optional. Node roles: master|core|task")

	command.SetFlagValues(cmd, "only-start", "true", "false")
	command.SetFlagValues(cmd, "only-stop", "true", "false")
	command.SetFlagValues(cmd, "node-role", "master", "core", "task")
	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("service-name")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")

	return cmd
}
