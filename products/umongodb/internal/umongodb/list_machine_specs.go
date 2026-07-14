package umongodb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListMachineSpecs implements `umongodb list-machine-specs`.
func newListMachineSpecs(ctx *cli.Context) *cobra.Command {
	var common request.CommonBase
	var classTypeFilter string

	type specRow struct {
		MachineTypeId string
		Description   string
		Cpu           int
		MemoryGB      int
		ClassType     string
		DiskTypes     string
	}

	cmd := &cobra.Command{
		Use:   "list-machine-specs",
		Short: "List MongoDB machine specifications",
		Long:  "List available MongoDB machine types grouped by class type, including supported disk types.",
		Run: func(c *cobra.Command, args []string) {
			params := map[string]interface{}{
				"Action": "ListUMongoDBMachineSpec",
				"Region": common.GetRegion(),
				"Zone":   common.GetZone(),
			}
			if projectID := common.GetProjectId(); projectID != "" {
				params["ProjectId"] = projectID
			}
			if classTypeFilter != "" {
				params["ClassType"] = classTypeFilter
			}

			payload, err := genericCall(ctx, "ListUMongoDBMachineSpec", params)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			dataSet, ok := payload["DataSet"].([]interface{})
			if !ok {
				// No machine specs — return empty list, not an error.
				ctx.PrintList([]specRow{})
				return
			}

			rows := make([]specRow, 0)
			for _, item := range dataSet {
				spec, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				classType, _ := spec["ClassType"].(string)

				// Flatten disk types
				diskTypes := ""
				if dt, ok := spec["DiskType"].([]interface{}); ok {
					parts := make([]string, 0, len(dt))
					for _, d := range dt {
						if s, ok := d.(string); ok {
							parts = append(parts, s)
						}
					}
					diskTypes = fmt.Sprintf("[%s]", strings.Join(parts, ", "))
				}

				// Flatten compute types
				if ct, ok := spec["ComputeType"].([]interface{}); ok {
					for _, c := range ct {
						m, ok := c.(map[string]interface{})
						if !ok {
							continue
						}
						id, _ := m["MachineTypeId"].(string)
						desc, _ := m["Description"].(string)
						var cpu, mem int
						if v, ok := m["Cpu"].(float64); ok {
							cpu = int(v)
						}
						if v, ok := m["Memory"].(float64); ok {
							mem = int(v)
						}
						rows = append(rows, specRow{
							MachineTypeId: id,
							Description:   desc,
							Cpu:           cpu,
							MemoryGB:      mem,
							ClassType:     classType,
							DiskTypes:     diskTypes,
						})
					}
				}
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&classTypeFilter, "class-type", "", "Optional. Filter by class type: O or N.")

	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	return cmd
}

