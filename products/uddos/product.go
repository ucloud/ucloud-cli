// Package uddos ...
//
// @Brief  UDDoS 高防产品 CLI 入口
//
// @File   product.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/09
//
// @CopyRights(C) UCloud All rights reserved.
package uddos

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/products/uddos/internal/mainland"
	"github.com/ucloud/ucloud-cli/products/uddos/internal/overseas"
)

type product struct{}

// New returns the uddos product (registered via hack/gen-products).
//
// @Brief  创建 uddos 产品实例
//
// @Param
//
// @Return cli.Product
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/09
func New() cli.Product { return product{} }

// Metadata returns the product metadata.
//
// @Brief  返回产品元数据
//
// @Param
//
// @Return cli.Metadata
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/09
func (product) Metadata() cli.Metadata {
	return cli.Metadata{Name: "uddos", Commands: []string{"uddos"}}
}

// NewCommand builds the uddos root command and mounts subcommand groups.
//
// @Brief  构建 uddos 根命令及子命令组
//
// @Param  ctx *cli.Context
//
// @Return []*cobra.Command
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/09
func (product) NewCommand(ctx *cli.Context) []*cobra.Command {
	cmd := &cobra.Command{
		Use:   "uddos",
		Short: "Manage UCloud DDoS protection services",
		Long:  "Manage UCloud DDoS high-protection services for mainland China and overseas",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(mainland.NewCommand(ctx))
	cmd.AddCommand(overseas.NewCommand(ctx))
	return []*cobra.Command{cmd}
}
