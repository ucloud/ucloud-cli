package pgsql

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/upgsql"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newList ucloud pgsql db list
func newList(ctx *cli.Context) *cobra.Command {
	var instanceID string
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLInstanceRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List UPgSQL instances",
		Long:  "List UPgSQL instances",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.ListUPgSQLInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []PgsqlInstanceRow{}
			for _, ins := range resp.DataSet {
				if instanceID != "" && ins.InstanceID != instanceID {
					continue
				}
				list = append(list, toInstanceRow(ins))
				for _, slave := range ins.DataSet {
					list = append(list, toInstanceRow(upgsql.UDBInstanceSet{
						Zone:         slave.Zone,
						InstanceID:   slave.InstanceID,
						Name:         slave.Name,
						DBVersion:    slave.DBVersion,
						InstanceMode: slave.InstanceMode,
						State:        slave.State,
						VPCID:        slave.VPCID,
						SubnetID:     slave.SubnetID,
						IP:           slave.IP,
					}))
				}
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&instanceID, "instance-id", "", "Optional. List the specified UPgSQL instance only")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}

// toInstanceRow maps an SDK UDBInstanceSet to a PgsqlInstanceRow. role is non-empty
// for readonly slaves (prefixed with ⮭ to indent under the master row).
func toInstanceRow(ins upgsql.UDBInstanceSet) PgsqlInstanceRow {
	row := PgsqlInstanceRow{
		Name:         ins.Name,
		InstanceID:   ins.InstanceID,
		Zone:         ins.Zone,
		State:        ins.State,
		IP:           ins.IP,
		VPC:          ins.VPCID,
		Subnet:       ins.SubnetID,
		InstanceMode: ins.InstanceMode,
		DBVersion:    ins.DBVersion,
	}
	return row
}
