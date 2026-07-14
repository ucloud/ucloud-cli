package uhadoop

import (
	"fmt"

	"github.com/spf13/cobra"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestartService ucloud uhadoop restart-service
func newRestartService(ctx *cli.Context) *cobra.Command {
	var yes bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewRestartUHadoopServiceRequest()
	var nodeIds []string
	var nodeRoles []string
	cmd := &cobra.Command{
		Use:          "restart-service",
		Short:        "Restart/start/stop a UHadoop cluster service",
		Long:         `Restart, start, or stop a service on a UHadoop cluster`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if req.InstanceId == nil || *req.InstanceId == "" {
				return fmt.Errorf("--instance-id is required")
			}
			if req.ServiceName == nil || *req.ServiceName == "" {
				return fmt.Errorf("--service-name is required")
			}

			action := "restart"
			if req.OnlyStart != nil && *req.OnlyStart {
				action = "start"
			}
			if req.OnlyStop != nil && *req.OnlyStop {
				action = "stop"
			}

			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to %s service %s on cluster %s?", action, *req.ServiceName, *req.InstanceId))
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			req.NodeId = nodeIds
			req.NodeRole = nodeRoles
			resp, err := client.RestartUHadoopService(req)
			if err != nil {
				return err
			}
			if resp.RetCode != 0 {
				return fmt.Errorf("[%d] %s", resp.RetCode, resp.Message)
			}
			fmt.Fprintf(ctx.Err(), "Service %s %s on cluster %s, state: %s\n", *req.ServiceName, action, *req.InstanceId, resp.State)
			ctx.PrintJSON(resp)
			return nil
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.InstanceId = cmd.Flags().String("instance-id", "", "Required. Cluster instance ID")
	req.ServiceName = cmd.Flags().String("service-name", "", "Required. Service name (e.g. Hive, Spark, Hdfs, Yarn)")
	req.ApplicationVersion = cmd.Flags().String("application-version", "", "Optional. Application version, if set, operates on all services of the app")
	req.OnlyStart = cmd.Flags().Bool("only-start", false, "Optional. Only start the service")
	req.OnlyStop = cmd.Flags().Bool("only-stop", false, "Optional. Only stop the service")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")
	cmd.Flags().StringSliceVar(&nodeIds, "node-id", nil, "Optional. Node IDs to filter, can be specified multiple times")
	cmd.Flags().StringSliceVar(&nodeRoles, "node-role", nil, "Optional. Node roles to filter: master|core|task, can be specified multiple times")

	command.SetFlagValues(cmd, "only-start", "true", "false")
	command.SetFlagValues(cmd, "only-stop", "true", "false")
	command.SetFlagValues(cmd, "node-role", "master", "core", "task")

	return cmd
}
