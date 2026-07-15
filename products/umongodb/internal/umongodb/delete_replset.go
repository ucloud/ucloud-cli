package umongodb

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDeleteReplset implements `umongodb delete-replset`.
func newDeleteReplset(ctx *cli.Context) *cobra.Command {
	return newDeleteCmd(ctx, deleteOpts{
		use:        "delete-replset",
		short:      "Delete MongoDB replica set instances",
		long:       "Delete one or more MongoDB replica set instances. The cluster is stopped before deletion unless --skip-stop is set.",
		action:     "DeleteUMongoDBReplSet",
		idParam:    "ClusterId",
		idFlagDesc: "Cluster ID(s) of replica set instances to delete.",
	})
}
