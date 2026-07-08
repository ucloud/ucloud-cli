package mysql

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResize ucloud udb resize
func newResize(ctx *cli.Context) *cobra.Command {
	var diskTypes = []string{"normal", "sata_ssd", "pcie_ssd", "normal_volume", "sata_ssd_volume", "pcie_ssd_volume"}
	var async, yes bool
	var idNames []string
	var memory, disk int
	var diskType string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewResizeUDBInstanceRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Reszie MySQL instances, such as memory size, disk size and disk type",
		Long:  "Reszie MySQL instances, such as memory size, disk size and disk type",
		Run: func(c *cobra.Command, args []string) {
			if diskType != "" {
				switch diskType {
				case "normal":
					req.InstanceType = sdk.String("Normal")
				case "sata_ssd":
					req.InstanceType = sdk.String("SATA_SSD")
				case "pcie_ssd":
					req.InstanceType = sdk.String("PCIE_SSD")
				case "normal_volume":
					req.InstanceType = sdk.String("Normal_Volume")
				case "sata_ssd_volume":
					req.InstanceType = sdk.String("SATA_SSD_Volume")
				case "pcie_ssd_volume":
					req.InstanceType = sdk.String("PCIE_SSD_Volume")
				default:
					req.InstanceType = &diskType
				}
			}

			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.DBId = &id
				any, err := describeUdbByID(ctx)(id, nil)
				if err != nil {
					ctx.HandleError(err)
					continue
				}

				ins, ok := any.(*udb.UDBInstanceSet)
				if !ok {
					continue
				}

				if memory != 0 {
					req.MemoryLimit = sdk.Int(memory * 1000)
				} else {
					req.MemoryLimit = &ins.MemoryLimit
				}
				if disk != 0 {
					req.DiskSpace = &disk
				} else {
					req.DiskSpace = &ins.DiskSpace
				}

				if ins.State == UDB_RUNNING {
					ok, err := ctx.Confirm(yes, fmt.Sprintf("Need to shut down udb[%s] before upgrading, whether to continue?", idname))
					if err != nil {
						ctx.HandleError(err)
						continue
					}
					if !ok {
						continue
					}
					stopReq := client.NewStopUDBInstanceRequest()
					stopReq.ProjectId = req.ProjectId
					stopReq.Region = req.Region
					stopReq.Zone = req.Zone
					stopReq.DBId = req.DBId
					stopUdbIns(ctx, stopReq, false, w)
				}
				_, err = client.ResizeUDBInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(w, "udb[%s] is resizing\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is resizing", idname)
					ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(*req.DBId, text, []string{UDB_RUNNING, UDB_SHUTOFF, UDB_FAIL, UDB_UPGRADE_FAIL})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "resize", Status: "Resizing"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to restart")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.IntVar(&memory, "memory-size-gb", 0, "Optional. Memory size of udb instance. From 1 to 128. Unit GB")
	flags.IntVar(&disk, "disk-size-gb", 0, "Optional. Disk size of udb instance. From 20 to 3000 according to memory size. Unit GB. Step 10GB")
	flags.StringVar(&diskType, "disk-type", "", fmt.Sprintf("Optional. Disk type of udb instance. Accept values:%s", strings.Join(diskTypes, ", ")))
	req.StartAfterUpgrade = flags.Bool("start-after-upgrade", true, "Optional. Automatic start the UDB instances after upgrade")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")

	command.SetFlagValues(cmd, "disk-type", diskTypes...)
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("udb-id")

	return cmd
}
