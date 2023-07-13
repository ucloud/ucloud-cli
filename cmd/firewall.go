// Copyright © 2018 NAME HERE tony.li@ucloud.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

// NewCmdFirewall  ucloud firewall
func NewCmdFirewall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "List and manipulate extranet firewall",
		Long:  `List and manipulate extranet firewall`,
		Args:  cobra.NoArgs,
	}
	writer := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdFirewallList(writer))
	cmd.AddCommand(NewCmdFirewallCreate(writer))
	cmd.AddCommand(NewCmdFirewallAddRule(writer))
	cmd.AddCommand(NewCmdFirewallDeleteRule(writer))
	cmd.AddCommand(NewCmdFirewallApply())
	cmd.AddCommand(NewCmdFirewallCopy())
	cmd.AddCommand(NewCmdFirewallDelete())
	cmd.AddCommand(NewCmdFirewallResource(writer))
	cmd.AddCommand(NewCmdFirewallUpdate(writer))

	return cmd
}

// FirewallRow 表格行
type FirewallRow struct {
	ResourceID          string
	FirewallName        string
	Rule                string
	Group               string
	RuleAmount          int
	BoundResourceAmount int
	CreationTime        string
}

// NewCmdFirewallList ucloud firewall list
func NewCmdFirewallList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeFirewallRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List extranet firewall",
		Long:  `List extranet firewall`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := base.BizClient.DescribeFirewall(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []FirewallRow{}
			for _, fw := range resp.DataSet {
				row := FirewallRow{}
				row.ResourceID = fw.FWId
				row.FirewallName = fw.Name
				row.Group = fw.Tag
				row.RuleAmount = len(fw.Rule)
				row.BoundResourceAmount = fw.ResourceCount
				row.CreationTime = base.FormatDate(fw.CreateTime)
				if fw.Remark != "" {
					row.FirewallName += "\nremark:" + fw.Remark + "\n"
				}
				for _, r := range fw.Rule {
					rule := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
					row.Rule += rule + "\n"
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.FWId = flags.String("firewall-id", "", "Optional. The Rsource ID of firewall. Return all firewalls by default.")
	req.ResourceType = flags.String("bound-resource-type", "", "Optional. The type of resource bound on the firewall")
	req.ResourceId = flags.String("bound-resource-id", "", "Optional. The resource ID of resource bound on the firewall")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	return cmd
}

func parseRulesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// NewCmdFirewallCreate ucloud firewall create
func NewCmdFirewallCreate(out io.Writer) *cobra.Command {
	var rulesFilePath string
	var rules []string

	req := base.BizClient.NewCreateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create firewall",
		Long:    "Create firewall",
		Example: `ucloud firewall create --name test3 --rules "TCP|22|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt`,
		Run: func(c *cobra.Command, args []string) {
			if rules == nil && rulesFilePath == "" {
				fmt.Fprintln(out, "Error: flags rules and rules-file can't be both empty")
				return
			}
			if rulesFilePath != "" {
				lines, err := parseRulesFromFile(rulesFilePath)
				if err != nil {
					base.HandleError(err)
					return
				}
				rules = append(rules, lines...)
			}
			req.Rule = rules
			resp, err := base.BizClient.CreateFirewall(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("firewall[%s] created\n", resp.FWId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&rules, "rules", nil, "Required if rules-file doesn't exist. Schema: Protocol|Port|IP|Action|Level. Prototol range 'TCP','UDP','ICMP' and 'GRE'; Port is a local port accessed by source address, port range [0-65535]; IP is the source address of the network packet that requests ucloud host resource, supporting IP address and network segment, such as '120.132.69.216' or '0.0.0.0/0'; Action is the processing behavior of the packet when the firewall is in effect, including 'ACCEPT' AND 'DROP'; Level, when a rule is added to a firewall, the rules take effect in order of level, which range 'HIGH','MEDIUM' and 'LOW'. For example, 'TCP|22|192.168.1.1/22|DROP|LOW'")
	flags.StringVar(&rulesFilePath, "rules-file", "", "Required if rules doesn't exist. Path of rules file, in which each rule occupies one line. Schema: Protocol|Port|IP|Action|Level.")
	req.Name = flags.String("name", "", "Required. Name of firewall to create")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Tag = flags.String("group", "", "Optional. Group of the firewall to create")
	req.Remark = flags.String("remark", "", "Optional. Remark of the firewall to create")
	cmd.MarkFlagRequired("name")
	flags.SetFlagValuesFunc("rules-file", func() []string {
		return base.GetFileList("")
	})
	return cmd
}

// NewCmdFirewallAddRule ucloud firewall add-rule
func NewCmdFirewallAddRule(out io.Writer) *cobra.Command {
	var rulesFilePath string
	var fwIDs []string
	req := base.BizClient.NewUpdateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "add-rule",
		Short:   "Add rule to firewall instance",
		Long:    "Add rule to firewall instance",
		Example: `ucloud firewall add-rule --fw-id firewall-2xxxxz/test.lxj2 --rules "TCP|24|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt`,
		Run: func(c *cobra.Command, args []string) {
			if req.Rule == nil && rulesFilePath == "" {
				fmt.Fprintln(out, "Error: flags rules and rules-file can't be both empty")
				return
			}
			for _, fwID := range fwIDs {
				id := base.PickResourceID(fwID)
				req.FWId = &id
				firewall, err := getFirewall(*req.FWId, *req.ProjectId, *req.Region)
				if err != nil {
					base.HandleError(err)
					return
				}
				ruleMap := map[string]bool{}
				for _, r := range firewall.Rule {
					ruleStr := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
					ruleMap[ruleStr] = true
				}
				if rulesFilePath != "" {
					rules, err := parseRulesFromFile(rulesFilePath)
					if err != nil {
						base.HandleError(err)
						return
					}
					req.Rule = append(req.Rule, rules...)
				}
				for _, r := range req.Rule {
					ruleMap[r] = true
				}
				req.Rule = []string{}
				for r := range ruleMap {
					r = strings.TrimSpace(r)
					req.Rule = append(req.Rule, r)
				}
				_, err = base.BizClient.UpdateFirewall(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("firewall[%s] updated\n", fwID)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls to update")
	flags.StringSliceVar(&req.Rule, "rules", nil, "Required if rules-file is empay. Rules to add to firewall. Schema:'Protocol|Port|IP|Action|Level'. See 'ucloud firewall create --help' for detail.")
	flags.StringVar(&rulesFilePath, "rules-file", "", "Required if rules is empty. Path of rules file, in which each rule occupies one line. Schema: Protocol|Port|IP|Action|Level.")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("fw-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})
	flags.SetFlagValuesFunc("rules-file", func() []string {
		return base.GetFileList("")
	})

	cmd.MarkFlagRequired("fw-id")
	return cmd
}

// NewCmdFirewallDeleteRule ucloud firewall remove-rule
func NewCmdFirewallDeleteRule(out io.Writer) *cobra.Command {
	var rulesFilePath string
	var fwIDs []string
	req := base.BizClient.NewUpdateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "remove-rule",
		Short:   "Remove rule from firewall instance",
		Long:    "Remove rule from firewall instance",
		Example: `ucloud firewall remove-rule --fw-id firewall-2cxxxz/test.lxj2 --rules "TCP|24|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt`,
		Run: func(c *cobra.Command, args []string) {
			if req.Rule == nil && rulesFilePath == "" {
				fmt.Fprintln(out, "Error: flags rules and rules-file can't be both empty")
				return
			}
			for _, fwID := range fwIDs {
				id := base.PickResourceID(fwID)
				req.FWId = &id
				firewall, err := getFirewall(*req.FWId, *req.ProjectId, *req.Region)
				if err != nil {
					base.HandleError(err)
					return
				}
				ruleMap := map[string]bool{}
				for _, r := range firewall.Rule {
					ruleStr := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
					ruleMap[ruleStr] = true
				}
				if rulesFilePath != "" {
					rules, err := parseRulesFromFile(rulesFilePath)
					if err != nil {
						base.HandleError(err)
						return
					}
					req.Rule = append(req.Rule, rules...)
				}
				for _, r := range req.Rule {
					r = strings.TrimSpace(r)
					delete(ruleMap, r)
				}
				req.Rule = []string{}
				for r := range ruleMap {
					req.Rule = append(req.Rule, r)
				}
				if len(req.Rule) == 0 {
					fmt.Fprintf(out, "Error: rules can't be all deleted\n")
					return
				}
				_, err = base.BizClient.UpdateFirewall(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "firewall[%s] updated\n", fwID)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls to update")
	flags.StringSliceVar(&req.Rule, "rules", nil, "Required if rules-file is empay. Rules to add to firewall. Schema:'Protocol|Port|IP|Action|Level'. See 'ucloud firewall create --help' for detail.")
	flags.StringVar(&rulesFilePath, "rules-file", "", "Required if rules is empty. Path of rules file, in which each rule occupies one line. Schema: Protocol|Port|IP|Action|Level.")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("fw-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")
	return cmd
}

// NewCmdFirewallApply ucloud firewall apply
func NewCmdFirewallApply() *cobra.Command {
	req := base.BizClient.NewGrantFirewallRequest()
	resourceIDs := []string{}
	fwID := ""
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Applay firewall to ucloud service",
		Long:    "Applay firewall to ucloud service",
		Example: "ucloud firewall apply --fw-id firewall-xxx --resource-id uhost-xxx --resource-type uhost",
		Run: func(c *cobra.Command, args []string) {
			req.FWId = sdk.String(base.PickResourceID(fwID))
			for _, id := range resourceIDs {
				req.ResourceId = sdk.String(id)
				_, err := base.BizClient.GrantFirewall(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				base.Cxt.Printf("firewall[%s] applied to %s[%s]\n", fwID, *req.ResourceType, id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&fwID, "fw-id", "", "Required. Resource ID of firewall to apply to some ucloud resource")
	req.ResourceType = flags.String("resource-type", "", "Required. Resource type of resource to be applied firewall. Range 'uhost','unatgw','upm','hadoophost','fortresshost','udhost','udockhost','dbaudit'.")
	flags.StringSliceVar(&resourceIDs, "resource-id", nil, "Resource ID of resources to be applied firewall")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValues("resource-type", "uhost", "unatgw", "upm", "hadoophost", "fortresshost", "udhost", "udockhost", "dbaudit")
	flags.SetFlagValuesFunc("fw-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("resource-type")

	return cmd
}

// NewCmdFirewallCopy ucloud firewall copy
func NewCmdFirewallCopy() *cobra.Command {
	srcFirewall := ""
	srcRegion := ""
	req := base.BizClient.NewCreateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "copy",
		Short:   "Copy firewall",
		Long:    "Copy firewall",
		Example: "ucloud firewall copy --src-fw firewall-xxx --target-region cn-bj2 --name test",
		Run: func(c *cobra.Command, args []string) {
			fwID := base.PickResourceID(srcFirewall)
			firewall, err := getFirewall(fwID, *req.ProjectId, srcRegion)

			if err != nil {
				base.HandleError(err)
				return
			}
			req.Tag = sdk.String(firewall.Tag)
			req.Remark = sdk.String(firewall.Remark)
			for _, r := range firewall.Rule {
				rstr := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
				req.Rule = append(req.Rule, rstr)
			}
			resp, err := base.BizClient.CreateFirewall(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			base.Cxt.Printf("firewall[%s] created from %s\n", resp.FWId, srcFirewall)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&srcFirewall, "src-fw", "", "Required. ResourceID or name of source firewall")
	req.Name = flags.String("name", "", "Required. Name of new firewall")
	flags.StringVar(&srcRegion, "region", base.ConfigIns.Region, "Optional. Current region, used to fetch source firewall")
	req.Region = flags.String("target-region", base.ConfigIns.Region, "Optional. Copy firewall to target region")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	flags.SetFlagValuesFunc("src-fw", func() []string {
		return getFirewallIDNames(*req.ProjectId, srcRegion)
	})
	flags.SetFlagValuesFunc("target-region", getRegionList)
	flags.SetFlagValuesFunc("region", getRegionList)

	cmd.MarkFlagRequired("src-fw-id")
	cmd.MarkFlagRequired("name")

	return cmd
}

// NewCmdFirewallDelete ucloud firewall delete
func NewCmdFirewallDelete() *cobra.Command {
	req := base.BizClient.NewDeleteFirewallRequest()
	ids := []string{}
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete firewall by resource ids or names",
		Long:    "Delete firewall by resource ids or names",
		Example: "ucloud firewall delete --fw-id firewall-xxx",
		Run: func(c *cobra.Command, args []string) {
			for _, id := range ids {
				req.FWId = sdk.String(base.PickResourceID(id))
				_, err := base.BizClient.DeleteFirewall(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				base.Cxt.Printf("firewall[%s] deleted\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "fw-id", nil, "Required. Resource IDs of firewall to delete")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("fw-id")
	flags.SetFlagValuesFunc("fw-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})

	return cmd
}

// FirewallResourceRow 表格行
type FirewallResourceRow struct {
	ResourceName string
	ResourceID   string
	ResourceType string
	IntranetIP   string
	Group        string
	Remark       string
}

// NewCmdFirewallResource ucloud firewall resource
func NewCmdFirewallResource(out io.Writer) *cobra.Command {
	fwID := ""
	req := base.BizClient.NewDescribeFirewallResourceRequest()
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "List resources that has been applied the firewall",
		Long:  "List resources that has been applied the firewall",
		Run: func(c *cobra.Command, args []string) {
			req.FWId = sdk.String(base.PickResourceID(fwID))
			resp, err := base.BizClient.DescribeFirewallResource(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []FirewallResourceRow{}
			for _, rs := range resp.ResourceSet {
				row := FirewallResourceRow{}
				row.ResourceName = rs.Name
				row.ResourceID = rs.ResourceID
				row.ResourceType = rs.ResourceType
				row.IntranetIP = rs.PrivateIP
				row.Group = rs.Tag
				row.Remark = rs.Remark
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&fwID, "fw-id", "", "Required. Resource ID of firewall")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	flags.SetFlagValuesFunc("fw-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")

	return cmd
}

// NewCmdFirewallUpdate ucloud firewall update
func NewCmdFirewallUpdate(out io.Writer) *cobra.Command {
	fwIDs := []string{}
	req := base.BizClient.NewUpdateFirewallAttributeRequest()
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update firewall attribute, such as name,group and remark.",
		Long:    "Update firewall attribute, such as name,group and remark.",
		Example: `ucloud firewall update --fw-id firewall-2xxxx/test2 --name test_update.1 --remark "this is a remark"`,
		Run: func(c *cobra.Command, args []string) {
			if *req.Name == "" && *req.Tag == "" && *req.Remark == "" {
				fmt.Fprintln(out, "Error: name, group and remark can't be all empty")
				return
			}
			if *req.Name == "" {
				req.Name = nil
			}
			if *req.Tag == "" {
				req.Tag = nil
			}
			if *req.Remark == "" {
				req.Remark = nil
			}
			for _, id := range fwIDs {
				req.FWId = sdk.String(base.PickResourceID(id))
				_, err := base.BizClient.UpdateFirewallAttribute(req)
				if err != nil {
					base.HandleError(err)
					return
				}
				fmt.Fprintf(out, "firewall[%s] updated\n", id)
			}
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls")
	req.Region = flags.String("region", base.ConfigIns.Region, "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", base.ConfigIns.ProjectID, "Optional. Project-id, see 'ucloud project list'")
	req.Name = flags.String("name", "", "Name of firewall")
	req.Tag = flags.String("group", "", "Group of firewall")
	req.Remark = flags.String("remark", "", "Remark of firewall")

	flags.SetFlagValuesFunc("fw-id", func() []string {
		return getFirewallIDNames(*req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")

	return cmd
}

func getFirewallIDNames(project, region string) (idNames []string) {
	list, err := getAllFirewallIns(project, region)
	if err != nil {
		return
	}
	for _, f := range list {
		idNames = append(idNames, f.FWId+"/"+f.Name)
	}
	return
}

func getFirewall(fwNameID, project, region string) (*unet.FirewallDataSet, error) {
	var firewall *unet.FirewallDataSet
	list, err := getAllFirewallIns(project, region)
	if err != nil {
		return nil, err
	}
	for i, fw := range list {
		if fw.FWId == fwNameID || fw.Name == fwNameID {
			firewall = &list[i]
		}
	}
	if firewall == nil {
		return nil, fmt.Errorf("firwall[%s] does not exist", fwNameID)
	}
	return firewall, nil
}

func getAllFirewallIns(project, region string) ([]unet.FirewallDataSet, error) {
	req := base.BizClient.NewDescribeFirewallRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	list := []unet.FirewallDataSet{}
	for offset, limit := 0, 100; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := base.BizClient.DescribeFirewall(req)
		if err != nil {
			return nil, err
		}
		for _, fw := range resp.DataSet {
			list = append(list, fw)
		}
		if resp.TotalCount < offset+limit {
			break
		}
	}
	return list, nil
}
