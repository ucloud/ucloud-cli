package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newPromoteSlave ucloud udb promote-slave
func newPromoteSlave(ctx *cli.Context) *cobra.Command {
	var ids []string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewPromoteUDBSlaveRequest()
	cmd := &cobra.Command{
		Use:   "promote-slave",
		Short: "Promote slave db to master",
		Long:  "Promote slave db to master",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			// loop aborts on first error (legacy); defer still emits partial results
			defer func() { ctx.EmitResult(results...) }()
			for _, id := range ids {
				req.DBId = sdk.String(id)
				_, err := client.PromoteUDBSlave(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(w, "udb[%s] was promoted\n", *req.DBId)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "promote-slave", Status: "Promoted"})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, "udb-id", nil, "Required. Resource ID of slave db to promote")
	req.IsForce = flags.Bool("is-force", false, "Optional. Force to promote slave db or not. If the slave db falls behind, the force promote may lose some data")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("udb-id")

	return cmd
}
