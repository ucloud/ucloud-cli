package onboarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// Terminal states a Poller waits on. A real product imports these from its own
// status table; the example defines them locally so it depends only on the
// platform packages and the SDK, never on another product's internals.
const (
	stateRunning = "Running"
	stateShutoff = "Shutoff"
	stateFail    = "Fail"
)

// versionValues is a static candidate set for the create --version flag,
// registered via command.SetFlagValues.
var versionValues = []string{"mysql-5.7", "mysql-8.0"}

// newList implements `example list`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams (the
// aggregate binder), ctx.PickResourceID, ctx.PrintList, ctx.HandleError,
// command.SetCompletion, command.SetFlagValues.
func newList(ctx *cli.Context) *cobra.Command {
	// One authed SDK client per command, built from the constructor. The Run
	// func only needs this to type-check; the example is never executed.
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List example instances",
		Long:  "List example instances in the active region/zone/project.",
		Run: func(c *cobra.Command, args []string) {
			if req.DBId != nil && *req.DBId != "" {
				*req.DBId = ctx.PickResourceID(*req.DBId)
			}
			resp, err := client.DescribeUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]instanceRow, 0, len(resp.DataSet))
			for _, ins := range resp.DataSet {
				rows = append(rows, instanceRow{
					ResourceID: ins.DBId,
					Name:       ins.Name,
					Zone:       ins.Zone,
					Mode:       ins.InstanceMode,
					Spec:       fmt.Sprintf("%s|%dMB|%dGB", ins.DBTypeId, ins.MemoryLimit, ins.DiskSpace),
					Status:     ins.State,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	// Optional resource-id filter, named after the product.
	req.DBId = flags.String(resourceIDFlag, "", "Optional. List only the specified instance.")
	req.ClassType = sdk.String("sql")

	// One call binds region/zone/project plus --limit/--offset (present on this
	// request) with the per-invocation defaults and the injected completion
	// providers. This is the primary, preferred binder.
	ctx.BindCommonParams(cmd, req)

	// Dynamic completion for the resource-id flag.
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}

// newDescribe implements `example describe`.
//
// Platform APIs exercised: cli.NewServiceClient, the non-aggregate binders
// (ctx.BindRegion / ctx.BindZone / ctx.BindProjectID — shown here once for the
// case where you want per-field control), cli.DescribeRow for detail rows,
// ctx.PrintList, ctx.PickResourceID, ctx.HandleError, command.SetCompletion.
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Show details of one example instance",
		Long:  "Show the full attribute/value detail of a single example instance.",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = ctx.PickResourceID(*req.DBId)
			resp, err := client.DescribeUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.DataSet) == 0 {
				ctx.HandleError(fmt.Errorf("instance %q not found", *req.DBId))
				return
			}
			ins := resp.DataSet[0]

			// cli.DescribeRow renders a single resource as attribute/content
			// rows. In table mode the field names "Attribute" and "Content"
			// become the two column headers.
			rows := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: ins.DBId},
				{Attribute: "Name", Content: ins.Name},
				{Attribute: "Zone", Content: ins.Zone},
				{Attribute: "Mode", Content: ins.InstanceMode},
				{Attribute: "Version", Content: ins.DBTypeId},
				{Attribute: "Memory(MB)", Content: fmt.Sprintf("%d", ins.MemoryLimit)},
				{Attribute: "Disk(GB)", Content: fmt.Sprintf("%d", ins.DiskSpace)},
				{Attribute: "VirtualIP", Content: ins.VirtualIP},
				{Attribute: "Port", Content: fmt.Sprintf("%d", ins.Port)},
				{Attribute: "Status", Content: ins.State},
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBId = flags.String(resourceIDFlag, "", "Required. Resource ID of the instance to describe.")

	// Non-aggregate binding: bind each common flag explicitly. Equivalent to
	// BindCommonParams for region/zone/project, shown here for the case where a
	// command needs to bind them individually (e.g. to interleave custom flags).
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}

// newCreate implements `example create`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams,
// ctx.Poller(...).Spoll (the wait path), ctx.Out, ctx.HandleError,
// command.SetFlagValues, MarkFlagRequired with "Required." descriptions, the
// --async pattern.
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewCreateUDBInstanceRequest()

	var async bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an example instance",
		Long:  "Create an example instance and, unless --async is set, wait for it to become Running.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.CreateUDBInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if async {
				// --async: return immediately without polling.
				fmt.Fprintf(ctx.Out(), "%s[%s] is creating\n", productName, resp.DBId)
				return
			}
			// Synchronous: poll until the instance reaches a terminal state.
			text := fmt.Sprintf("%s[%s] is creating", productName, resp.DBId)
			ctx.Poller(describeByID(ctx)).Spoll(resp.DBId, text, []string{stateRunning, stateFail})
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

// newDelete implements `example delete`.
//
// Platform APIs exercised: cli.NewServiceClient, ctx.BindCommonParams,
// ctx.Confirm (the destructive-op guard), the --yes/-y pattern,
// ctx.PickResourceID, ctx.Out, ctx.HandleError, command.SetCompletion.
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBInstanceRequest()

	var ids []string
	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete example instances",
		Long:  "Delete one or more example instances by resource ID.",
		Run: func(c *cobra.Command, args []string) {
			// Destructive: gate on confirmation unless --yes was passed.
			if !ctx.Confirm(yes, "Are you sure you want to delete the instance(s)?") {
				return
			}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.DeleteUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.Out(), "%s[%s] deleted\n", productName, id)
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to delete.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}

// newStart implements `example start`.
func newStart(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewStartUDBInstanceRequest()

	var ids []string
	var async bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start example instances",
		Long:  "Start one or more stopped example instances.",
		Run: func(c *cobra.Command, args []string) {
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.StartUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is starting", productName, id)
				if async {
					fmt.Fprintln(ctx.Out(), text)
					continue
				}
				ctx.Poller(describeByID(ctx)).Spoll(id, text, []string{stateRunning, stateFail})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to start.")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		// Only stopped instances are startable.
		return listResourceIDs(ctx, []string{stateShutoff}, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}

// newStop implements `example stop`.
func newStop(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewStopUDBInstanceRequest()

	var ids []string
	var async bool

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop example instances",
		Long:  "Stop one or more running example instances.",
		Run: func(c *cobra.Command, args []string) {
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.StopUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is stopping", productName, id)
				if async {
					fmt.Fprintln(ctx.Out(), text)
					continue
				}
				ctx.Poller(describeByID(ctx)).Spoll(id, text, []string{stateShutoff, stateFail})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to stop.")
	req.ForceToKill = flags.Bool("force", false, "Optional. Force-stop the instance(s).")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetFlagValues(cmd, "force", "true", "false")
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		// Only running instances are stoppable.
		return listResourceIDs(ctx, []string{stateRunning}, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}

// newRestart implements `example restart`.
func newRestart(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewRestartUDBInstanceRequest()

	var ids []string
	var async bool

	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart example instances",
		Long:  "Restart one or more example instances.",
		Run: func(c *cobra.Command, args []string) {
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)
				req.DBId = sdk.String(id)
				if _, err := client.RestartUDBInstance(req); err != nil {
					ctx.HandleError(err)
					continue
				}
				text := fmt.Sprintf("%s[%s] is restarting", productName, id)
				if async {
					fmt.Fprintln(ctx.Out(), text)
					continue
				}
				ctx.Poller(describeByID(ctx)).Spoll(id, text, []string{stateRunning, stateFail})
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. Resource ID(s) of instances to restart.")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	ctx.BindCommonParams(cmd, req)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return listResourceIDs(ctx, nil, derefStr(req.Region), derefStr(req.Zone), derefStr(req.ProjectId))
	})

	return cmd
}

// describeByID returns the Poller describe func: given a resource id it fetches
// the current resource so the Poller can read its state field. The signature
// (func(string, *request.CommonBase) (interface{}, error)) is exactly what
// ctx.Poller expects.
func describeByID(ctx *cli.Context) func(string, *request.CommonBase) (interface{}, error) {
	return func(id string, common *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, udb.NewClient)
		req := client.NewDescribeUDBInstanceRequest()
		if common != nil {
			req.CommonBase = *common
		}
		req.DBId = sdk.String(id)
		resp, err := client.DescribeUDBInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) == 0 {
			return nil, fmt.Errorf("instance %q not found", id)
		}
		// Return a *struct whose exported State field the Poller reads by name.
		return &resp.DataSet[0], nil
	}
}

// derefStr safely dereferences a *string bound by a flag, returning "" for nil.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
