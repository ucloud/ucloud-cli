// Package ip ...
//
// @Brief  创建国内高防IP命令
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

// newCreate 构建 uddos mainland ip create 命令
//
// @Brief  构建国内高防 ip create 子命令，调用 CreateBGPServiceIP
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
		Short:   "Create a mainland BGP high-protection IP",
		Long:    "Create a new BGP DDoS protection IP for the specified mainland service",
		Example: "  ucloud uddos mainland ip create --resource-id ghp-xxxxx",
		Run: func(cmd *cobra.Command, args []string) {
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":     "CreateBGPServiceIP",
				"ResourceId": resourceID,
				"TypeIP":     typeIP,
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
