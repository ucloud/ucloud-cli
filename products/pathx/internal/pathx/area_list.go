package pathx

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newAreaList ucloud pathx area list
func newAreaList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	areaGetReq := client.NewDescribeUGA3AreaRequest()
	optimizationReq := client.NewDescribeUGA3OptimizationRequest()
	var timeRange, accelerationArea, originDomain, originIp string
	var noAccel bool
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List origin area or acceleration area information",
		Long:    "Provide optional flags to get the optional list of global access source stations",
		Example: "ucloud pathx area list --origin-ip 0.0.0.0 --origin-domain test.com",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			if len(originDomain) == 0 && len(originIp) == 0 {
				response, err := client.DescribeUGA3Area(areaGetReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				forwardAreas := response.AreaSet
				if len(forwardAreas) == 0 {
					ctx.HandleError(fmt.Errorf("Not found the origin area list"))
					return
				}
				areasGroup := make(map[string][]PathxOptionalAreaRow)
				for _, item := range forwardAreas {
					areasGroup[item.ContinentCode] = append(areasGroup[item.ContinentCode], PathxOptionalAreaRow{
						AreaCode:    item.AreaCode,
						Area:        item.Area,
						CountryCode: item.CountryCode,
						FlagUnicode: item.FlagUnicode,
						FlagEmoji:   item.FlagEmoji,
					})
				}
				fmt.Fprintln(w, "Origin areas :")
				for area, rows := range areasGroup {
					fmt.Fprintf(w, "ContinentCode:  %s\n", area)
					ctx.PrintList(rows)
					fmt.Fprintln(w)
				}
				return
			}
			areaGetReq.Domain = &originDomain
			areaGetReq.IPList = &originIp
			response, err := client.DescribeUGA3Area(areaGetReq)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			forwardAreas := response.AreaSet
			if len(forwardAreas) == 0 {
				ctx.HandleError(fmt.Errorf("Not found the origin area list"))
				return
			}
			forwardArea := forwardAreas[0]
			fmt.Fprintf(w, "Recommend origin area:(%s)\n", forwardArea.ContinentCode)
			ctx.PrintList([]PathxOptionalAreaRow{{
				AreaCode:    forwardArea.AreaCode,
				Area:        forwardArea.Area,
				CountryCode: forwardArea.CountryCode,
				FlagUnicode: forwardArea.FlagUnicode,
				FlagEmoji:   forwardArea.FlagEmoji,
			}})
			fmt.Fprintln(w)
			if !noAccel {
				areaCode := forwardAreas[0].AreaCode
				optimizationReq.AreaCode = &areaCode
				optimizationReq.AccelerationArea = &accelerationArea
				optimizationReq.TimeRange = &timeRange
				optimizationReq.SetProjectIdRef(areaGetReq.GetProjectIdRef())
				optimizationReq.SetRegionRef(areaGetReq.GetRegionRef())
				optimizationReq.SetZoneRef(areaGetReq.GetZoneRef())
				optimizationResponse, err := client.DescribeUGA3Optimization(optimizationReq)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				accelerationInfos := optimizationResponse.AccelerationInfos
				if len(accelerationInfos) == 0 {
					ctx.HandleError(fmt.Errorf("Not found the acceleration area information."))
					return
				}
				fmt.Fprintln(w, "Acceleration areas :")
				for _, item := range accelerationInfos {
					if len(accelerationArea) == 0 {
						fmt.Fprintf(w, "%s(%s):\n", item.AccelerationName, item.AccelerationArea)
					}
					list := make([]PathxOptimizationRow, 0)
					for _, node := range item.NodeInfo {
						list = append(list, PathxOptimizationRow{
							Area:         node.Area,
							AreaCode:     node.AreaCode,
							CountryCode:  node.CountryCode,
							FlagUnicode:  node.FlagUnicode,
							FlagEmoji:    node.FlagEmoji,
							Latency:      fmt.Sprintf("%s%s", strconv.FormatFloat(node.Latency, 'g', 12, 64), "ms"),
							LatencyWAN:   fmt.Sprintf("%s%s", strconv.FormatFloat(node.LatencyInternet, 'g', 12, 64), "ms"),
							LatencyPathX: fmt.Sprintf("%s%s", strconv.FormatFloat(node.LatencyOptimization, 'g', 12, 64), "%"),
							Loss:         fmt.Sprintf("%s%s", strconv.FormatFloat(node.Loss, 'g', 12, 64), "%"),
							LossWAN:      fmt.Sprintf("%s%s", strconv.FormatFloat(node.LossInternet, 'g', 12, 64), "%"),
							LossPathx:    fmt.Sprintf("%s%s", strconv.FormatFloat(node.LossOptimization, 'g', 12, 64), "%"),
						})
					}
					ctx.PrintList(list)
				}
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindProjectID(cmd, areaGetReq)
	ctx.BindRegion(cmd, areaGetReq)
	ctx.BindZone(cmd, areaGetReq)
	flags.StringVar(&timeRange, "time-range", "", "Optional. The default value is 1 day. Acceptable values:'Hour','Day','Week',and its value is not case sensitive")
	flags.StringVar(&accelerationArea, "accel", "", "Optional. The acceleration area,acceptable values:'Global','AP','EU','ME','OA','AF','NA','SA'")
	flags.StringVar(&originDomain, "origin-domain", "", "Optional. If you fill in the IP or domain name, a region will be recommended as the first in the return list")
	flags.StringVar(&originIp, "origin-ip", "", "Optional. If you fill in the IP or domain name, a region will be recommended as the first IP collection of the source station in the return list, split by ',' example:110.10.10.1,111.100.0.10 ")
	flags.BoolVar(&noAccel, "no-accel", false, "Optional. If it is specified,the print result will not be displayed acceleration areas")
	return cmd
}
