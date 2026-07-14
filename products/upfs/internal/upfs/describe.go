package upfs

import (
	"fmt"

	"github.com/spf13/cobra"

	upfssdk "github.com/ucloud/ucloud-sdk-go/services/upfs"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDescribe ucloud upfs describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, upfssdk.NewClient)
	req := client.NewDescribeUPFSVolumeRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe UPFS volume(s)",
		Long:  "Describe UPFS volume(s)",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeUPFSVolume(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []VolumeRow{}
			for _, vol := range resp.DataSet {
				row := VolumeRow{
					ResourceID:   vol.VolumeId,
					Name:         vol.VolumeName,
					Group:        vol.Tag,
					Size:         fmt.Sprintf("%dGB", vol.Size),
					ProtocolType: vol.ProtocolType,
					MountAddress: vol.MountAddress,
					ChargeType:   vol.ChargeType,
					State:        vol.IsExpired,
					CreationTime: common.FormatDate(vol.CreateTime),
					Expiration:   common.FormatDate(vol.ExpiredTime),
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.VolumeId = flags.String("volume-id", "", "Optional. Resource ID of the UPFS volume")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	return cmd
}

// describeUpfsByID returns the poller's describe func, closing over ctx so it
// can build an authed upfs client.
func describeUpfsByID(ctx *cli.Context) func(volumeID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(volumeID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, upfssdk.NewClient)
		req := client.NewDescribeUPFSVolumeRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.VolumeId = &volumeID
		limit := 50
		req.Limit = &limit
		resp, err := client.DescribeUPFSVolume(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) < 1 {
			return nil, nil
		}
		return &resp.DataSet[0], nil
	}
}
