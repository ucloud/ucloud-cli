package umongodb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDeleteSharded implements `umongodb delete-sharded`.
func newDeleteSharded(ctx *cli.Context) *cobra.Command {
	return newDeleteCmd(ctx, deleteOpts{
		use:        "delete-sharded",
		short:      "Delete MongoDB sharded cluster instances",
		long:       "Delete one or more MongoDB sharded cluster instances. The cluster is stopped before deletion unless --skip-stop is set.",
		action:     "DeleteUMongoDBShardedCluster",
		idParam:    "ShardedClusterId",
		idFlagDesc: "Cluster ID(s) of sharded cluster instances to delete.",
	})
}
