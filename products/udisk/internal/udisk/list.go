package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList ucloud udisk list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewDescribeUDiskRequest()
	typeMap := map[string]string{
		"DataDisk":    "Oridinary-Data-Disk",
		"SystemDisk":  "Oridinary-System-Disk",
		"SSDDataDisk": "SSD-Data-Disk",
	}
	arkModeMap := map[string]string{
		"Yes": "true",
		"No":  "false",
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List udisk instance",
		Long:  "List udisk instance",
		Run: func(cmd *cobra.Command, args []string) {
			for key, val := range typeMap {
				if *req.DiskType == val {
					*req.DiskType = key
				}
			}
			resp, err := client.DescribeUDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []DiskRow{}
			for _, disk := range resp.DataSet {
				row := DiskRow{
					ResourceID:     disk.UDiskId,
					Name:           disk.Name,
					Group:          disk.Tag,
					Size:           fmt.Sprintf("%dGB", disk.Size),
					Type:           typeMap[disk.DiskType],
					EnableDataArk:  arkModeMap[disk.UDataArkMode],
					MountUHost:     fmt.Sprintf("%s/%s", disk.UHostName, disk.UHostIP),
					MountPoint:     disk.DeviceName,
					State:          disk.Status,
					CreationTime:   common.FormatDate(disk.CreateTime),
					ExpirationTime: common.FormatDate(disk.ExpiredTime),
				}
				if disk.UHostIP == "" {
					row.MountUHost = ""
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
	req.UDiskId = flags.String("udisk-id", "", "Optional. Resource ID of the udisk to search")
	req.DiskType = flags.String("udisk-type", "", "Optional. Optional. Type of the udisk to search. 'Oridinary-Data-Disk','Oridinary-System-Disk' or 'SSD-Data-Disk'")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset")
	req.Limit = cmd.Flags().Int("limit", 50, "Optional. Limit")
	command.SetFlagValues(cmd, "udisk-type", "Oridinary-Data-Disk", "Oridinary-System-Disk", "SSD-Data-Disk")
	return cmd
}
