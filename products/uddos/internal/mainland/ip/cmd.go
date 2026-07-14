// Package ip ...
//
// @Brief  国内高防IP管理命令聚合
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
package ip

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// NewCommand 构建 uddos mainland ip 命令组
//
// @Brief  构建国内高防 ip 命令组并挂载子命令
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
		Use:   "ip",
		Short: "Manage mainland DDoS protection IPs",
		Long:  "List, create and delete mainland BGP DDoS protection IPs",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newDelete(ctx))
	return cmd
}

func strVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func intVal(m map[string]interface{}, key string) int {
	v, _ := m[key].(float64)
	return int(v)
}
