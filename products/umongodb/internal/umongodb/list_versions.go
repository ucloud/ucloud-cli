package umongodb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListVersions implements `umongodb list-versions`.
func newListVersions(ctx *cli.Context) *cobra.Command {
	var common request.CommonBase

	type versionRow struct {
		Version    string
		EngineType string
	}

	cmd := &cobra.Command{
		Use:   "list-versions",
		Short: "List available MongoDB versions",
		Long:  "List MongoDB versions supported in the current region/zone.",
		Run: func(c *cobra.Command, args []string) {
			params := map[string]interface{}{
				"Action": "ListUMongoDBVersion",
				"Region": common.GetRegion(),
				"Zone":   common.GetZone(),
			}
			if projectID := common.GetProjectId(); projectID != "" {
				params["ProjectId"] = projectID
			}

			payload, err := genericCall(ctx, "ListUMongoDBVersion", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			dataSet, ok := payload["DataSet"].([]interface{})
			if !ok {
				// No versions — return empty list, not an error.
				ctx.PrintList([]versionRow{})
				return
			}

			// Get default version
			defaultVer := ""
			if dv, ok := payload["DefaultDBVersion"].(map[string]interface{}); ok {
				defaultVer, _ = dv["DBVersion"].(string)
			}

			rows := make([]versionRow, 0, len(dataSet))
			for _, item := range dataSet {
				v, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				ver, _ := v["DBVersion"].(string)
				eng, _ := v["EngineType"].(string)
				if ver != "" {
					if ver == defaultVer {
						eng += " (default)"
					}
					rows = append(rows, versionRow{Version: ver, EngineType: eng})
				}
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	return cmd
}
