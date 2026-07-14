// Package overseas ...
//
// @Brief  海外高防命令组聚合
//
// @File   cmd.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
//
// @CopyRights(C) UCloud All rights reserved.
package overseas

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	overseasip "github.com/ucloud/ucloud-cli/products/uddos/internal/overseas/ip"
	overseassvc "github.com/ucloud/ucloud-cli/products/uddos/internal/overseas/service"
)

// NewCommand 构建 uddos overseas 命令组
//
// @Brief  构建海外高防命令组并挂载 service、ip、rule 子命令
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
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overseas",
		Short: "Manage overseas DDoS high-protection services",
		Long:  "Manage UCloud overseas (海外高防) DDoS protection services and IPs",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(overseasip.NewCommand(ctx))
	cmd.AddCommand(overseassvc.NewCommand(ctx))
	return cmd
}
