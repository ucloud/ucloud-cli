package cmd

import (
	"reflect"

	"github.com/spf13/pflag"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
)

func bindRegion(req request.Common, flags *pflag.FlagSet) {
	var region string
	flags.StringVar(&region, "region", base.ConfigIns.Region, "Optional. Override default region, see 'ucloud region'")
	flags.SetFlagValuesFunc("region", getRegionList)
	req.SetRegionRef(&region)
}

func bindRegionS(region *string, flags *pflag.FlagSet) {
	*region = base.ConfigIns.Region
	flags.StringVar(region, "region", base.ConfigIns.Region, "Optional. Override default region, see 'ucloud region'")
	flags.SetFlagValuesFunc("region", getRegionList)
}

func bindZone(req request.Common, flags *pflag.FlagSet) {
	var zone string
	flags.StringVar(&zone, "zone", base.ConfigIns.Zone, "Optional. Override default available zone, see 'ucloud region'")
	flags.SetFlagValuesFunc("zone", func() []string {
		return getZoneList(req.GetRegion())
	})
	req.SetZoneRef(&zone)
}

func bindProjectID(req request.Common, flags *pflag.FlagSet) {
	var project string
	flags.StringVar(&project, "project-id", base.ConfigIns.ProjectID, "Optional. Override default project-id, see 'ucloud project list'")
	flags.SetFlagValuesFunc("project-id", getProjectList)
	req.SetProjectIdRef(&project)
}

func bindProjectIDS(project *string, flags *pflag.FlagSet) {
	*project = base.ConfigIns.ProjectID
	flags.StringVar(project, "project-id", base.ConfigIns.ProjectID, "Optional. Override default project-id, see 'ucloud project list'")
	flags.SetFlagValuesFunc("project-id", getProjectList)
}

func bindGroup(req interface{}, flags *pflag.FlagSet) {
	group := flags.String("group", "", "Optional. Business group")
	v := reflect.ValueOf(req).Elem()
	f := v.FieldByName("Tag")
	f.Set(reflect.ValueOf(group))
}

func bindLimit(req interface{}, flags *pflag.FlagSet) {
	limit := flags.Int("limit", 100, "Optional. The maximum number of resources per page")
	v := reflect.ValueOf(req).Elem()
	f := v.FieldByName("Limit")
	f.Set(reflect.ValueOf(limit))
}

func bindOffset(req interface{}, flags *pflag.FlagSet) {
	limit := flags.Int("offset", 0, "Optional. The index(a number) of resource which start to list")
	v := reflect.ValueOf(req).Elem()
	f := v.FieldByName("Limit")
	f.Set(reflect.ValueOf(limit))
}
