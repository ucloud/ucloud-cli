// Package mainland ...
//
// @Brief  国内高防命令组聚合
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
package mainland

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	mainlandip "github.com/ucloud/ucloud-cli/products/uddos/internal/mainland/ip"
	mainlandrule "github.com/ucloud/ucloud-cli/products/uddos/internal/mainland/rule"
	mainlandsvc "github.com/ucloud/ucloud-cli/products/uddos/internal/mainland/service"
)

// NewCommand 构建 uddos mainland 命令组
//
// @Brief  构建国内高防命令组并挂载 service、ip 子命令
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
		Use:   "mainland",
		Short: "Manage mainland China DDoS high-protection services",
		Long:  "Manage UCloud mainland China (国内高防) DDoS protection services and IPs",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(mainlandip.NewCommand(ctx))
	cmd.AddCommand(mainlandrule.NewCommand(ctx))
	cmd.AddCommand(mainlandsvc.NewCommand(ctx))
	return cmd
}
