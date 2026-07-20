package ugn

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// routeRow is the table row for routes in ugn route.
type routeRow struct {
	DstAddr       string
	NextHopID     string
	NextHopType   string
	NextHopRegion string
	Priority      int
	Conflict      string
	Deny          string
	Restrict      string
}

// newRoute ucloud ugn route
func newRoute(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewGetUGNRouteTableRequest()

	cmd := &cobra.Command{
		Use:   "route",
		Short: "Show route table of one ugn instance",
		Long:  "Show route table of one ugn instance",
		Run: func(c *cobra.Command, args []string) {
			*req.UGNID = ctx.PickResourceID(*req.UGNID)
			resp, err := client.GetUGNRouteTable(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			if *req.Type == "Final" {
				for _, vr := range resp.VRoutes {
					fmt.Fprintf(ctx.ProgressWriter(), "Network %s (%d routes):\n", vr.NetworkId, len(vr.Routes))
					printRouteRows(ctx, vr.Routes)
					fmt.Fprintln(ctx.ProgressWriter())
				}
			} else {
				fmt.Fprintf(ctx.ProgressWriter(), "Route Table (%s, %d routes):\n", *req.Type, len(resp.Routes))
				printRouteRows(ctx, resp.Routes)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")
	req.Type = flags.String("type", "Final", "Route table type: Origin/Middle/Final")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("ugn-id")
	ctx.SetCompletion(cmd, "ugn-id", func() []string {
		return getAllUGNIdNames(ctx, *req.ProjectId)
	})
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)
	command.SetFlagValues(cmd, "type", "Origin", "Middle", "Final")

	return cmd
}

func printRouteRows(ctx *cli.Context, routes []ugnsdk.SimpleRoute) {
	if len(routes) == 0 {
		return
	}
	rows := make([]routeRow, 0, len(routes))
	for _, r := range routes {
		rows = append(rows, routeRow{
			DstAddr:       r.DstAddr,
			NextHopID:     r.NextHopID,
			NextHopType:   r.NextHopType,
			NextHopRegion: r.NextHopRegion,
			Priority:      r.Priority,
			Conflict:      strconv.FormatBool(r.Conflict),
			Deny:          strconv.FormatBool(r.Deny),
			Restrict:      strconv.FormatBool(r.Restrict),
		})
	}
	ctx.PrintList(rows)
}
