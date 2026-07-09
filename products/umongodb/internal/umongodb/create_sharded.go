package umongodb

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newCreateSharded implements `umongodb create-sharded`.
func newCreateSharded(ctx *cli.Context) *cobra.Command {

	var name, password, version string
	var shardCount, nodeCount, diskSpaceGB int
	var machineTypeID string
	var mongosNodeCount int
	var mongosMachineTypeID string
	var port int
	var templateID string
	var vpcID, subnetID, tag string
	var chargeType string
	var quantity int

	client := cli.NewServiceClient(ctx, umongodb.NewClient)
	req := client.NewCreateUMongoDBShardedClusterRequest()

	cmd := &cobra.Command{
		Use:   "create-sharded",
		Short: "Create a MongoDB sharded cluster",
		Long:  "Create a MongoDB sharded cluster asynchronously. Use 'umongodb list' to check creation status.",
		Run: func(c *cobra.Command, args []string) {
			// Auto-default template ID if not specified
			if templateID == "" {
				id, err := getDefaultTemplateID(ctx, version, "SharedCluster", *req.ProjectId, *req.Region)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				templateID = id
			}

			req.Name = &name
			req.AdminPassword = &password
			req.DBVersion = &version
			req.ShardCount = &shardCount
			req.NodeCount = &nodeCount
			req.DiskSpace = &diskSpaceGB
			req.MachineTypeId = &machineTypeID
			req.TemplateId = &templateID

			// Optional params — only set when explicitly changed
			if c.Flags().Changed("mongos-node-count") {
				req.MongosNodeCount = &mongosNodeCount
			}
			if c.Flags().Changed("mongos-machine-type-id") {
				req.MongosMachineTypeId = &mongosMachineTypeID
			}
			if c.Flags().Changed("port") {
				req.ListenPort = &port
			}
			if c.Flags().Changed("vpc-id") {
				req.VPCId = &vpcID
			}
			if c.Flags().Changed("subnet-id") {
				req.SubnetId = &subnetID
			}
			if c.Flags().Changed("tag") {
				req.Tag = &tag
			}
			if c.Flags().Changed("charge-type") {
				req.ChargeType = &chargeType
			}
			if c.Flags().Changed("quantity") {
				req.Quantity = &quantity
			}

			// MongoDB creation is slow; use 5-minute timeout
			req.WithTimeout(5 * time.Minute)

			_, err := client.CreateUMongoDBShardedCluster(req)
			if err != nil {
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
	flags.IntVar(&shardCount, "shard-count", 0, "Required. Number of shards.")
	flags.IntVar(&nodeCount, "node-count", 0, "Required. Number of nodes per shard.")
	flags.IntVar(&diskSpaceGB, "disk-space-gb", 0, "Required. Data node disk space in GB (20-32000, multiples of 10).")
	flags.StringVar(&machineTypeID, "machine-type-id", "", "Required. Data node machine type ID, e.g. o.mongo2m.medium.")

	// Optional flags
	flags.IntVar(&mongosNodeCount, "mongos-node-count", 0, "Optional. Mongos node count.")
	flags.StringVar(&mongosMachineTypeID, "mongos-machine-type-id", "", "Optional. Mongos node machine type ID.")
	flags.IntVar(&port, "port", 27017, "Optional. Service port.")
	flags.StringVar(&templateID, "template-id", "", "Optional. Config template ID. Auto-fetched if omitted.")
	flags.StringVar(&vpcID, "vpc-id", "", "Optional. VPC ID. See 'ucloud vpc list'.")
	flags.StringVar(&subnetID, "subnet-id", "", "Optional. Subnet ID. See 'ucloud subnet list'.")
	flags.StringVar(&tag, "tag", "", "Optional. Business group name.")
	flags.StringVar(&chargeType, "charge-type", "Month", "Optional. Charge type: Year / Month / Dynamic / Trial.")
	flags.IntVar(&quantity, "quantity", 1, "Optional. Purchase duration in months.")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	// Completions
	command.SetCompletion(cmd, "version", func() []string {
		return getMongoDBVersionList(ctx, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "machine-type-id", func() []string {
		return getMongoDBMachineSpecList(ctx, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "template-id", func() []string {
		return getMongoDBTemplateList(ctx, version, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, vpcID, *req.ProjectId, *req.Region)
	})
	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Dynamic", "Trial")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("version")
	cmd.MarkFlagRequired("shard-count")
	cmd.MarkFlagRequired("node-count")
	cmd.MarkFlagRequired("disk-space-gb")
	cmd.MarkFlagRequired("machine-type-id")

	return cmd
}
