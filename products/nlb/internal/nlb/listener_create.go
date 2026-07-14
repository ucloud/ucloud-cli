package nlb

import (
	"fmt"

	"github.com/spf13/cobra"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListenerCreate implements `nlb listener create`.
func newListenerCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewCreateNLBListenerRequest()

	var nlbID string
	var healthCheckPort int
	var healthCheckType, healthCheckReqMsg, healthCheckResMsg string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an NLB listener",
		Long:  "Create a listener on the specified NLB instance.",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.NLBId = sdk.String(ctx.PickResourceID(nlbID))
			hc := &nlbsdk.CreateNLBListenerParamHealthCheckConfig{
				Enabled: sdk.Bool(true),
				Type:    sdk.String(healthCheckType),
				Port:    sdk.Int(healthCheckPort),
			}
			if healthCheckReqMsg != "" {
				hc.ReqMsg = sdk.String(healthCheckReqMsg)
			}
			if healthCheckResMsg != "" {
				hc.ResMsg = sdk.String(healthCheckResMsg)
			}
			req.HealthCheckConfig = hc

			resp, err := client.CreateNLBListener(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "nlb-listener[%s] created\n", resp.ListenerId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.ListenerId, Action: "create-listener", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&nlbID, resourceIDFlag, "", "Required. Resource ID of the NLB instance to add the listener to.")
	req.Protocol = flags.String("protocol", "", "Required. Listen protocol: TCP/UDP.")
	req.Name = flags.String("name", "", "Optional. Listener name, 1-255 chars.")
	req.Scheduler = flags.String("scheduler", "RoundRobin", "Optional. Load balancing algorithm: RoundRobin/SourceHash/LeastConn/WeightLeastConn/WeightRoundRobin.")
	req.StartPort = flags.Int("start-port", 1, "Required. Start port of the listen port range.")
	req.EndPort = flags.Int("end-port", 65535, "Required. End port of the listen port range.")
	req.StickinessTimeout = flags.Int("stickiness-timeout", 0, "Optional. Session stickiness timeout in seconds, [60-900], 0 disables it.")
	req.ForwardSrcIPMethod = flags.String("forward-src-ip-method", "", "Optional. Source IP passthrough method: \"\"/None/Toa/ProxyProto.")
	flags.IntVar(&healthCheckPort, "health-check-port", 0, "Optional. Health check probe port, [1-65535] (0 allowed for Ping).")
	flags.StringVar(&healthCheckType, "health-check-type", "Port", "Optional. Health check method: Port/UDP/Ping.")
	flags.StringVar(&healthCheckReqMsg, "health-check-req-msg", "", "Optional. UDP health check request string.")
	flags.StringVar(&healthCheckResMsg, "health-check-res-msg", "", "Optional. UDP health check expected response string.")

	command.SetFlagValues(cmd, "protocol", "TCP", "UDP")
	command.SetFlagValues(cmd, "scheduler", "RoundRobin", "SourceHash", "LeastConn", "WeightLeastConn", "WeightRoundRobin")
	command.SetFlagValues(cmd, "forward-src-ip-method", "", "None", "Toa", "ProxyProto")
	command.SetFlagValues(cmd, "health-check-type", "Port", "UDP", "Ping")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getAllNLBIDNames(ctx, derefStr(req.ProjectId), derefStr(req.Region))
	})

	cmd.MarkFlagRequired(resourceIDFlag)
	cmd.MarkFlagRequired("protocol")
	cmd.MarkFlagRequired("start-port")
	cmd.MarkFlagRequired("end-port")

	return cmd
}
