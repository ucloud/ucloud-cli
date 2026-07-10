package token

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newGet ucloud urocketmq token get
func newGet(ctx *cli.Context) *cobra.Command {
	var display bool
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewGetURocketMQTokenRequest()
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get URocketMQ token details",
		Long:  "Get URocketMQ token details. AKSK secret key is only shown with --display for security.",
		Run: func(cmd *cobra.Command, args []string) {
			if display {
				req.Display = sdk.String("true")
			}
			resp, err := client.GetURocketMQToken(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			renderTokenGet(ctx, &resp.Token, display)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID. See 'ucloud urocketmq service list'")
	req.TokenId = cmd.Flags().String("token-id", "", "Required. Token ID")
	cmd.Flags().BoolVar(&display, "display", false, "Optional. Display AKSK secret key in plaintext")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "token-id", func() []string {
		return TokenList(ctx, *req.ProjectId, *req.Region, *req.ServiceId)
	})
	cmd.MarkFlagRequired("service-id")
	cmd.MarkFlagRequired("token-id")

	return cmd
}

// renderTokenGet selects the row struct by format + display state.
// Security: table mode always uses tokenRowDefault (no AKSK); json/yaml includes SecretKey only with --display.
func renderTokenGet(ctx *cli.Context, t *urocketmq.Token, showSecret bool) {
	if ctx.Format() != cli.OutputTable {
		if showSecret {
			ctx.PrintList([]tokenRowWithSecret{{
				TokenId:          t.TokenId,
				Name:             t.Name,
				TopicConsumePerm: t.TopicConsumePerm,
				TopicProducePerm: t.TopicProducePerm,
				Type:             t.Type,
				CreateTime:       common.FormatDate(t.CreateTime),
				ModifyTime:       common.FormatDate(t.ModifyTime),
				AccessKey:        t.AKSK.AccessKey,
				SecretKey:        t.AKSK.SecretKey,
			}})
			return
		}
		ctx.PrintList([]tokenRow{{
			TokenId:          t.TokenId,
			Name:             t.Name,
			TopicConsumePerm: t.TopicConsumePerm,
			TopicProducePerm: t.TopicProducePerm,
			Type:             t.Type,
			CreateTime:       common.FormatDate(t.CreateTime),
			ModifyTime:       common.FormatDate(t.ModifyTime),
			AccessKey:        t.AKSK.AccessKey,
		}})
		return
	}

	// table mode never outputs SecretKey
	ctx.PrintList([]tokenRowDefault{{
		TokenId:          t.TokenId,
		Name:             t.Name,
		TopicConsumePerm: t.TopicConsumePerm,
		TopicProducePerm: t.TopicProducePerm,
		Type:             t.Type,
		CreateTime:       common.FormatDate(t.CreateTime),
	}})
}
