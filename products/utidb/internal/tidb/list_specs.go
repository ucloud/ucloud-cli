package tidb

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newListSpecs ucloud utidb list-specs
func newListSpecs(ctx *cli.Context) *cobra.Command {
	var nodeTypes string

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewGetTiDBClusterUhostSpecsRequest()

	cmd := &cobra.Command{
		Use:   "list-specs",
		Short: "List available uhost specs",
		Long:  helpListSpecsLong,
		Run: func(c *cobra.Command, args []string) {
			types := strings.Split(nodeTypes, ",")
			for i := range types {
				types[i] = strings.TrimSpace(types[i])
			}
			specs, err := getTiDBClusterUhostSpecs(ctx, req.GetRegion(), req.GetZone(), req.GetProjectId(), types)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []specRow{}
			for _, s := range specs {
				rows = append(rows, newSpecRowFromData(s))
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&nodeTypes, "node-types", "", "Required. Comma-separated node types: tidb, tikv, pd, tiflash")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("node-types")
	command.SetCompletion(cmd, "node-types", func() []string {
		return listNodeTypes(ctx, req.GetRegion(), req.GetZone())
	})

	return cmd
}
