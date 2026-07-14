package pgsql

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// supabaseCreateFlags holds the shared business flags for CreateUSupabase and
// CreateUMemoryDB (the two requests have identical business fields).
type supabaseCreateFlags struct {
	common                                                                    *supabaseCommon
	instanceName, dashboardName, dashboardPassword, upgsqlUserName, dbVersion string
	upgsqlPassword, machineType, subnetID, vpcID, instanceMode, chargeType    string
	usupabasePort, paramGroupID, upgsqlPort, diskSpace, quantity              int
	labels                                                                    []string
}

func bindSupabaseCreate(cmd *cobra.Command, ctx *cli.Context) *supabaseCreateFlags {
	f := &supabaseCreateFlags{}
	f.common = bindSupabaseCommon(cmd, ctx)
	flags := cmd.Flags()
	flags.StringVar(&f.instanceName, "name", "", "Required. Supabase instance name")
	flags.StringVar(&f.dashboardName, "dashboard-name", "", "Required. Dashboard user name")
	flags.StringVar(&f.dashboardPassword, "dashboard-password", "", "Required. Dashboard password")
	flags.IntVar(&f.usupabasePort, "supabase-port", 8000, "Optional. Supabase service port, default 8000")
	flags.StringVar(&f.upgsqlUserName, "pgsql-user", "", "Required. UPgSQL user name")
	flags.StringVar(&f.dbVersion, "db-version", "", "Required. UPgSQL version, e.g. postgresql-13.4")
	flags.IntVar(&f.paramGroupID, "param-group-id", 0, "Required. UPgSQL param group ID")
	flags.StringVar(&f.upgsqlPassword, "pgsql-password", "", "Required. UPgSQL password")
	flags.IntVar(&f.upgsqlPort, "pgsql-port", 5432, "Optional. UPgSQL port, default 5432")
	flags.IntVar(&f.diskSpace, "disk-size-gb", 0, "Required. Disk space (GiB)")
	flags.StringVar(&f.machineType, "machine-type", "", "Required. Machine type, e.g. o.pgsql2m.medium. See 'ucloud pgsql db list-machine-type'")
	flags.StringVar(&f.subnetID, "subnet-id", "", "Required. Subnet ID")
	flags.StringVar(&f.vpcID, "vpc-id", "", "Required. VPC ID")
	flags.StringVar(&f.instanceMode, "mode", "Normal", "Optional. Normal / HA")
	flags.StringVar(&f.chargeType, "charge-type", "Month", "Optional. Year / Month / Dynamic")
	flags.IntVar(&f.quantity, "quantity", 1, "Optional. Purchase duration")
	flags.StringSliceVar(&f.labels, "label", nil, "Optional. Resource label key=value (repeatable)")
	command.SetFlagValues(cmd, "db-version", pgsqlVersionList...)
	command.SetFlagValues(cmd, "mode", "Normal", "HA")
	command.SetFlagValues(cmd, "charge-type", "Year", "Month", "Dynamic")
	return f
}

func (f *supabaseCreateFlags) params() map[string]interface{} {
	p := f.common.params()
	p["InstanceName"] = f.instanceName
	p["DashboardName"] = f.dashboardName
	p["DashboardPassword"] = f.dashboardPassword
	p["USupabasePort"] = f.usupabasePort
	p["UPgSQLUserName"] = f.upgsqlUserName
	p["DBVersion"] = f.dbVersion
	p["ParamGroupID"] = f.paramGroupID
	p["UPgSQLPassword"] = f.upgsqlPassword
	p["UPgSQLPort"] = f.upgsqlPort
	p["DiskSpace"] = f.diskSpace
	p["MachineType"] = f.machineType
	p["SubnetID"] = f.subnetID
	p["VPCID"] = f.vpcID
	p["InstanceMode"] = f.instanceMode
	p["ChargeType"] = f.chargeType
	p["Quantity"] = f.quantity
	labels := []map[string]interface{}{}
	for _, l := range f.labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) == 2 {
			labels = append(labels, map[string]interface{}{"Key": parts[0], "Value": parts[1]})
		}
	}
	p["Labels"] = labels
	return p
}

// runSupabaseCreate is shared by create and create-memory-db: invoke the action,
// poll the returned InstanceID to Running, emit the result.
func runSupabaseCreate(ctx *cli.Context, action string, f *supabaseCreateFlags, async bool) {
	params := f.params()
	payload, err := invokeSupabase(ctx, action, params)
	if err != nil {
		ctx.HandleError(err)
		return
	}
	instanceID := getString(payload, "InstanceID")
	if instanceID == "" {
		ctx.HandleError(fmt.Errorf("empty InstanceID in response"))
		return
	}
	w := ctx.ProgressWriter()
	if async {
		fmt.Fprintf(w, "supabase[%s] is initializing\n", instanceID)
	} else {
		text := fmt.Sprintf("supabase[%s] is initializing", instanceID)
		ctx.PollerTo(w, describeSupabaseByID(ctx, f.common.region, f.common.zone, f.common.projectID, f.common.memoryDB)).
			Spoll(instanceID, text, []string{SUPABASE_RUNNING, SUPABASE_FAIL})
	}
	ctx.EmitResult(cli.OpResultRow{ResourceID: instanceID, Action: "create", Status: "Initializing"})
}

// newSupabaseCreate ucloud pgsql supabase create
func newSupabaseCreate(ctx *cli.Context) *cobra.Command {
	var async bool
	var f *supabaseCreateFlags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a USupabase instance",
		Long:  "Create a USupabase instance (deploys a Supabase stack onto a UPgSQL host)",
		Run: func(c *cobra.Command, args []string) {
			runSupabaseCreate(ctx, "CreateUSupabase", f, async)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	f = bindSupabaseCreate(cmd, ctx)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish")

	for _, req := range []string{"name", "dashboard-name", "dashboard-password", "pgsql-user", "db-version", "param-group-id", "pgsql-password", "disk-size-gb", "machine-type", "subnet-id", "vpc-id"} {
		cmd.MarkFlagRequired(req)
	}
	return cmd
}
