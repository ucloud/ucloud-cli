package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreateReadonly ucloud pgsql db create-readonly
func newCreateReadonly(ctx *cli.Context) *cobra.Command {
	var async bool
	client := newUPgSQLClient(ctx)
	req := client.NewCreateUPgSQLReadonlyRequest()
	cmd := &cobra.Command{
		Use:   "create-readonly",
		Short: "Create a readonly replica for a UPgSQL instance",
		Long:  "Create a readonly replica synchronizing from a source UPgSQL instance",
		Run: func(c *cobra.Command, args []string) {
			*req.SrcInstanceID = ctx.PickResourceID(*req.SrcInstanceID)
			if c.Flags().Changed("vpc-id") {
				*req.VPCID = ctx.PickResourceID(*req.VPCID)
			}
			if c.Flags().Changed("subnet-id") {
				*req.SubnetID = ctx.PickResourceID(*req.SubnetID)
			}
			resp, err := client.CreateUPgSQLReadonly(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			instanceID := resp.InstanceID
			if instanceID == "" {
				ctx.HandleError(fmt.Errorf("empty InstanceID in response"))
				return
			}
			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "pgsql[%s] is initializing\n", instanceID)
			} else {
				text := fmt.Sprintf("pgsql[%s] is initializing", instanceID)
				ctx.PollerTo(w, describePgsqlByID(ctx)).Spoll(instanceID, text, []string{PGSQL_RUNNING, PGSQL_INIT_FAILED, PGSQL_START_FAILED})
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "create-readonly", Status: "Initializing"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of the readonly replica")
	req.SrcInstanceID = flags.String("src-instance-id", "", "Required. Resource ID of the source UPgSQL instance")
	req.MachineType = flags.String("machine-type", "", "Required. Machine type ID, e.g. o.pgsql2m.medium. See 'ucloud pgsql db list-machine-type'")
	req.DiskSpace = flags.Int("disk-size-gb", 0, "Required. Disk space (GiB)")
	req.Port = flags.Int("port", 5432, "Optional. Port of the readonly replica, default 5432")
	req.VPCID = flags.String("vpc-id", "", "Optional. VPC ID. Defaults to the source instance's VPC")
	req.SubnetID = flags.String("subnet-id", "", "Optional. Subnet ID. Defaults to the source instance's subnet")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetCompletion(cmd, "src-instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, req.GetProjectId(), req.GetRegion())
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCID, req.GetProjectId(), req.GetRegion())
	})
	command.SetCompletion(cmd, "machine-type", func() []string {
		return listMachineTypeIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("src-instance-id")
	cmd.MarkFlagRequired("machine-type")
	cmd.MarkFlagRequired("disk-size-gb")

	return cmd
}
