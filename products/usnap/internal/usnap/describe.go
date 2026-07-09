package usnap

import (
	"fmt"

	"github.com/spf13/cobra"

	usnapsdk "github.com/ucloud/ucloud-sdk-go/services/usnap"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDescribe ucloud usnap describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, usnapsdk.NewClient)
	req := client.NewDescribeSnapshotServiceRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe USnap snapshot service(s)",
		Long:  "Describe USnap snapshot service(s)",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeSnapshotService(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []SnapshotServiceRow{}
			for _, svc := range resp.DataSet {
				row := SnapshotServiceRow{
					ResourceID:   svc.ServiceId,
					VDiskID:      svc.VDiskId,
					VDiskName:    svc.VDiskName,
					VDiskSize:    fmt.Sprintf("%dGB", svc.VDiskSize),
					VDiskType:    svc.VDiskType,
					Group:        svc.Tag,
					ChargeType:   svc.ChargeType,
					AutoRenew:    svc.AutoRenew,
					Status:       svc.Status,
					Zone:         svc.Zone,
					CreationTime: common.FormatDate(svc.CreateTime),
					Expiration:   common.FormatDate(svc.ExpiredTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.SnapshotServiceId = flags.String("service-id", "", "Optional. Resource ID of the snapshot service")
	req.VDiskId = flags.String("vdisk-id", "", "Optional. Resource ID of the disk")
	req.SnapshotId = flags.String("snapshot-id", "", "Optional. Resource ID of the snapshot")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")

	return cmd
}

// describeUsnapByID returns the poller's describe func, closing over ctx so it
// can build an authed usnap client.
func describeUsnapByID(ctx *cli.Context) func(serviceID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(serviceID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, usnapsdk.NewClient)
		req := client.NewDescribeSnapshotServiceRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.SnapshotServiceId = &serviceID
		limit := 50
		req.Limit = &limit
		resp, err := client.DescribeSnapshotService(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) < 1 {
			return nil, nil
		}
		return &resp.DataSet[0], nil
	}
}
