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

var dbDiskTypeList = []string{"normal", "sata_ssd", "pcie_ssd"}

// newCreateSlave ucloud udb create-slave
func newCreateSlave(ctx *cli.Context) *cobra.Command {
	var diskType string
	var async bool
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewCreateUDBSlaveRequest()
	cmd := &cobra.Command{
		Use:   "create-slave",
		Short: "Create slave database",
		Long:  "Create slave database",
		Run: func(c *cobra.Command, args []string) {
			*req.SrcId = ctx.PickResourceID(*req.SrcId)
			switch diskType {
			case "normal":
				req.UseSSD = sdk.Bool(false)
			case "sata_ssd":
				req.UseSSD = sdk.Bool(true)
				req.SSDType = sdk.String("SATA")
			case "pcie_ssd":
				req.UseSSD = sdk.Bool(true)
				req.SSDType = sdk.String("PCI-E")
			}
			*req.MemoryLimit *= 1000
			resp, err := client.CreateUDBSlave(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "udb[%s] is initializing\n", resp.DBId)
			} else {
				ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(resp.DBId, fmt.Sprintf("udb[%s] is initializing", resp.DBId), []string{UDB_RUNNING, UDB_FAIL})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.DBId, Action: "create-slave", Status: "Initializing"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.SrcId = flags.String("master-udb-id", "", "Required. Resource ID of master UDB instance")
	req.Name = flags.String("name", "", "Required. Name of the slave DB to create")
	req.Port = flags.Int("port", 3306, "Optional. Port of the slave db service")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringVar(&diskType, "disk-type", "Normal", fmt.Sprintf("Optional. Setting this flag means using SSD disk. Accept values: %s", strings.Join(dbDiskTypeList, ", ")))
	req.MemoryLimit = flags.Int("memory-size-gb", 1, "Optional. Memory size of udb instance. From 1 to 128. Unit GB")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish")
	req.IsLock = flags.Bool("is-lock", false, "Optional. Lock master DB or not")

	cmd.MarkFlagRequired("master-udb-id")
	cmd.MarkFlagRequired("name")

	command.SetFlagValues(cmd, "disk-type", dbDiskTypeList...)
	command.SetCompletion(cmd, "master-udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}
