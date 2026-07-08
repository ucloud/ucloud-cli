package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newVServerCreate returns ucloud ulb vserver create.
func newVServerCreate(ctx *cli.Context) *cobra.Command {
	sslID := ""
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewCreateVServerRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create ULB VServer instance",
		Long:  "Create ULB VServer instance",
		Run: func(c *cobra.Command, args []string) {
			if *req.ListenType == "RequestProxy" && (*req.ClientTimeout <= 0 || *req.ClientTimeout > 86400) {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, client-timeout-seconds in the range of (0,86400]")
				return
			}
			if *req.ListenType == "PacketsTransmit" && (*req.ClientTimeout <= 0 || *req.ClientTimeout > 86400) {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, client-timeout-seconds in the range of [60，900]")
				return
			}
			if *req.Protocol == "HTTPS" && sslID == "" {
				fmt.Fprintln(ctx.ProgressWriter(), "Error, SSL Certificate is needed when you choose HTTPS")
				return
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			resp, err := client.CreateVServer(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ulb-vserver[%s] created\n", resp.VServerId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.VServerId, Action: "create-vserver", Status: "Created"})
			if *req.Protocol == "HTTPS" && sslID != "" {
				bindReq := client.NewBindSSLRequest()
				bindReq.Region = req.Region
				bindReq.ProjectId = req.ProjectId
				bindReq.SSLId = sdk.String(ctx.PickResourceID(sslID))
				bindReq.VServerId = sdk.String(resp.VServerId)
				bindReq.ULBId = req.ULBId
				_, err := client.BindSSL(bindReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ssl certificate[%s] bind with vserver[%s] of ulb[%s]\n", sslID, *bindReq.VServerId, *bindReq.ULBId)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB instance which the VServer to create belongs to")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.VServerName = flags.String("name", "", "Optional. Name of VServer to create")
	req.ListenType = flags.String("listen-type", "RequestProxy", "Optional. Listen type, 'RequestProxy' or 'PacketsTransmit'")
	req.Protocol = flags.String("protocol", "HTTP", "Optional. Protocol of VServer instance, 'HTTP','HTTPS','TCP' for listen type 'RequestProxy' and 'TCP','UDP' for listen type 'PacketsTransmit'")
	req.FrontendPort = flags.Int("port", 80, "Optional. Port of VServer instance")
	flags.StringVar(&sslID, "ssl-id", "", "Optional. Required if you choose HTTPS, Resource ID of SSL Certificate")
	req.Method = flags.String("lb-method", "Roundrobin", "Optional. LB methods, accept values:Roundrobin,Source,ConsistentHash,SourcePort,ConsistentHashPort,WeightRoundrobin and Leastconn. \nConsistentHash,SourcePort and ConsistentHashPort are effective for listen type PacketsTransmit only;\nLeastconn is effective for listen type RequestProxy only;\nRoundrobin,Source and WeightRoundrobin are effective for both listen types")
	req.PersistenceType = flags.String("session-maintain-mode", "None", "Optional. The method of maintaining user's session. Accept values: 'None','ServerInsert' and 'UserDefined'. 'None' meaning don't maintain user's session'; 'ServerInsert' meaning auto create session key; 'UserDefined' meaning specify session key which accpeted by flag seesion-maintain-key by yourself")
	req.PersistenceInfo = flags.String("session-maintain-key", "", "Optional. Specify a key for maintaining session")
	req.ClientTimeout = flags.Int("client-timeout-seconds", 60, "Optional.Unit seconds. For 'RequestProxy', it's lifetime for idle connections, range (0，86400]. For 'PacketsTransmit', it's the duration of the connection is maintained, range [60，900]")
	req.MonitorType = flags.String("health-check-mode", "Port", "Optional. Method of checking real server's status of health. Accept values:'Port','Path'")
	req.Domain = flags.String("health-check-domain", "", "Optional. Skip this flag if health-check-mode is assigned Port")
	req.Path = flags.String("health-check-path", "", "Optional. Skip this flags if health-check-mode is assigned Port")

	command.SetFlagValues(cmd, "listen-type", "RequestProxy", "PacketsTransmit")
	command.SetFlagValues(cmd, "protocol", "HTTP", "HTTPS", "TCP", "UDP")
	command.SetCompletion(cmd, "lb-method", func() []string {
		if *req.ListenType == "RequestProxy" {
			return []string{"Roundrobin", "Source", "WeightRoundrobin", "Leastconn"}
		} else if *req.ListenType == "PacketsTransmit" {
			return []string{"Roundrobin", "Source", "WeightRoundrobin", "ConsistentHash", "SourcePort", "ConsistentHashPort"}
		}
		return []string{"Roundrobin", "Source", "WeightRoundrobin", "ConsistentHash", "SourcePort", "ConsistentHashPort", "Leastconn"}
	})
	command.SetFlagValues(cmd, "session-maintain-mode", "None", "ServerInsert", "UserDefined")
	command.SetFlagValues(cmd, "health-check-mode", "Port", "Path")
	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "ssl-id", func() []string {
		return getAllSSLCertIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")

	return cmd
}
