package redis

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListBlock returns ucloud redis list-block.
func newListBlock(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDescribeUMemBlockInfoRequest()
	cmd := &cobra.Command{
		Use:   "list-block",
		Short: "List block info of distributed redis",
		Long:  "List block info of distributed redis",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeUMemBlockInfo(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []BlockRow{}
			for _, b := range resp.DataSet {
				row := BlockRow{
					BlockID:    b.BlockId,
					BlockName:  b.BlockName,
					BlockVip:   b.BlockVip,
					BlockPort:  b.BlockPort,
					BlockType:  b.BlockType,
					BlockState: b.BlockState,
					BlockSize:  b.BlockSize,
					UsedSize:   b.BlockUsedSize,
					SlotBegin:  b.BlockSlotBegin,
					SlotEnd:    b.BlockSlotEnd,
					ReadWeight: b.BlockReadWeight,
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.SpaceId = flags.String("umem-id", "", "Required. Resource ID of the distributed redis")
	req.Limit = sdk.Int(100)
	req.Offset = sdk.Int(0)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("umem-id")
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}
