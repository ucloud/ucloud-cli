// Package ip ...
//
// @Brief  创建海外高防IP命令
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
package ip

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newCreate 构建 uddos overseas ip create 命令
//
// @Brief  构建海外高防 ip create 子命令，调用 CreateBGPServiceIP
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
	var resourceID, typeIP, remark, tag string

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create an overseas BGP high-protection IP",
		Long:    "Create a new BGP DDoS protection IP for the specified overseas service",
		Example: "  ucloud uddos overseas ip create --resource-id nap-xxxxx",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)

			resolvedEIPRegion, err := lookupEIPRegion(client, resourceID)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			params := map[string]interface{}{
				"Action":     "CreateBGPServiceIP",
				"ResourceId": resourceID,
				"TypeIP":     typeIP,
				"EIPRegion":  resolvedEIPRegion,
			}
			if cmd.Flags().Changed("remark") {
				params["Remark"] = remark
			}
			if cmd.Flags().Changed("tag") {
				params["Tag"] = tag
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			resp, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("CreateBGPServiceIP: %w", err))
				return
			}
			payload := resp.GetPayload()
			defenceIP, _ := payload["DefenceIP"].(string)
			fmt.Fprintf(ctx.ProgressWriter(), "BGP IP created: %s\n", defenceIP)
			ctx.EmitResult(cli.OpResultRow{ResourceID: defenceIP, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&typeIP, "type-ip", "TypeFree", "Optional. IP type: TypeFree or TypeCharge, default TypeFree")
	flags.StringVar(&remark, "remark", "", "Optional. Remark for this IP")
	flags.StringVar(&tag, "tag", "", "Optional. Business group tag")
	cmd.MarkFlagRequired("resource-id")
	return cmd
}

// lookupEIPRegion resolves the EIPRegion for an overseas service by querying
// DescribeHighProtectGameServiceInfo then GetNapServiceConfig.
func lookupEIPRegion(client *uaccount.UAccountClient, resourceID string) (string, error) {
	// Step 1: fetch service details to get EngineRoom and LineType
	svcReq := client.NewGenericRequest()
	if err := svcReq.SetPayload(map[string]interface{}{
		"Action":     "DescribeNapServiceInfo",
		"ResourceId": resourceID,
		"NapType":    2, // APAC / overseas
		"Limit":      1,
	}); err != nil {
		return "", fmt.Errorf("DescribeNapServiceInfo set payload: %w", err)
	}
	svcResp, err := client.GenericInvoke(svcReq)
	if err != nil {
		return "", fmt.Errorf("DescribeNapServiceInfo: %w", err)
	}
	svcPayload := svcResp.GetPayload()
	serviceInfo, _ := svcPayload["ServiceInfo"].([]interface{})
	if len(serviceInfo) == 0 {
		return "", fmt.Errorf("service %s not found", resourceID)
	}
	svc, ok := serviceInfo[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected service info format")
	}
	// EngineRoom is returned as []interface{} (comma-split array); take the first element.
	engineRoom := ""
	if rooms, ok := svc["EngineRoom"].([]interface{}); ok && len(rooms) > 0 {
		engineRoom, _ = rooms[0].(string)
	}
	lineType := strVal(svc, "LineType")
	areaLine := strVal(svc, "AreaLine")

	// Step 2: fetch service config to get IpInfo region list
	cfgReq := client.NewGenericRequest()
	if err := cfgReq.SetPayload(map[string]interface{}{
		"Action":     "GetNapServiceConfig",
		"AreaLine":   areaLine,
		"EngineRoom": engineRoom,
		"LineType":   lineType,
	}); err != nil {
		return "", fmt.Errorf("GetNapServiceConfig set payload: %w", err)
	}
	cfgResp, err := client.GenericInvoke(cfgReq)
	if err != nil {
		return "", fmt.Errorf("GetNapServiceConfig: %w", err)
	}
	cfgPayload := cfgResp.GetPayload()
	configs, _ := cfgPayload["NapServiceConfig"].([]interface{})
	if len(configs) == 0 {
		return "", fmt.Errorf("no service config found for areaLine=%s engineRoom=%s lineType=%s", areaLine, engineRoom, lineType)
	}
	cfg, ok := configs[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected service config format")
	}
	ipInfoList, _ := cfg["IpInfo"].([]interface{})
	if len(ipInfoList) == 0 {
		return "", fmt.Errorf("IpInfo is empty in service config")
	}
	first, ok := ipInfoList[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected IpInfo entry format")
	}
	region := strVal(first, "Region")
	if region == "" {
		return "", fmt.Errorf("IpInfo Region is empty")
	}
	return region, nil
}
