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

// newBind ucloud eip bind
func newBind(ctx *cli.Context) *cobra.Command {
	var projectID, region, resourceID, resourceType *string
	var eipIDs []string
	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "Bind EIP with uhost",
		Long:    "Bind EIP with uhost",
		Example: "ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx",
		Run: func(cmd *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, eipID := range eipIDs {
				if err := bindEIP(ctx, resourceID, resourceType, &eipID, projectID, region); err == nil {
					results = append(results, cli.OpResultRow{ResourceID: ctx.PickResourceID(eipID), Action: "bind", Status: "Bound"})
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	cmd.Flags().StringSliceVar(&eipIDs, "eip-id", nil, "Required. EIPId to bind")
	resourceID = cmd.Flags().String("resource-id", "", "Required. ResourceID , which is the UHostId of uhost")
	resourceType = cmd.Flags().String("resource-type", "uhost", "Requried. ResourceType, type of resource to bind with eip. 'uhost','vrouter','ulb','upm','hadoophost'.eg..")
	projectID = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")

	command.SetFlagValues(cmd, "resource-type", "uhost", "vrouter", "ulb", "upm", "hadoophost", "fortresshost", "udockhost", "udhost", "natgw", "udb", "vpngw", "ucdr", "dbaudit")
	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *projectID, *region, []string{EIP_FREE}, nil)
	})

	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-id")

	return cmd
}

// bindEIP binds an EIP to a resource. Ported from cmd/eip.go
// (base.BizClient → cli.NewServiceClient; progress→ProgressWriter,
// errors→ctx.HandleError). Returns a non-nil error when the bind fails so the
// caller only emits a structured "Bound" result on success (machine output must
// not report success for a failed operation).
func bindEIP(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region *string) error {
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ctx, ip, *projectID, *region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			*eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(ctx.PickResourceID(*eipID))
	req.ProjectId = sdk.String(ctx.PickResourceID(*projectID))
	req.Region = region
	_, err := client.BindEIP(req)
	if err != nil {
		ctx.HandleError(err)
		return err
	}
	fmt.Fprintf(ctx.ProgressWriter(), "bind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
	return nil
}

// sbindEIP binds an EIP to a resource, returning a log trail instead of
// printing (used for concurrent flows). Ported from cmd/eip.go; the
// base.ToQueryMap request-log line is dropped (platform SDK handler logs
// requests now, D-C).
// Retained as the canonical product-local copy for uhost Part 6 (see batch-1
// plan); not yet called within the eip product.
func sbindEIP(ctx *cli.Context, resourceID, resourceType, eipID, projectID, region *string) ([]string, error) {
	logs := make([]string, 0)
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ctx, ip, *projectID, *region)
		if err != nil {
			ctx.HandleError(err)
		} else {
			*eipID = id
		}
	}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(ctx.PickResourceID(*eipID))
	req.ProjectId = sdk.String(ctx.PickResourceID(*projectID))
	req.Region = region
	_, err := client.BindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("bind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("bind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}
