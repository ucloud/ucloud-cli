package gssh

import (
	"strings"

	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

var areaCodeMap = map[string]string{
	"LAX": "LosAngeles",
	"SIN": "Singapore",
	"HKG": "HongKong",
	"HND": "Tokyo",
	"IAD": "Washington",
	"FRA": "Frankfurt",
	"LOS": "Lagos",
}

var regionLabel = map[string]string{
	"cn-bj1":       "Beijing1",
	"cn-bj2":       "Beijing2",
	"cn-sh2":       "Shanghai2",
	"cn-gd":        "Guangzhou",
	"cn-qz":        "Quanzhou",
	"hk":           "Hongkong",
	"us-ca":        "LosAngeles",
	"us-ws":        "Washington",
	"ge-fra":       "Frankfurt",
	"th-bkk":       "Bangkok",
	"kr-seoul":     "Seoul",
	"sg":           "Singapore",
	"tw-kh":        "Kaohsiung",
	"rus-mosc":     "Moscow",
	"jpn-tky":      "Tokyo",
	"tw-tp":        "TaiPei",
	"uae-dubai":    "Dubai",
	"idn-jakarta":  "Jakarta",
	"ind-mumbai":   "Bombay",
	"bra-saopaulo": "SaoPaulo",
	"uk-london":    "London",
	"afr-nigeria":  "Lagos",
}

// newArea ucloud gssh location
func newArea(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	req := client.NewDescribeGlobalSSHAreaRequest()
	cmd := &cobra.Command{
		Use:   "location",
		Short: "List SSH server locations and covered areas",
		Long:  "List SSH server locations and covered areas",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeGlobalSSHArea(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]GsshLocation, 0)
			for _, item := range resp.AreaSet {
				row := GsshLocation{
					AirportCode:       item.AreaCode,
					SSHServerLocation: areaCodeMap[item.AreaCode],
				}
				regionLabels := make([]string, 0)
				for _, region := range item.RegionSet {
					regionLabels = append(regionLabels, regionLabel[region])
				}
				row.CoveredArea = strings.Join(regionLabels, ",")
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	return cmd
}
