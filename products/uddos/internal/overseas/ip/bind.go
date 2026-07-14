// Package ip ...
//
// @Brief  绑定海外高防IP命令
//
// @File   bind.go
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

// newBind 构建 uddos overseas ip bind 命令
//
// @Brief  构建海外高防 ip bind 子命令，调用 BindNapIP（仅支持海外高防）
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
func newBind(ctx *cli.Context) *cobra.Command {
	var eipID, resourceType, resourceID, bindResourceID, napIP string

	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "Bind a defence IP to a resource",
		Long:    "Bind an overseas DDoS protection IP to a cloud resource (EIP binding)",
		Example: "  ucloud uddos overseas ip bind --eip-id eip-xxxxx --resource-id nap-xxxxx --resource-type uhost --bind-resource-id uhost-xxxxx --nap-ip 1.2.3.4",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			if err := checkOverseasService(client, resourceID); err != nil {
				ctx.HandleError(err)
				return
			}
			params := map[string]interface{}{
				"Action":         "BindNapIP",
				"EIPId":          eipID,
				"ResourceType":   resourceType,
				"ResourceId":     resourceID,
				"BindResourceId": bindResourceID,
				"NapIp":          napIP,
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			_, err := client.GenericInvoke(req)
			if err != nil {
				ctx.HandleError(fmt.Errorf("BindNapIP: %w", err))
				return
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: napIP, Action: "bind", Status: "Bound"})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&eipID, "eip-id", "", "Required. EIP resource ID")
	flags.StringVar(&resourceType, "resource-type", "", "Required. Resource type (e.g. uhost)")
	flags.StringVar(&resourceID, "resource-id", "", "Required. High-protection service resource ID")
	flags.StringVar(&bindResourceID, "bind-resource-id", "", "Required. Resource ID to bind the IP to")
	flags.StringVar(&napIP, "nap-ip", "", "Required. DDoS protection IP address")
	cmd.MarkFlagRequired("eip-id")
	cmd.MarkFlagRequired("resource-type")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("bind-resource-id")
	cmd.MarkFlagRequired("nap-ip")
	return cmd
}
