package cmd

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/ux"
)

func bindRegion(req request.Common, cmd *cobra.Command) {
	var region string
	cmd.Flags().StringVar(&region, "region", base.ConfigIns.Region, "Optional. Override default region for this command invocation, see 'ucloud region'")
	command.SetCompletion(cmd, "region", getRegionList)
	req.SetRegionRef(&region)
}

func bindRegionS(region *string, cmd *cobra.Command) {
	*region = base.ConfigIns.Region
	cmd.Flags().StringVar(region, "region", base.ConfigIns.Region, "Optional. Override default region for this command invocation, see 'ucloud region'")
	command.SetCompletion(cmd, "region", getRegionList)
}

func bindZone(req request.Common, cmd *cobra.Command) {
	var zone string
	cmd.Flags().StringVar(&zone, "zone", base.ConfigIns.Zone, "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(req.GetRegion())
	})
	req.SetZoneRef(&zone)
}

func bindZoneEmpty(req request.Common, cmd *cobra.Command) {
	var zone string
	cmd.Flags().StringVar(&zone, "zone", "", "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(req.GetRegion())
	})
	req.SetZoneRef(&zone)
}

func bindZoneEmptyS(zone, region *string, cmd *cobra.Command) {
	cmd.Flags().StringVar(zone, "zone", "", "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(*region)
	})
}

func bindZoneS(zone, region *string, cmd *cobra.Command) {
	*zone = base.ConfigIns.Zone
	cmd.Flags().StringVar(zone, "zone", base.ConfigIns.Zone, "Optional. Override default availability zone for this command invocation, see 'ucloud region'")
	command.SetCompletion(cmd, "zone", func() []string {
		return getZoneList(*region)
	})
}

func bindProjectID(req request.Common, cmd *cobra.Command) {
	var project string
	cmd.Flags().StringVar(&project, "project-id", base.ConfigIns.ProjectID, "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	command.SetCompletion(cmd, "project-id", getProjectList)
	req.SetProjectIdRef(&project)
}

func bindProjectIDS(project *string, cmd *cobra.Command) {
	*project = base.ConfigIns.ProjectID
	cmd.Flags().StringVar(project, "project-id", base.ConfigIns.ProjectID, "Optional. Override default project-id for this command invocation, see 'ucloud project list'")
	command.SetCompletion(cmd, "project-id", getProjectList)
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
	offset := flags.Int("offset", 0, "Optional. The index(a number) of resource which start to list")
	v := reflect.ValueOf(req).Elem()
	f := v.FieldByName("Offset")
	f.Set(reflect.ValueOf(offset))
}

func bindChargeType(req interface{}, cmd *cobra.Command) {
	chargeType := cmd.Flags().String("charge-type", "Month", "Optional. Enumeration value.'Year',pay yearly;'Month',pay monthly; 'Dynamic', pay hourly; 'Trial', free trial(need permission)")
	v := reflect.ValueOf(req).Elem()
	f := v.FieldByName("ChargeType")
	f.Set(reflect.ValueOf(chargeType))
	command.SetFlagValues(cmd, "charge-type", "Month", "Dynamic", "Year")
}

func bindQuantity(req interface{}, flags *pflag.FlagSet) {
	quanitiy := flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	v := reflect.ValueOf(req).Elem()
	f := v.FieldByName("Quantity")
	f.Set(reflect.ValueOf(quanitiy))
}

type concurrentAction struct {
	reqs       []request.Common
	actionFunc func(request.Common) (bool, []string)
	wg         *sync.WaitGroup
	result     chan bool
	tokens     chan bool
}

func newConcurrentAction(reqs []request.Common, limit int, actionFunc func(request.Common) (bool, []string)) *concurrentAction {
	if limit <= 0 {
		limit = 10
	}
	return &concurrentAction{
		reqs:       reqs,
		actionFunc: actionFunc,
		wg:         &sync.WaitGroup{},
		result:     make(chan bool),
		tokens:     make(chan bool, limit), //控制并发量，最多是个并发
	}
}

func (c *concurrentAction) actionFuncWrapper(req request.Common) {
	c.tokens <- true
	success, logs := c.actionFunc(req)
	c.result <- success
	logs = append([]string{"========================================"}, logs...)
	base.LogInfo(logs...)
	<-c.tokens
	time.Sleep(time.Second / 5)
	c.wg.Done()
}

func (c *concurrentAction) Do() {
	count := len(c.reqs)
	success, fail := 0, 0
	refresh := ux.NewRefresh()
	//同时执行任务数量大于5时，不再单独显示每一个任务的进行情况，而是聚合显示
	if count > 5 {
		ux.Doc.Disable()
		refresh.Do(fmt.Sprintf("total:%d, doing:%d, success:%d, fail:%d", count, len(c.tokens), success, fail))
	}
	go func() {
		for {
			select {
			case ret := <-c.result:
				if ret {
					success++
				} else {
					fail++
				}

			case <-time.Tick(time.Second / 30):
				if count == (success+fail) && fail > 0 {
					fmt.Printf("Check logs in %s\n", base.GetLogFilePath())
					return
				}
				if count > 5 {
					refresh.Do(fmt.Sprintf("total:%d, doing:%d, success:%d, fail:%d", count, len(c.tokens), success, fail))
				}
			}
		}
	}()

	for _, req := range c.reqs {
		c.wg.Add(1)
		go c.actionFuncWrapper(req)
	}

	c.wg.Wait()
}
