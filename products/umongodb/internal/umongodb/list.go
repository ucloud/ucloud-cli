package umongodb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList implements `umongodb list`.
func newList(ctx *cli.Context) *cobra.Command {
	var common request.CommonBase

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List MongoDB instances",
		Long:  "List MongoDB instances in the active region/zone/project.",
		Run: func(c *cobra.Command, args []string) {
			params := map[string]interface{}{
				"Action": "ListUMongoDBInstances",
				"Region": common.GetRegion(),
			}
			if zone := common.GetZone(); zone != "" {
				params["Zone"] = zone
			}
			if projectID := common.GetProjectId(); projectID != "" {
				params["ProjectId"] = projectID
			}

			payload, err := genericCall(ctx, "ListUMongoDBInstances", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			dataSet, ok := payload["DataSet"].([]interface{})
			if !ok {
				// No instances in this region — return empty list, not an error.
				ctx.PrintList([]instanceRow{})
				return
			}

			rows := make([]instanceRow, 0, len(dataSet))
			for _, item := range dataSet {
				ins, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				row := instanceRow{
					ResourceID: strVal(ins, "ClusterId"),
					Name:       strVal(ins, "Name"),
					Version:    strVal(ins, "DBVersion"),
					Status:     strVal(ins, "State"),
				}
				row.ClusterType = strVal(ins, "ClusterType")
				row.ConnectURL = strVal(ins, "ConnectURL")

				// DiskSpace
				if v, ok := ins["DiskSpace"].(float64); ok {
					row.DiskGB = int(v)
				}

				// Machine type from nested DataComputeType
				if dct, ok := ins["DataComputeType"].(map[string]interface{}); ok {
					row.MachineType = strVal(dct, "Description")
				}

				rows = append(rows, row)
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

// strVal extracts a string value from a map, returning "" if missing or wrong type.
func strVal(m map[string]interface{}, key string) string {
	v, ok := m[key].(string)
	if !ok {
		return ""
	}
	return v
}
