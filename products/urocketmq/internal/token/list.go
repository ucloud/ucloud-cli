package token

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/urocketmq/internal/service"
)

// newList ucloud urocketmq token list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, urocketmq.NewClient)
	req := client.NewListURocketMQTokenRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List URocketMQ tokens",
		Long:  "List URocketMQ tokens",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.ListURocketMQToken(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			listToken(ctx, resp.TokenList)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.ServiceId = cmd.Flags().String("service-id", "", "Required. Service ID")
	req.Limit = cmd.Flags().Int("limit", 20, "Optional. Limit default 20, max value 100")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")

	command.SetCompletion(cmd, "service-id", func() []string {
		return service.ServiceList(ctx, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("service-id")

	return cmd
}

// listToken renders the token list. json/yaml emits full-field tokenRow; table mode uses tokenRowDefault.
func listToken(ctx *cli.Context, tokens []urocketmq.TokenDetail) {
	list := make([]tokenRow, 0, len(tokens))
	for _, t := range tokens {
		list = append(list, tokenRow{
			TokenId:          t.TokenId,
			Name:             t.Name,
			TopicConsumePerm: t.TopicConsumePerm,
			TopicProducePerm: t.TopicProducePerm,
			Type:             t.Type,
			CreateTime:       common.FormatDate(t.CreateTime),
			ModifyTime:       common.FormatDate(t.ModifyTime),
			AccessKey:        t.AKSK.AccessKey,
		})
	}

	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	rows := make([]tokenRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, tokenRowDefault{
			TokenId:          r.TokenId,
			Name:             r.Name,
			TopicConsumePerm: r.TopicConsumePerm,
			TopicProducePerm: r.TopicProducePerm,
			Type:             r.Type,
			CreateTime:       r.CreateTime,
		})
	}
	ctx.PrintList(rows)
}
