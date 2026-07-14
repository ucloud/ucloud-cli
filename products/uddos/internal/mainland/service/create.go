// Package service ...
//
// @Brief  创建国内高防服务命令
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

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// mainlandAreaLineEngineRooms 国内 area-line → 合法 engine-room 列表
var mainlandAreaLineEngineRooms = map[string][]string{
	"EastChina":  {"Zaozhuang", "Yangzhou"},
	"NorthChina": {"Shijiazhuang"},
}

// newCreate 构建 uddos mainland service create 命令
//
// @Brief  构建国内高防 service create 子命令，调用 BuyHighProtectGameService（国内参数）
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
	var chargeType, areaLine, engineRoom, name string
	var quantity, srcBandwidth, defenceBaseFlow, defenceMaxFlow int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a mainland DDoS high-protection service",
		Long:  "Create a mainland China DDoS high-protection service via BuyHighProtectGameService (ForwardType=Proxy).",
		Example: `  Area line / engine room pairings:
  --area-line      --engine-room
  EastChina        Zaozhuang / Yangzhou
  NorthChina       Shijiazhuang

  Defence flow valid values (Gbps): 30/40/50/60/70/80/100/200/300/400/500/600/700/800
    --defence-max-flow must be >= --defence-base-flow

  # Create a mainland service (East China, Zaozhuang)
  ucloud uddos mainland service create --charge-type Month --quantity 1 \
    --area-line EastChina --engine-room Zaozhuang --src-bandwidth 100 \
    --defence-base-flow 30 --defence-max-flow 50 --name my-service

  # Create a mainland service (North China, Shijiazhuang)
  ucloud uddos mainland service create --charge-type Month --quantity 1 \
    --area-line NorthChina --engine-room Shijiazhuang --src-bandwidth 100 \
    --defence-base-flow 30 --defence-max-flow 30 --name my-service`,
		Run: func(cmd *cobra.Command, args []string) {
			if chargeType != "Month" && chargeType != "Year" {
				ctx.HandleError(fmt.Errorf(`invalid --charge-type %q, must be "Month" or "Year"`, chargeType))
				return
			}

			validRooms, isValidAreaLine := mainlandAreaLineEngineRooms[areaLine]
			if !isValidAreaLine {
				ctx.HandleError(fmt.Errorf(`invalid --area-line %q, must be "EastChina" or "NorthChina"`, areaLine))
				return
			}
			validSet := make(map[string]bool, len(validRooms))
			for _, r := range validRooms {
				validSet[r] = true
			}
			if !validSet[engineRoom] {
				desc := ""
				for i, r := range validRooms {
					if i > 0 {
						desc += "/"
					}
					desc += r
				}
				ctx.HandleError(fmt.Errorf(
					"invalid --engine-room %q for --area-line %q, valid values: %s",
					engineRoom, areaLine, desc,
				))
				return
			}

			if srcBandwidth < 50 {
				ctx.HandleError(fmt.Errorf("--src-bandwidth minimum is 50 for mainland, got %d", srcBandwidth))
				return
			}
			if srcBandwidth%10 != 0 {
				ctx.HandleError(fmt.Errorf("--src-bandwidth must be a multiple of 10 for mainland, got %d", srcBandwidth))
				return
			}

			validFlows := map[int]bool{
				30: true, 40: true, 50: true, 60: true, 70: true, 80: true,
				100: true, 200: true, 300: true, 400: true, 500: true,
				600: true, 700: true, 800: true,
			}
			const validFlowDesc = "30/40/50/60/70/80/100/200/300/400/500/600/700/800"
			if !validFlows[defenceBaseFlow] {
				ctx.HandleError(fmt.Errorf("invalid --defence-base-flow %d, must be one of: %s", defenceBaseFlow, validFlowDesc))
				return
			}
			if !validFlows[defenceMaxFlow] {
				ctx.HandleError(fmt.Errorf("invalid --defence-max-flow %d, must be one of: %s", defenceMaxFlow, validFlowDesc))
				return
			}
			if defenceMaxFlow < defenceBaseFlow {
				ctx.HandleError(fmt.Errorf("--defence-max-flow (%d) must be >= --defence-base-flow (%d)", defenceMaxFlow, defenceBaseFlow))
				return
			}

			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":                     "BuyHighProtectGameService",
				"ChargeType":                 chargeType,
				"Quantity":                   quantity,
				"LineType":                   "BGP",
				"SrcBandwidth":               srcBandwidth,
				"EngineRoom":                 []string{engineRoom},
				"AreaLine":                   areaLine,
				"ForwardType":                "Proxy",
				"DefenceDDosBaseFlowArr":     []int{defenceBaseFlow},
				"DefenceDDosMaxFlowArr":      []int{defenceMaxFlow},
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
			ctx.EmitResult(cli.OpResultRow{ResourceID: resourceID, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&chargeType, "charge-type", "", `Required. Billing type: "Month" or "Year"`)
	flags.IntVar(&quantity, "quantity", 0, "Required. Billing duration")
	flags.StringVar(&areaLine, "area-line", "", `Required. "EastChina" or "NorthChina"`)
	flags.StringVar(&engineRoom, "engine-room", "", "Required. Zaozhuang/Yangzhou (EastChina) or Shijiazhuang (NorthChina)")
	flags.IntVar(&srcBandwidth, "src-bandwidth", 0, "Required. Source bandwidth (Mbps), min 50, multiple of 10")
	flags.IntVar(&defenceBaseFlow, "defence-base-flow", 0, "Required. Base defence flow (Gbps): 30/40/50/60/70/80/100/200/300/400/500/600/700/800")
	flags.IntVar(&defenceMaxFlow, "defence-max-flow", 0, "Required. Max defence flow (Gbps), must be >= defence-base-flow")
	flags.StringVar(&name, "name", "", "Required. Service name")
	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("quantity")
	cmd.MarkFlagRequired("area-line")
	cmd.MarkFlagRequired("engine-room")
	cmd.MarkFlagRequired("src-bandwidth")
	cmd.MarkFlagRequired("defence-base-flow")
	cmd.MarkFlagRequired("defence-max-flow")
	cmd.MarkFlagRequired("name")
	return cmd
}
