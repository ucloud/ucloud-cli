package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPromoteToHA ucloud udb promote-to-ha 低频操作 暂不开放
// Migrated for parity but NOT mounted (mirrors cmd/mysql.go's commented-out registration).
func newPromoteToHA(ctx *cli.Context) *cobra.Command {
	var idNames []string
	var common request.CommonBase
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewPromoteUDBInstanceToHARequest()
	cmd := &cobra.Command{
		Use:   "promote-to-ha",
		Short: "Promote db of normal mode to high availability db. ",
		Long:  "Promote db of normal mode to high availability db",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			for _, idname := range idNames {
				id := ctx.PickResourceID(idname)
				req.DBId = &id
				_, err := client.PromoteUDBInstanceToHA(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(id, fmt.Sprintf("udb[%s] is synchronizing data", id), []string{UDB_TOBE_SWITCH, UDB_FAIL})
				any, err := describeUdbByID(ctx)(id, nil)
				if err != nil {
					ctx.HandleError(fmt.Errorf("udb[%s] promoted failed, please contact technical support; %v", idname, err))
					continue
				}
				ins, ok := any.(*udb.UDBInstanceSet)
				if !ok {
					ctx.HandleError(fmt.Errorf("udb[%s] promoted failed, please contact technical support", idname))
					continue
				}
				if ins.State != UDB_TOBE_SWITCH {
					ctx.HandleError(fmt.Errorf("udb[%s] promoted failed, please contact technical support. udb[%s]'s status:%s", idname, idname, ins.State))
					continue
				}
				switchReq := client.NewSwitchUDBInstanceToHARequest()
				switchReq.DBId = &id
				switchReq.Region = req.Region
				switchReq.ProjectId = req.ProjectId
				switchReq.ChargeType = &ins.ChargeType
				switchReq.Quantity = sdk.String("0")
				// Original read base.ConfigIns.Zone (global default zone); products
				// must not import base. This command is migrated for parity but is
				// NOT mounted (see newMysqlDB), so it is unreachable. Zone falls back
				// to the bound region's CommonBase zone (empty here). See report.
				switchReq.Zone = sdk.String(common.GetZone())
				switchResp, err := client.SwitchUDBInstanceToHA(switchReq)
				if err != nil {
					ctx.HandleError(fmt.Errorf("udb[%s] promoted failed, please contact technical support; %v", idname, err))
					continue
				}
				ctx.PollerTo(w, describeUdbByID(ctx)).Spoll(switchResp.DBId, fmt.Sprintf("udb[%s] is switching to high availability mode", switchResp.DBId), []string{UDB_RUNNING, UDB_FAIL})
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringSliceVar(&idNames, "udb-id", nil, "Required. Resource ID of UDB instances to be promoted as high availability mode")

	cmd.MarkFlagRequired("udb-id")
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, "")
	})
	return cmd
}
