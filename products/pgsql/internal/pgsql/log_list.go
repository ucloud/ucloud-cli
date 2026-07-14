package pgsql

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newLogList ucloud pgsql log list
func newLogList(ctx *cli.Context) *cobra.Command {
	var beginTime, endTime string
	client := newUPgSQLClient(ctx)
	req := client.NewListUPgSQLLogRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List logs of a UPgSQL instance within a time range",
		Long:  "List logs of a UPgSQL instance within a time range",
		Run: func(c *cobra.Command, args []string) {
			*req.InstanceID = ctx.PickResourceID(*req.InstanceID)
			bt, err := time.Parse(common.DateTimeLayout, beginTime)
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid begin-time (use %s): %w", common.DateTimeLayout, err))
				return
			}
			req.BeginTime = sdk.Int(int(bt.Unix()))
			et, err := time.Parse(common.DateTimeLayout, endTime)
			if err != nil {
				ctx.HandleError(fmt.Errorf("invalid end-time (use %s): %w", common.DateTimeLayout, err))
				return
			}
			req.EndTime = sdk.Int(int(et.Unix()))
			resp, err := client.ListUPgSQLLog(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []PgsqlLogRow{}
			for _, l := range resp.DataSet {
				list = append(list, PgsqlLogRow{
					Name:      l.Name,
					Size:      fmt.Sprintf("%dB", l.Size),
					BeginTime: common.FormatDateTime(l.BeginTime),
					EndTime:   common.FormatDateTime(l.EndTime),
				})
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.InstanceID = flags.String("instance-id", "", "Required. Resource ID of the UPgSQL instance")
	flags.StringVar(&beginTime, "begin-time", "", "Required. Begin time, e.g. 2019-01-02/15:04:05")
	flags.StringVar(&endTime, "end-time", "", "Required. End time, e.g. 2019-01-02/15:04:05")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("instance-id")
	cmd.MarkFlagRequired("begin-time")
	cmd.MarkFlagRequired("end-time")
	command.SetCompletion(cmd, "instance-id", func() []string {
		return getUPgSQLIDList(ctx, req.GetProjectId(), req.GetRegion(), req.GetZone())
	})

	return cmd
}
