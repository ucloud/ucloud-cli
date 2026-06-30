package command

import (
	"reflect"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// Defaults carries the per-invocation default region/zone/project for flag binding.
type Defaults struct {
	Region    string
	Zone      string
	ProjectID string
}

// BindRegion binds a --region flag whose value is shared with req via SetRegionRef.
func BindRegion(cmd *cobra.Command, req request.Common, def Defaults, regionList func() []string) {
	var region string
	cmd.Flags().StringVar(&region, "region", def.Region, "Optional. Override default region for this command invocation, see 'ucloud region'")
	SetCompletion(cmd, "region", regionList)
	req.SetRegionRef(&region)
}

// BindRegionS binds a --region flag into the caller-provided region pointer.
func BindRegionS(cmd *cobra.Command, region *string, def Defaults, regionList func() []string) {
	*region = def.Region
	cmd.Flags().StringVar(region, "region", def.Region, "Optional. Override default region for this command invocation, see 'ucloud region'")
	SetCompletion(cmd, "region", regionList)
}

// BindZone binds a --zone flag (default = def.Zone) whose completion is
// zoneList(req.GetRegion()) evaluated lazily.
func BindZone(cmd *cobra.Command, req request.Common, def Defaults, zoneList func(region string) []string) {
	var zone string
	cmd.Flags().StringVar(&zone, "zone", def.Zone, "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	SetCompletion(cmd, "zone", func() []string { return zoneList(req.GetRegion()) })
	req.SetZoneRef(&zone)
}

// BindZoneEmpty is like BindZone but the default is "" (matches cmd's bindZoneEmpty).
func BindZoneEmpty(cmd *cobra.Command, req request.Common, zoneList func(region string) []string) {
	var zone string
	cmd.Flags().StringVar(&zone, "zone", "", "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	SetCompletion(cmd, "zone", func() []string { return zoneList(req.GetRegion()) })
	req.SetZoneRef(&zone)
}

// BindProjectID binds a --project-id flag shared with req via SetProjectIdRef.
func BindProjectID(cmd *cobra.Command, req request.Common, def Defaults, projectList func() []string) {
	var project string
	cmd.Flags().StringVar(&project, "project-id", def.ProjectID, "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	SetCompletion(cmd, "project-id", projectList)
	req.SetProjectIdRef(&project)
}

// BindProjectIDS binds a --project-id flag into the caller-provided project pointer.
func BindProjectIDS(cmd *cobra.Command, project *string, def Defaults, projectList func() []string) {
	*project = def.ProjectID
	cmd.Flags().StringVar(project, "project-id", def.ProjectID, "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	SetCompletion(cmd, "project-id", projectList)
}

// BindLimit binds a --limit flag into req.Limit via reflection.
func BindLimit(cmd *cobra.Command, req interface{}) {
	limit := cmd.Flags().Int("limit", 100, "Optional. The maximum number of resources per page")
	reflect.ValueOf(req).Elem().FieldByName("Limit").Set(reflect.ValueOf(limit))
}

// BindOffset binds a --offset flag into req.Offset via reflection.
func BindOffset(cmd *cobra.Command, req interface{}) {
	offset := cmd.Flags().Int("offset", 0, "Optional. The index(a number) of resource which start to list")
	reflect.ValueOf(req).Elem().FieldByName("Offset").Set(reflect.ValueOf(offset))
}

// BindChargeType binds a --charge-type flag into req.ChargeType via reflection.
func BindChargeType(cmd *cobra.Command, req interface{}) {
	chargeType := cmd.Flags().String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly; 'Dynamic', pay hourly; 'Trial', free trial(need permission)")
	reflect.ValueOf(req).Elem().FieldByName("ChargeType").Set(reflect.ValueOf(chargeType))
	SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
}

// BindQuantity binds a --quantity flag into req.Quantity via reflection.
func BindQuantity(cmd *cobra.Command, req interface{}) {
	quantity := cmd.Flags().Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	reflect.ValueOf(req).Elem().FieldByName("Quantity").Set(reflect.ValueOf(quantity))
}

// BindGroup binds a --group flag into req.Tag via reflection (verbatim from
// cmd/util.go bindGroup; req must have a settable `Tag *string` field).
func BindGroup(cmd *cobra.Command, req interface{}) {
	group := cmd.Flags().String("group", "", "Optional. Business group")
	reflect.ValueOf(req).Elem().FieldByName("Tag").Set(reflect.ValueOf(group))
}

// hasField reports whether req (a pointer to a struct) has a settable field
// with the given name. It is used to guard optional reflection-bound flags so
// that a req lacking the field is simply skipped instead of panicking.
func hasField(req interface{}, name string) bool {
	v := reflect.ValueOf(req)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return false
	}
	f := v.FieldByName(name)
	return f.IsValid() && f.CanSet()
}

// BindCommonParams binds all common flags onto cmd in a single call.
//
// It always binds region/zone/project when req satisfies request.Common,
// reusing the presence-safe per-field binders. It then binds the optional
// list/charge flags (--limit/--offset/--charge-type/--quantity) ONLY for the
// fields that actually exist on req, so a request lacking them does not panic.
func BindCommonParams(
	cmd *cobra.Command,
	req interface{},
	def Defaults,
	regionList func() []string,
	zoneList func(region string) []string,
	projectList func() []string,
) {
	if common, ok := req.(request.Common); ok {
		BindRegion(cmd, common, def, regionList)
		BindZone(cmd, common, def, zoneList)
		BindProjectID(cmd, common, def, projectList)
	}

	if hasField(req, "Limit") {
		BindLimit(cmd, req)
	}
	if hasField(req, "Offset") {
		BindOffset(cmd, req)
	}
	if hasField(req, "ChargeType") {
		BindChargeType(cmd, req)
	}
	if hasField(req, "Quantity") {
		BindQuantity(cmd, req)
	}
}
