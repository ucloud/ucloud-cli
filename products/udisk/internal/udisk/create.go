package udisk

import (
	"fmt"

	"github.com/spf13/cobra"

	udisksdk "github.com/ucloud/ucloud-sdk-go/services/udisk"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud udisk create
func newCreate(ctx *cli.Context) *cobra.Command {
	var async *bool
	var count *int
	var enableDataArk *string
	var snapshotID *string
	client := cli.NewServiceClient(ctx, udisksdk.NewClient)
	req := client.NewCreateUDiskRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create udisk instance",
		Long:  "Create udisk instance",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			if *count > 10 || *count < 1 {
				fmt.Fprintf(w, "Error, count should be between 1 and 10\n")
				return
			}
			if *enableDataArk == "true" {
				req.UDataArkMode = sdk.String("Yes")
			} else {
				req.UDataArkMode = sdk.String("No")
			}

			if *req.DiskType == "Oridinary" {
				*req.DiskType = "DataDisk"
			} else if *req.DiskType == "SSD" {
				*req.DiskType = "SSDDataDisk"
			}
			if *snapshotID != "" {
				cloneReq := client.NewCloneUDiskSnapshotRequest()
				cloneReq.UDataArkMode = req.UDataArkMode
				cloneReq.SourceId = snapshotID
				cloneReq.ProjectId = req.ProjectId
				cloneReq.Region = req.Region
				cloneReq.Zone = req.Zone
				cloneReq.Name = req.Name
				cloneReq.Size = req.Size
				cloneReq.ChargeType = req.ChargeType
				cloneReq.Quantity = req.Quantity
				for i := 0; i < *count; i++ {
					resp, err := client.CloneUDiskSnapshot(cloneReq)
					if err != nil {
						ctx.HandleError(err)
						return
					}
					if count := len(resp.UDiskId); count == 1 {
						text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
						if *async {
							fmt.Fprintln(w, text)
						} else {
							ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId[0], text, []string{DISK_AVAILABLE, DISK_FAILED})
						}
						results = append(results, cli.OpResultRow{ResourceID: resp.UDiskId[0], Action: "create", Status: "Initializing"})
					} else if count > 1 {
						fmt.Fprintf(w, "udisk:%v created\n", resp.UDiskId)
						for _, id := range resp.UDiskId {
							results = append(results, cli.OpResultRow{ResourceID: id, Action: "create", Status: "Created"})
						}
					} else {
						fmt.Fprintln(ctx.Err(), "Error: none udisk created")
					}
				}
			} else {
				for i := 0; i < *count; i++ {
					resp, err := client.CreateUDisk(req)
					if err != nil {
						ctx.HandleError(err)
						return
					}
					if count := len(resp.UDiskId); count == 1 {
						text := fmt.Sprintf("udisk:%v is initializing", resp.UDiskId)
						if *async {
							fmt.Fprintln(w, text)
						} else {
							ctx.PollerTo(w, describeUdiskByID(ctx)).Spoll(resp.UDiskId[0], text, []string{DISK_AVAILABLE, DISK_FAILED})
						}
						results = append(results, cli.OpResultRow{ResourceID: resp.UDiskId[0], Action: "create", Status: "Initializing"})
					} else if count > 1 {
						fmt.Fprintf(w, "udisk:%v created\n", resp.UDiskId)
						for _, id := range resp.UDiskId {
							results = append(results, cli.OpResultRow{ResourceID: id, Action: "create", Status: "Created"})
						}
					} else {
						fmt.Fprintln(ctx.Err(), "Error: none udisk created")
					}
				}
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Name = flags.String("name", "", "Required. Name of the udisk to create")
	req.Size = flags.Int("size-gb", 10, "Required. Size of the udisk to create. Unit:GB. Normal udisk [1,8000]; SSD udisk [1,4000] ")
	snapshotID = flags.String("snapshot-id", "", "Optional. Resource ID of a snapshot, which will apply to the udisk being created. If you set this option, 'udisk-type' will be omitted.")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	req.ChargeType = flags.String("charge-type", "Dynamic", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	enableDataArk = flags.String("enable-data-ark", "false", "Optional. DataArk supports real-time backup, which can restore the udisk back to any moment within the last 12 hours.")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.DiskType = flags.String("udisk-type", "Oridinary", "Optional. 'Ordinary' or 'SSD'")
	async = flags.Bool("async", false, "Optional. Do not wait for the long-running operation to finish.")
	count = flags.Int("count", 1, "Optional. The count of udisk to create. Range [1,10]")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "enable-data-ark", "true", "false")
	command.SetFlagValues(cmd, "udisk-type", "Oridinary", "SSD")

	cmd.MarkFlagRequired("size-gb")
	cmd.MarkFlagRequired("name")

	return cmd
}
