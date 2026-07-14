package bw

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newPkgCreate returns ucloud bw pkg create.
func newPkgCreate(ctx *cli.Context) *cobra.Command {
	var start, end *string
	timeLayout := "2006-01-02/15:04:05"
	ids := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateBandwidthPackageRequest()
	loc, _ := time.LoadLocation("Local")
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create bandwidth package",
		Long:    "Create bandwidth package",
		Example: "ucloud bw pkg create --eip-id eip-xxx --bandwidth-mb 20 --start-time 2018-12-15/09:20:00 --end-time 2018-12-16/09:20:00",
		Run: func(c *cobra.Command, args []string) {
			st, err := time.ParseInLocation(timeLayout, *start, loc)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			et, err := time.ParseInLocation(timeLayout, *end, loc)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if st.Sub(time.Now()) < 0 {
				fmt.Fprintln(ctx.ProgressWriter(), "start-time must be after the current time")
				return
			}
			du := et.Unix() - st.Unix()
			if du <= 0 {
				fmt.Fprintln(ctx.ProgressWriter(), "end-time must be after the start-time")
				return
			}
			req.EnableTime = sdk.Int(int(st.Unix()))
			req.TimeRange = sdk.Int(int(du))

			results := []cli.OpResultRow{}
			for _, id := range ids {
				id = ctx.PickResourceID(id)
				req.EIPId = &id
				resp, err := client.CreateBandwidthPackage(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "bandwidth package[%s] created for eip[%s]\n", resp.BandwidthPackageId, id)
				results = append(results, cli.OpResultRow{ResourceID: resp.BandwidthPackageId, Action: "create", Status: "Created"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "eip-id", nil, "Required. Resource ID of eip to be bound with created bandwidth package")
	start = flags.String("start-time", "", "Required. The time to enable bandwidth package. Local time, for example '2018-12-25/08:30:00'")
	end = flags.String("end-time", "", "Required. The time to disable bandwidth package. Local time, for example '2018-12-26/08:30:00'")
	req.Bandwidth = flags.Int("bandwidth-mb", 0, "Required. bandwidth of the bandwidth package to create.Range [1,800]. Unit:'Mb'.")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetCompletion(cmd, "eip-id", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{EIP_USED}, []string{EIP_CHARGE_BANDWIDTH})
	})

	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("start-time")
	cmd.MarkFlagRequired("end-time")
	cmd.MarkFlagRequired("bandwidth-mb")
	return cmd
}
