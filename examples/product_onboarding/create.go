package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// versionValues is a static candidate set for the create --version flag,
// registered via command.SetFlagValues.
var versionValues = []string{"mysql-5.7", "mysql-8.0"}

// newCreate implements `example create`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams,
// ctx.PollerTo(...).Spoll (the wait path), ctx.ProgressWriter, ctx.EmitResult,
// ctx.HandleError, command.SetFlagValues, MarkFlagRequired with "Required."
// descriptions, the --async pattern.
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewCreateUDBInstanceRequest()

	var async bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an example instance",
		Long:  "Create an example instance and, unless --async is set, wait for it to become Running.",
		Run: func(c *cobra.Command, args []string) {
			// Human narration goes to the progress writer: stdout in table
			// mode, stderr in machine (json/yaml) modes so stdout stays
			// machine-parseable.
			w := ctx.ProgressWriter()
			resp, err := client.CreateUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			text := fmt.Sprintf("%s[%s] is creating", productName, resp.DBId)
			if async {
				// --async: narrate and return without polling.
				fmt.Fprintln(w, text)
			} else {
				// Synchronous: poll until the instance reaches a terminal state.
				ctx.PollerTo(w, describeByID(ctx)).Spoll(resp.DBId, text, []string{stateRunning, stateFail})
			}
			// Machine (json/yaml) modes: emit the structured result row on
			// stdout. In table mode EmitResult is a no-op — the narration
			// above is the result.
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.DBId, Action: "create", Status: "Creating"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Required flags: description starts with "Required." AND MarkFlagRequired.
	req.Name = flags.String("name", "", "Required. Instance name, at least 6 characters.")
	req.AdminPassword = flags.String("password", "", "Required. Admin password.")
	req.DBTypeId = flags.String("version", "", "Required. DB version, e.g. mysql-8.0.")

	// Optional flags: description starts with "Optional.".
	req.Port = flags.Int("port", 3306, "Optional. Service port.")
	req.DiskSpace = flags.Int("disk-size-gb", 20, "Optional. Disk size in GiB.")
	req.MemoryLimit = flags.Int("memory-size-mb", 1000, "Optional. Memory size in MB.")
	req.ParamGroupId = flags.Int("param-group-id", 0, "Optional. Parameter group ID.")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for creation to finish.")

	// Aggregate binder also wires --charge-type/--quantity here because the
	// create request carries those fields.
	ctx.BindCommonParams(cmd, req)

	// Static candidate set for an enum flag.
	command.SetFlagValues(cmd, "version", versionValues...)

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("version")

	return cmd
}
