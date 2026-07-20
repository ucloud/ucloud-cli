package ugn

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// getNetworkRow is the table row for networks in ugn get.
type getNetworkRow struct {
	NetworkID  string
	Name       string
	OrgName    string
	Region     string
	Type       string
	CreateTime string
}

// getBwRow is the table row for bw packages in ugn get.
type getBwRow struct {
	PackageID     string
	Name          string
	BandwidthMbps float64
	RegionA       string
	RegionB       string
	Path          string
	QoS           string
	PayMode       string
	CreateTime    string
	ExpireTime    string
}

// getRouteRow is the table row for routes in ugn get.
type getRouteRow struct {
	DstAddr       string
	NextHopID     string
	NextHopType   string
	NextHopRegion string
	Priority      int
	Conflict      string
	Deny          string
	Restrict      string
}

// getPolicyRow is the table row for policies in ugn get.
type getPolicyRow struct {
	PolicyID      string
	Name          string
	Priority      int
	Direction     string
	Action        string
	RoutePriority int
	Enabled       string
	DstAddrs      string
	SrcAddrs      string
	CreateTime    string
}

// newGet ucloud ugn get
func newGet(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewDescribeSimpleUGNRequest()
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show details of one ugn instance",
		Long:  "Show details of one ugn instance",
		Run: func(c *cobra.Command, args []string) {
			*req.UGNID = ctx.PickResourceID(*req.UGNID)
			resp, err := client.DescribeSimpleUGN(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			// 基本信息
			basic := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: resp.UGN.UGNID},
				{Attribute: "Name", Content: resp.UGN.Name},
				{Attribute: "Remark", Content: resp.UGN.Remark},
				{Attribute: "CreateTime", Content: common.FormatDate(resp.UGN.CreateTime)},
				{Attribute: "NetworkCount", Content: fmt.Sprintf("%d", resp.UGN.NetworkCount)},
				{Attribute: "BwPackageCount", Content: fmt.Sprintf("%d", resp.UGN.BwPackageCount)},
			}
			printDescribe(ctx, basic)

			// 网络实例列表
			fmt.Fprintln(ctx.ProgressWriter())
			fmt.Fprintf(ctx.ProgressWriter(), "Networks (%d):\n", len(resp.Networks))
			if len(resp.Networks) > 0 {
				networkRows := make([]getNetworkRow, 0, len(resp.Networks))
				for _, nw := range resp.Networks {
					networkRows = append(networkRows, getNetworkRow{
						NetworkID:  nw.NetworkID,
						Name:       nw.Name,
						OrgName:    nw.OrgName,
						Region:     nw.Region,
						Type:       nw.Type,
						CreateTime: common.FormatDate(nw.CreateTime),
					})
				}
				ctx.PrintList(networkRows)
			}

			// 带宽包列表
			fmt.Fprintln(ctx.ProgressWriter())
			fmt.Fprintf(ctx.ProgressWriter(), "Bandwidth Packages (%d):\n", len(resp.BwPackages))
			if len(resp.BwPackages) > 0 {
				bwRows := make([]getBwRow, 0, len(resp.BwPackages))
				for _, bw := range resp.BwPackages {
					expireTime := ""
					if bw.ExpireTime > 0 {
						expireTime = common.FormatDate(bw.ExpireTime)
					}
					path := bw.Path
					if path == "None" {
						path = "IGP"
					}
					bwRows = append(bwRows, getBwRow{
						PackageID:     bw.PackageID,
						Name:          bw.Name,
						BandwidthMbps: bw.BandWidth,
						RegionA:       bw.RegionA,
						RegionB:       bw.RegionB,
						Path:          path,
						QoS:           bw.Qos,
						PayMode:       bw.PayMode,
						CreateTime:    common.FormatDate(bw.CreateTime),
						ExpireTime:    expireTime,
					})
				}
				ctx.PrintList(bwRows)
			}

			// 路由列表
			fmt.Fprintln(ctx.ProgressWriter())
			fmt.Fprintf(ctx.ProgressWriter(), "Routes (%d):\n", len(resp.Routes))
			if len(resp.Routes) > 0 {
				routeRows := make([]getRouteRow, 0, len(resp.Routes))
				for _, r := range resp.Routes {
					routeRows = append(routeRows, getRouteRow{
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
				ctx.PrintList(routeRows)
			}

			// 路由策略列表
			fmt.Fprintln(ctx.ProgressWriter())
			fmt.Fprintf(ctx.ProgressWriter(), "Policies (%d):\n", len(resp.Policies))
			if len(resp.Policies) > 0 {
				policyRows := make([]getPolicyRow, 0, len(resp.Policies))
				for _, p := range resp.Policies {
					dstAddrs := make([]string, 0, len(p.DstNetworks))
					for _, n := range p.DstNetworks {
						if len(n.Prefixes) > 0 {
							dstAddrs = append(dstAddrs, fmt.Sprintf("%s(%s)", n.NetworkId, strings.Join(n.Prefixes, ",")))
						} else {
							dstAddrs = append(dstAddrs, n.NetworkId)
						}
					}
					srcAddrs := make([]string, 0, len(p.SrcNetworks))
					for _, n := range p.SrcNetworks {
						if len(n.Prefixes) > 0 {
							srcAddrs = append(srcAddrs, fmt.Sprintf("%s(%s)", n.NetworkId, strings.Join(n.Prefixes, ",")))
						} else {
							srcAddrs = append(srcAddrs, n.NetworkId)
						}
					}
					policyRows = append(policyRows, getPolicyRow{
						PolicyID:      p.PolicyId,
						Name:          p.Name,
						Priority:      p.Priority,
						Direction:     p.Direction,
						Action:        p.Action,
						RoutePriority: p.RoutePriority,
						Enabled:       strconv.FormatBool(p.Enabled),
						DstAddrs:      strings.Join(dstAddrs, ", "),
						SrcAddrs:      strings.Join(srcAddrs, ", "),
						CreateTime:    common.FormatDate(p.CreateTime),
					})
				}
				ctx.PrintList(policyRows)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance to describe")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("ugn-id")
	ctx.SetCompletion(cmd, "ugn-id", func() []string {
		return getAllUGNIdNames(ctx, *req.ProjectId)
	})
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)

	return cmd
}

// printDescribe renders describe rows without column headers in table mode,
// printing each attribute/content pair as an aligned key-value row.
func printDescribe(ctx *cli.Context, rows []cli.DescribeRow) {
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(rows)
		return
	}
	maxWidth := 0
	for _, r := range rows {
		if len(r.Attribute) > maxWidth {
			maxWidth = len(r.Attribute)
		}
	}
	for _, r := range rows {
		fmt.Fprintf(ctx.Out(), "%-*s  %s\n", maxWidth, r.Attribute, r.Content)
	}
}
