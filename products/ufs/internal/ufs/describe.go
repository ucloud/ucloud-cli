package ufs

import (
	"fmt"

	"github.com/spf13/cobra"

	ufssdk "github.com/ucloud/ucloud-sdk-go/services/ufs"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDescribe ucloud ufs describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ufssdk.NewClient)
	req := client.NewDescribeUFSVolume2Request()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe UFS volume(s)",
		Long:  "Describe UFS volume(s)",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeUFSVolume2(req)
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
					UsedSize:     fmt.Sprintf("%dGB", vol.UsedSize),
					ProtocolType: vol.ProtocolType,
					StorageType:  vol.StorageType,
					MountPoints:  fmt.Sprintf("%d/%d", vol.TotalMountPointNum, vol.MaxMountPointNum),
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
	req.VolumeId = flags.String("volume-id", "", "Optional. Resource ID of the UFS volume")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	return cmd
}

// describeUfsByID returns the poller's describe func, closing over ctx so it
// can build an authed ufs client.
func describeUfsByID(ctx *cli.Context) func(volumeID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(volumeID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, ufssdk.NewClient)
		req := client.NewDescribeUFSVolume2Request()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.VolumeId = &volumeID
		limit := 50
		req.Limit = &limit
		resp, err := client.DescribeUFSVolume2(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) < 1 {
			return nil, nil
		}
		return &resp.DataSet[0], nil
	}
}
