package umongodb

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreateReplset implements `umongodb create-replset`.
// Uses GenericInvoke because the typed CreateUMongoDBReplSetRequest
// lacks the TemplateId field required by the API.
func newCreateReplset(ctx *cli.Context) *cobra.Command {
	var common request.CommonBase

	var name, password, version string
	var diskSpaceGB, nodeCount int
	var machineTypeID string
	var port int
	var templateID string
	var vpcID, subnetID, tag string
	var chargeType string
	var quantity int

	cmd := &cobra.Command{
		Use:   "create-replset",
		Short: "Create a MongoDB replica set",
		Long:  "Create a MongoDB replica set asynchronously. Use 'umongodb list' to check creation status.",
		Run: func(c *cobra.Command, args []string) {
			region := common.GetRegion()
			zone := common.GetZone()
			projectID := common.GetProjectId()

			// Auto-default template ID if not specified
			if templateID == "" {
				id, err := getDefaultTemplateID(ctx, version, "ReplicaSet", projectID, region)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				templateID = id
			}

			params := map[string]interface{}{
				"Action":        "CreateUMongoDBReplSet",
				"Region":        region,
				"Zone":          zone,
				"Name":          name,
				"AdminPassword": password,
				"DBVersion":     version,
				"DiskSpace":     diskSpaceGB,
				"MachineTypeId": machineTypeID,
				"NodeCount":     nodeCount,
				"TemplateId":    templateID,
			}
			if projectID != "" {
				params["ProjectId"] = projectID
			}

			// Optional params — only set when explicitly changed
			if c.Flags().Changed("port") {
				params["ListenPort"] = port
			}
			if c.Flags().Changed("vpc-id") {
				params["VPCId"] = vpcID
			}
			if c.Flags().Changed("subnet-id") {
				params["SubnetId"] = subnetID
			}
			if c.Flags().Changed("tag") {
				params["Tag"] = tag
			}
			if c.Flags().Changed("charge-type") {
				params["ChargeType"] = chargeType
			}
			if c.Flags().Changed("quantity") {
				params["Quantity"] = quantity
			}

			if _, err := genericCall(ctx, "CreateUMongoDBReplSet", params); err != nil {
				ctx.HandleError(err)
				return
			}

			w := ctx.ProgressWriter()
			fmt.Fprintln(w, fmt.Sprintf("%s is creating (use 'umongodb list' to check status)", name))
			ctx.EmitResult(cli.OpResultRow{ResourceID: name, Action: "create", Status: "Creating"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Required flags
	flags.StringVar(&name, "name", "", "Required. Instance name, at least 6 characters.")
	flags.StringVar(&password, "password", "", "Required. Admin password.")
	flags.StringVar(&version, "version", "", "Required. MongoDB version, e.g. \"MongoDB 6.0\".")
	flags.IntVar(&diskSpaceGB, "disk-space-gb", 0, "Required. Disk space in GB (20-32000, multiples of 10).")
	flags.StringVar(&machineTypeID, "machine-type-id", "", "Required. Machine type ID, e.g. o.mongo2m.medium.")
	flags.IntVar(&nodeCount, "node-count", 3, "Node count (3, 5, or 7).")

	// Optional flags
	flags.IntVar(&port, "port", 27017, "Optional. Service port.")
	flags.StringVar(&templateID, "template-id", "", "Optional. Config template ID. Auto-fetched if omitted.")
	flags.StringVar(&vpcID, "vpc-id", "", "Optional. VPC ID. See 'ucloud vpc list'.")
	flags.StringVar(&subnetID, "subnet-id", "", "Optional. Subnet ID. See 'ucloud subnet list'.")
	flags.StringVar(&tag, "tag", "", "Optional. Business group name.")
	flags.StringVar(&chargeType, "charge-type", "Month", "Optional. Charge type: Year / Month / Dynamic / Trial.")
	flags.IntVar(&quantity, "quantity", 1, "Optional. Purchase duration in months.")

	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	// Completions
	command.SetCompletion(cmd, "version", func() []string {
		return getMongoDBVersionList(ctx, common.GetRegion(), common.GetZone())
	})
	command.SetCompletion(cmd, "machine-type-id", func() []string {
		return getMongoDBMachineSpecList(ctx, common.GetRegion(), common.GetZone())
	})
	command.SetCompletion(cmd, "template-id", func() []string {
		return getMongoDBTemplateList(ctx, version, common.GetProjectId(), common.GetRegion())
	})
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, common.GetProjectId(), common.GetRegion())
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, vpcID, common.GetProjectId(), common.GetRegion())
	})
	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "node-count", "3", "5", "7")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("disk-space-gb")
	cmd.MarkFlagRequired("machine-type-id")

	return cmd
}
