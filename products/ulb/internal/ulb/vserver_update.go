package ulb

import (
	"fmt"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newVServerUpdate returns ucloud ulb vserver update.
func newVServerUpdate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewUpdateVServerAttributeRequest()
	vserverIDs := []string{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update attributes of VServer instances",
		Long:  "Update attributes of VServer instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.VServerName == "" {
				req.VServerName = nil
			}
			if *req.Method == "" {
				req.Method = nil
			}
			if *req.PersistenceType == "" {
				req.PersistenceType = nil
			}
			if *req.PersistenceInfo == "" {
				req.PersistenceInfo = nil
			}
			if *req.ClientTimeout == -1 {
				req.ClientTimeout = nil
			}
			if *req.MonitorType == "" {
				req.MonitorType = nil
			}
			if *req.Domain == "" {
				req.Domain = nil
			}
			if *req.Path == "" {
				req.Path = nil
			}
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			req.ULBId = sdk.String(ctx.PickResourceID(*req.ULBId))
			results := []cli.OpResultRow{}
			for _, idname := range vserverIDs {
				id := ctx.PickResourceID(idname)
				req.VServerId = sdk.String(id)
				_, err := client.UpdateVServerAttribute(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "ulb-vserver[%s] updated\n", *req.VServerId)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "update-vserver", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.ULBId = flags.String("ulb-id", "", "Required. Resource ID of ULB instance which the VServer to create belongs to")
	flags.StringSliceVar(&vserverIDs, "vserver-id", nil, "Required. Resource ID of Vserver to update")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.VServerName = flags.String("name", "", "Optional. Name of VServer")
	req.Method = flags.String("lb-method", "", "Optional. LB methods, accept values:Roundrobin,Source,ConsistentHash,SourcePort,ConsistentHashPort,WeightRoundrobin and Leastconn. \nConsistentHash,SourcePort and ConsistentHashPort are effective for listen type PacketsTransmit only;\nLeastconn is effective for listen type RequestProxy only;\nRoundrobin,Source and WeightRoundrobin are effective for both listen types")
	req.PersistenceType = flags.String("session-maintain-mode", "", "Optional. The method of maintaining user's session. Accept values: 'None','ServerInsert' and 'UserDefined'. 'None' meaning don't maintain user's session'; 'ServerInsert' meaning auto create session key; 'UserDefined' meaning specify session key which accpeted by flag seesion-maintain-key by yourself")
	req.PersistenceInfo = flags.String("session-maintain-key", "", "Optional. Specify a key for maintaining session")
	req.ClientTimeout = flags.Int("client-timeout-seconds", -1, "Optional.Unit seconds. For 'RequestProxy', it's lifetime for idle connections, range (0，86400]. For 'PacketsTransmit', it's the duration of the connection is maintained, range [60，900]")
	req.MonitorType = flags.String("health-check-mode", "", "Optional. Method of checking real server's status of health. Accept values:'Port','Path'")
	req.Domain = flags.String("health-check-domain", "", "Optional. Skip this flag if health-check-mode is assigned Port")
	req.Path = flags.String("health-check-path", "", "Optional. Skip this flags if health-check-mode is assigned Port")

	command.SetFlagValues(cmd, "lb-method", "Roundrobin", "Source", "WeightRoundrobin", "ConsistentHash", "SourcePort", "ConsistentHashPort", "Leastconn")
	command.SetFlagValues(cmd, "session-maintain-mode", "None", "ServerInsert", "UserDefined")
	command.SetFlagValues(cmd, "health-check-mode", "Port", "Path")
	command.SetCompletion(cmd, "ulb-id", func() []string {
		return getAllULBIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vserver-id", func() []string {
		ulbID := ctx.PickResourceID(*req.ULBId)
		return getAllVServerIDNames(ctx, ulbID, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("ulb-id")
	cmd.MarkFlagRequired("vserver-id")

	return cmd
}
