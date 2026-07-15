package pgsql

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreate ucloud pgsql db create
func newCreate(ctx *cli.Context) *cobra.Command {
	var paramGroupID int
	var async bool
	var password string

	client := newUPgSQLClient(ctx)
	req := client.NewCreateUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UPgSQL instance",
		Long:  "Create a UPgSQL instance on UCloud platform",
		Run: func(c *cobra.Command, args []string) {
			if len(*req.Name) < 6 {
				ctx.HandleError(fmt.Errorf("name must be at least 6 characters"))
				return
			}
			if password == "" {
				ctx.HandleError(fmt.Errorf("admin password is required"))
				return
			}

			// ParamGroupID: user-provided wins; otherwise auto-fetch a default template.
			if c.Flags().Changed("param-group-id") {
				req.ParamGroupID = sdk.Int(paramGroupID)
			} else {
				id, err := getDefaultParamGroupID(ctx, *req.DBVersion, req.GetProjectId(), req.GetRegion(), req.GetZone())
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.ParamGroupID = sdk.Int(id)
			}

			// VPCID/SubnetID accept "id/name" form; pick the id.
			*req.VPCID = ctx.PickResourceID(*req.VPCID)
			*req.SubnetID = ctx.PickResourceID(*req.SubnetID)
			req.AdminPassword = sdk.String(password)

			resp, err := client.CreateUPgSQLInstance(req)
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
			ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "create", Status: "Initializing"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Required flags
	req.Name = flags.String("name", "", "Required. Instance name, at least 6 characters")
	flags.StringVar(&password, "password", "", "Required. Admin password")
	req.DBVersion = flags.String("version", "", "Required. DB version. Options: postgresql-10.4, postgresql-13.4")
	req.MachineType = flags.String("machine-type", "", "Required. Machine type ID, e.g. o.pgsql2m.medium. See 'ucloud pgsql db list-machine-type'")
	req.VPCID = flags.String("vpc-id", "", "Required. VPC ID. See 'ucloud vpc list'")
	req.SubnetID = flags.String("subnet-id", "", "Required. Subnet ID. See 'ucloud subnet list'")

	// Optional flags
	flags.IntVar(&paramGroupID, "param-group-id", 0, "Optional. Param group ID. Auto-fetched if omitted. See 'ucloud pgsql conf list'")
	req.DiskSpace = flags.String("disk-size-gb", "100", "Optional. Disk size (GiB), at least 20, default 100")
	req.Port = flags.Int("port", 5432, "Optional. Port, default 5432")
	req.InstanceMode = flags.String("mode", "Normal", "Optional. Normal / HA")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for creation to finish")

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "version", pgsqlVersionList...)
	command.SetFlagValues(cmd, "mode", "Normal", "HA")
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, req.GetProjectId(), req.GetRegion())
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCID, req.GetProjectId(), req.GetRegion())
	})
	command.SetCompletion(cmd, "param-group-id", func() []string {
		return listParamTemplateIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})
	command.SetCompletion(cmd, "machine-type", func() []string {
		return listMachineTypeIDNames(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("machine-type")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("subnet-id")

	return cmd
}
