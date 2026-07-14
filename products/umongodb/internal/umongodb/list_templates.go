package umongodb

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListTemplates implements `umongodb list-templates`.
func newListTemplates(ctx *cli.Context) *cobra.Command {
	var versionFilter, clusterTypeFilter string

	type templateRow struct {
		TemplateId     string
		Name           string
		MongodbVersion string
		ClusterType    string
		TemplateType   string
	}

	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewListUMongoDBConfigTemplateRequest()

	cmd := &cobra.Command{
		Use:   "list-templates",
		Short: "List MongoDB config templates",
		Long:  "List MongoDB config templates for the current region.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.ListUMongoDBConfigTemplate(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			rows := make([]templateRow, 0, len(resp.DataSet))
			for _, t := range resp.DataSet {
				// Apply filters
				if versionFilter != "" && !strings.EqualFold(t.MongodbVersion, versionFilter) {
					continue
				}
				if clusterTypeFilter != "" && !strings.EqualFold(t.ClusterType, clusterTypeFilter) {
					continue
				}

				rows = append(rows, templateRow{
					TemplateId:     t.TemplateId,
					Name:           t.TemplateName,
					MongodbVersion: t.MongodbVersion,
					ClusterType:    t.ClusterType,
					TemplateType:   t.TemplateType,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&versionFilter, "version", "", "Optional. Filter by MongoDB version, e.g. \"MongoDB 6.0\".")
	flags.StringVar(&clusterTypeFilter, "cluster-type", "", "Optional. Filter by cluster type: ReplicaSet or SharedCluster.")

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	return cmd
}
