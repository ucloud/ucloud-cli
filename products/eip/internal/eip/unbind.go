package eip

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUnbind ucloud eip unbind
func newUnbind(ctx *cli.Context) *cobra.Command {
	eipIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUnBindEIPRequest()
	cmd := &cobra.Command{
		Use:     "unbind",
		Short:   "Unbind EIP with uhost",
		Long:    "Unbind EIP with uhost",
		Example: "ucloud eip unbind --eip-id eip-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			results := []cli.OpResultRow{}
			for _, eip := range eipIDs {
				eipIns, err := getEIP(ctx, ctx.PickResourceID(eip))
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.EIPId = sdk.String(ctx.PickResourceID(eip))
				req.ResourceId = sdk.String(eipIns.Resource.ResourceID)
				req.ResourceType = sdk.String(eipIns.Resource.ResourceType)
				_, err = client.UnBindEIP(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "unbind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
				results = append(results, cli.OpResultRow{ResourceID: *req.EIPId, Action: "unbind", Status: "Unbound"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&eipIDs, "eip-id", nil, "Required. Resource ID of eips to unbind with some resource")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("eip-id")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{EIP_USED}, nil)
	})

	return cmd
}

// unbindEIP unbinds an EIP from a resource, returning a log trail. Ported from
// cmd/eip.go; the base.ToQueryMap request-log line is dropped (platform SDK
// handler logs requests now, D-C).
// Retained as the canonical product-local copy for uhost Part 6 (see batch-1
// plan); not yet called within the eip product.
func unbindEIP(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region string) ([]string, error) {
	logs := make([]string, 0)
	eipID = ctx.PickResourceID(eipID)
	ip := net.ParseIP(eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ctx, ip, projectID, region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUnBindEIPRequest()
	req.ResourceId = &resourceID
	req.ResourceType = &resourceType
	req.EIPId = &eipID
	req.ProjectId = sdk.String(ctx.PickResourceID(projectID))
	req.Region = &region
	_, err := client.UnBindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("unbind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("unbind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}
