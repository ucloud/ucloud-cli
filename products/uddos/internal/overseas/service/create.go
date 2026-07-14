// Package service ...
//
// @Brief  创建海外高防服务命令
//
// @File   create.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
//
// @CopyRights(C) UCloud All rights reserved.
package service

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// overseasCityToCleaningCenter 海外城市 → 清洗中心（API EngineRoom 值）
var overseasCityToCleaningCenter = map[string]string{
	// 亚太：HongKong 清洗中心
	"HongKong":  "HongKong",
	"Taipei":    "HongKong",
	"Singapore": "HongKong",
	"Tokyo":     "HongKong",
	"Seoul":     "HongKong",
	"Bangkok":   "HongKong",
	"HoChiMinh": "HongKong",
	"Jakarta":   "HongKong",
	"Manila":    "HongKong",
	"Mumbai":    "HongKong",
	// 欧洲：Frankfurt 清洗中心
	"Frankfurt": "Frankfurt",
	"London":    "Frankfurt",
	"Moscow":    "Frankfurt",
	// 北美：Ashburn 清洗中心
	"Ashburn":    "Ashburn",
	"LosAngeles": "Ashburn",
	"Washington": "Ashburn",
}

// newCreate 构建 uddos overseas service create 命令
//
// @Brief  构建海外高防 service create 子命令，调用 BuyHighProtectGameService（海外参数，ForwardType=Passthrough）
//
// @Param  ctx *cli.Context
//
// @Return *cobra.Command
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
func newCreate(ctx *cli.Context) *cobra.Command {
	var chargeType, areaLine, name string
	var quantity, srcBandwidth int
	var async bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an overseas DDoS high-protection service",
		Long:  "Create an overseas DDoS high-protection service via BuyHighProtectGameService (ForwardType=Passthrough). Defence flow is fixed at 50 Gbps.",
		Example: `  AreaLine (--area-line)                         EngineRoom (API)
  HongKong / Taipei / Singapore / Tokyo /        HongKong
    Seoul / Bangkok / HoChiMinh / Jakarta /
    Manila / Mumbai
  Frankfurt / London / Moscow                    Frankfurt
  Ashburn / LosAngeles / Washington              Ashburn

  src-bandwidth rules (Mbps): <=300 step 50, 300~1000 step 100, 1000~5000 step 500

  # Create an overseas service (Asia Pacific, HongKong cleaning center)
  ucloud uddos overseas service create --charge-type Month --quantity 1 \
    --area-line HongKong --src-bandwidth 100 --name my-service

  # Create an overseas service (Europe, Frankfurt cleaning center)
  ucloud uddos overseas service create --charge-type Month --quantity 1 \
    --area-line Frankfurt --src-bandwidth 100 --name my-service`,
		Run: func(cmd *cobra.Command, args []string) {
			if chargeType != "Month" && chargeType != "Year" {
				ctx.HandleError(fmt.Errorf(`invalid --charge-type %q, must be "Month" or "Year"`, chargeType))
				return
			}

			cleaningCenter, ok := overseasCityToCleaningCenter[areaLine]
			if !ok {
				ctx.HandleError(fmt.Errorf(
					"invalid --area-line %q; valid values: HongKong/Taipei/Singapore/Tokyo/Seoul/Bangkok/HoChiMinh/Jakarta/Manila/Mumbai/Frankfurt/London/Moscow/Ashburn/LosAngeles/Washington",
					areaLine,
				))
				return
			}

			switch {
			case srcBandwidth < 50:
				ctx.HandleError(fmt.Errorf("--src-bandwidth minimum is 50 for overseas"))
				return
			case srcBandwidth > 5000:
				ctx.HandleError(fmt.Errorf("--src-bandwidth maximum is 5000 for overseas"))
				return
			case srcBandwidth <= 300 && srcBandwidth%50 != 0:
				ctx.HandleError(fmt.Errorf("--src-bandwidth must be a multiple of 50 when <= 300 (overseas), got %d", srcBandwidth))
				return
			case srcBandwidth > 300 && srcBandwidth <= 1000 && srcBandwidth%100 != 0:
				ctx.HandleError(fmt.Errorf("--src-bandwidth must be a multiple of 100 when 300~1000 (overseas), got %d", srcBandwidth))
				return
			case srcBandwidth > 1000 && srcBandwidth <= 5000 && srcBandwidth%500 != 0:
				ctx.HandleError(fmt.Errorf("--src-bandwidth must be a multiple of 500 when 1000~5000 (overseas), got %d", srcBandwidth))
				return
			}

			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":                     "BuyHighProtectGameService",
				"ChargeType":                 chargeType,
				"Quantity":                   quantity,
				"LineType":                   "BGP",
				"SrcBandwidth":               srcBandwidth,
				"EngineRoom":                 []string{cleaningCenter},
				"AreaLine":                   areaLine,
				"ForwardType":                "Passthrough",
				"DefenceDDosBaseFlowArr":     []int{50},
				"DefenceDDosMaxFlowArr":      []int{50},
				"HighProtectGameServiceName": name,
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("BuyHighProtectGameService: %w", err))
				return
			}
			payload := resp.GetPayload()
			resInfo, _ := payload["ResourceInfo"].(map[string]interface{})
			resourceID := strVal(resInfo, "ResourceId")
			if resourceID == "" {
				ctx.HandleError(fmt.Errorf("BuyHighProtectGameService returned no ResourceId; the purchase may have failed, check the console"))
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "overseas DDoS service created: %s\n", resourceID)
			if !async {
				ctx.PollerTo(ctx.ProgressWriter(), describeOverseasService(ctx)).Spoll(
					resourceID,
					fmt.Sprintf("service[%s] is initializing", resourceID),
					[]string{napServiceStatusStarted},
				)
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: resourceID, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&chargeType, "charge-type", "", `Required. Billing type: "Month" or "Year"`)
	flags.IntVar(&quantity, "quantity", 0, "Required. Billing duration")
	flags.StringVar(&areaLine, "area-line", "", "Required. AreaLine: HongKong/Frankfurt/Ashburn and their coverage cities (see --help)")
	flags.IntVar(&srcBandwidth, "src-bandwidth", 0, "Required. Source bandwidth (Mbps). <=300 step 50, 300~1000 step 100, 1000~5000 step 500")
	flags.StringVar(&name, "name", "", "Required. Service name")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the service to become available.")
	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("quantity")
	cmd.MarkFlagRequired("area-line")
	cmd.MarkFlagRequired("src-bandwidth")
	cmd.MarkFlagRequired("name")
	return cmd
}

// describeOverseasService 返回 poller 用的服务状态查询函数，
// 调用 DescribeNapServiceInfo（NapType=2）按 ResourceId 查询，返回带 Status 字段的结构体。
func describeOverseasService(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, _ *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uaccount.NewClient)
		req := client.NewGenericRequest()
		if err := req.SetPayload(map[string]interface{}{
			"Action":     "DescribeNapServiceInfo",
			"NapType":    2,
			"ResourceId": id,
			"Offset":     0,
			"Limit":      1,
		}); err != nil {
			return nil, fmt.Errorf("set payload: %w", err)
		}
		resp, err := client.GenericInvoke(req)
		if err != nil {
			return nil, fmt.Errorf("DescribeNapServiceInfo: %w", err)
		}
		serviceInfo, _ := resp.GetPayload()["ServiceInfo"].([]interface{})
		if len(serviceInfo) == 0 {
			return nil, nil // 尚未可见，poller 视为 pending 继续轮询
		}
		m, _ := serviceInfo[0].(map[string]interface{})
		return &serviceStatusRow{Status: strVal(m, "DefenceStatus")}, nil
	}
}
