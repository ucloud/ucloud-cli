package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListenerUpdate implements `nlb listener update`.
func newListenerUpdate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewUpdateNLBListenerAttributeRequest()

	var nlbID, listenerID string
	var name, remark, scheduler string
	var forwardSrcIPMethod, healthCheckType string
	var healthCheckReqMsg, healthCheckResMsg string
	var startPort, endPort, healthCheckPort int

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an NLB listener",
		Long:  "Update attributes of an NLB listener.",
		Run: func(c *cobra.Command, args []string) {
			flags := c.Flags()
			changedHealthCheckType := flags.Changed("health-check-type")
			changedHealthCheckPort := flags.Changed("health-check-port")
			changedHealthCheckReqMsg := flags.Changed("health-check-req-msg")
			changedHealthCheckResMsg := flags.Changed("health-check-res-msg")
			changed := name != "" || remark != "" || scheduler != "" ||
				flags.Changed("forward-src-ip-method") || flags.Changed("start-port") || flags.Changed("end-port") ||
				changedHealthCheckType || changedHealthCheckPort ||
				changedHealthCheckReqMsg || changedHealthCheckResMsg
			if !changed {
				ctx.HandleError(fmt.Errorf("nothing to update: set at least one of --name/--remark/--scheduler/--forward-src-ip-method/--start-port/--end-port/--health-check-type/--health-check-port/--health-check-req-msg/--health-check-res-msg"))
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			req.ListenerId = sdk.String(ctx.PickResourceID(listenerID))
			if name != "" {
				req.Name = &name
			}
			if remark != "" {
				req.Remark = &remark
			}
			if scheduler != "" {
				req.Scheduler = &scheduler
			}
			if flags.Changed("forward-src-ip-method") {
				req.ForwardSrcIPMethod = &forwardSrcIPMethod
			}
			if flags.Changed("start-port") {
				req.StartPort = sdk.Int(startPort)
			}
			if flags.Changed("end-port") {
				req.EndPort = sdk.Int(endPort)
			}
			if changedHealthCheckType || changedHealthCheckPort || changedHealthCheckReqMsg || changedHealthCheckResMsg {
				hc := &nlbsdk.UpdateNLBListenerAttributeParamHealthCheckConfig{}
				if changedHealthCheckType {
					hc.Type = &healthCheckType
				}
				if changedHealthCheckPort {
					hc.Port = sdk.Int(healthCheckPort)
				}
				if changedHealthCheckReqMsg {
					hc.ReqMsg = &healthCheckReqMsg
				}
				if changedHealthCheckResMsg {
					hc.ResMsg = &healthCheckResMsg
				}
				req.HealthCheckConfig = hc
			}
			if _, err := client.UpdateNLBListenerAttribute(req); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "nlb-listener[%s] updated\n", *req.ListenerId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: *req.ListenerId, Action: "update-listener", Status: "Updated"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance.")
	flags.StringVar(&listenerID, "listener-id", "", "Required. Resource ID of the listener to update.")
	flags.StringVar(&name, "name", "", "Optional. New listener name.")
	flags.StringVar(&remark, "remark", "", "Optional. New remark.")
	flags.StringVar(&scheduler, "scheduler", "", "Optional. New load balancing algorithm: RoundRobin/SourceHash/LeastConn/WeightLeastConn/WeightRoundRobin.")
	flags.StringVar(&forwardSrcIPMethod, "forward-src-ip-method", "", "Optional. Source IP passthrough method: \"\"/None/Toa/ProxyProto.")
	flags.IntVar(&startPort, "start-port", 1, "Optional. New start port of the listen port range (full-port-range listeners only).")
	flags.IntVar(&endPort, "end-port", 65535, "Optional. New end port of the listen port range (full-port-range listeners only).")
	flags.StringVar(&healthCheckType, "health-check-type", "", "Optional. Health check type: Port/UDP/Ping.")
	flags.IntVar(&healthCheckPort, "health-check-port", 0, "Optional. Health check probe port, [1-65535] (0 allowed for Ping).")
	flags.StringVar(&healthCheckReqMsg, "health-check-req-msg", "", "Optional. UDP health check request string.")
	flags.StringVar(&healthCheckResMsg, "health-check-res-msg", "", "Optional. UDP health check expected response string.")

	command.SetFlagValues(cmd, "scheduler", "RoundRobin", "SourceHash", "LeastConn", "WeightLeastConn", "WeightRoundRobin")
	command.SetFlagValues(cmd, "forward-src-ip-method", "", "None", "Toa", "ProxyProto")
	command.SetFlagValues(cmd, "health-check-type", "Port", "UDP", "Ping")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})
	command.SetCompletion(cmd, "listener-id", func() []string {
		return getAllListenerIDNames(ctx, nlbID, derefStr(req.ProjectId), derefStr(req.Region))
	})

	cmd.MarkFlagRequired(resourceIDFlag)
	cmd.MarkFlagRequired("listener-id")

	return cmd
}
