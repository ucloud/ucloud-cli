package firewall

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// NewCommand builds the `firewall` root command and mounts the 9 subcommands.
// Mirrors cmd/firewall.go NewCmdFirewall (same AddCommand order).
func NewCommand(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "List and manipulate extranet firewall",
		Long:  `List and manipulate extranet firewall`,
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newList(ctx))
	cmd.AddCommand(newCreate(ctx))
	cmd.AddCommand(newAddRule(ctx))
	cmd.AddCommand(newDeleteRule(ctx))
	cmd.AddCommand(newApply(ctx))
	cmd.AddCommand(newCopy(ctx))
	cmd.AddCommand(newDelete(ctx))
	cmd.AddCommand(newResource(ctx))
	cmd.AddCommand(newUpdate(ctx))

	return cmd
}

// newList ucloud firewall list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeFirewallRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List extranet firewall",
		Long:  `List extranet firewall`,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.DescribeFirewall(req)
			if err != nil {
				ctx.HandleError(err)
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
				row.CreationTime = common.FormatDate(fw.CreateTime)
				if fw.Remark != "" {
					row.FirewallName += "\nremark:" + fw.Remark + "\n"
				}
				for _, r := range fw.Rule {
					rule := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
					row.Rule += rule + "\n"
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.FWId = flags.String("firewall-id", "", "Optional. The Rsource ID of firewall. Return all firewalls by default.")
	req.ResourceType = flags.String("bound-resource-type", "", "Optional. The type of resource bound on the firewall")
	req.ResourceId = flags.String("bound-resource-id", "", "Optional. The resource ID of resource bound on the firewall")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	return cmd
}

// newCreate ucloud firewall create
func newCreate(ctx *cli.Context) *cobra.Command {
	var rulesFilePath string
	var rules []string

	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create firewall",
		Long:    "Create firewall",
		Example: `ucloud firewall create --name test3 --rules "TCP|22|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt`,
		Run: func(c *cobra.Command, args []string) {
			if rules == nil && rulesFilePath == "" {
				fmt.Fprintln(ctx.Err(), "Error: flags rules and rules-file can't be both empty")
				return
			}
			if rulesFilePath != "" {
				lines, err := parseRulesFromFile(rulesFilePath)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				rules = append(rules, lines...)
			}
			req.Rule = rules
			resp, err := client.CreateFirewall(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] created\n", resp.FWId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.FWId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&rules, "rules", nil, "Required if rules-file doesn't exist. Schema: Protocol|Port|IP|Action|Level. Prototol range 'TCP','UDP','ICMP' and 'GRE'; Port is a local port accessed by source address, port range [0-65535]; IP is the source address of the network packet that requests ucloud host resource, supporting IP address and network segment, such as '120.132.69.216' or '0.0.0.0/0'; Action is the processing behavior of the packet when the firewall is in effect, including 'ACCEPT' AND 'DROP'; Level, when a rule is added to a firewall, the rules take effect in order of level, which range 'HIGH','MEDIUM' and 'LOW'. For example, 'TCP|22|192.168.1.1/22|DROP|LOW'")
	flags.StringVar(&rulesFilePath, "rules-file", "", "Required if rules doesn't exist. Path of rules file, in which each rule occupies one line. Schema: Protocol|Port|IP|Action|Level.")
	req.Name = flags.String("name", "", "Required. Name of firewall to create")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.Tag = flags.String("group", "", "Optional. Group of the firewall to create")
	req.Remark = flags.String("remark", "", "Optional. Remark of the firewall to create")
	cmd.MarkFlagRequired("name")
	command.SetCompletion(cmd, "rules-file", func() []string {
		return common.GetFileList("")
	})
	return cmd
}

// newAddRule ucloud firewall add-rule
func newAddRule(ctx *cli.Context) *cobra.Command {
	var rulesFilePath string
	var fwIDs []string
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUpdateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "add-rule",
		Short:   "Add rule to firewall instance",
		Long:    "Add rule to firewall instance",
		Example: `ucloud firewall add-rule --fw-id firewall-2xxxxz/test.lxj2 --rules "TCP|24|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt`,
		Run: func(c *cobra.Command, args []string) {
			if req.Rule == nil && rulesFilePath == "" {
				fmt.Fprintln(ctx.Err(), "Error: flags rules and rules-file can't be both empty")
				return
			}
			results := []cli.OpResultRow{}
			for _, fwID := range fwIDs {
				id := ctx.PickResourceID(fwID)
				req.FWId = &id
				firewall, err := getFirewall(ctx, *req.FWId, *req.ProjectId, *req.Region)
				if err != nil {
					ctx.HandleError(err)
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
						ctx.HandleError(err)
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
				_, err = client.UpdateFirewall(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] updated\n", fwID)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "add-rule", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls to update")
	flags.StringSliceVar(&req.Rule, "rules", nil, "Required if rules-file is empay. Rules to add to firewall. Schema:'Protocol|Port|IP|Action|Level'. See 'ucloud firewall create --help' for detail.")
	flags.StringVar(&rulesFilePath, "rules-file", "", "Required if rules is empty. Path of rules file, in which each rule occupies one line. Schema: Protocol|Port|IP|Action|Level.")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "rules-file", func() []string {
		return common.GetFileList("")
	})

	cmd.MarkFlagRequired("fw-id")
	return cmd
}

// newDeleteRule ucloud firewall remove-rule
func newDeleteRule(ctx *cli.Context) *cobra.Command {
	var rulesFilePath string
	var fwIDs []string
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUpdateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "remove-rule",
		Short:   "Remove rule from firewall instance",
		Long:    "Remove rule from firewall instance",
		Example: `ucloud firewall remove-rule --fw-id firewall-2cxxxz/test.lxj2 --rules "TCP|24|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt`,
		Run: func(c *cobra.Command, args []string) {
			if req.Rule == nil && rulesFilePath == "" {
				fmt.Fprintln(ctx.Err(), "Error: flags rules and rules-file can't be both empty")
				return
			}
			results := []cli.OpResultRow{}
			for _, fwID := range fwIDs {
				id := ctx.PickResourceID(fwID)
				req.FWId = &id
				firewall, err := getFirewall(ctx, *req.FWId, *req.ProjectId, *req.Region)
				if err != nil {
					ctx.HandleError(err)
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
						ctx.HandleError(err)
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
					fmt.Fprintf(ctx.Err(), "Error: rules can't be all deleted\n")
					return
				}
				_, err = client.UpdateFirewall(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] updated\n", fwID)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "remove-rule", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls to update")
	flags.StringSliceVar(&req.Rule, "rules", nil, "Required if rules-file is empay. Rules to add to firewall. Schema:'Protocol|Port|IP|Action|Level'. See 'ucloud firewall create --help' for detail.")
	flags.StringVar(&rulesFilePath, "rules-file", "", "Required if rules is empty. Path of rules file, in which each rule occupies one line. Schema: Protocol|Port|IP|Action|Level.")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")
	return cmd
}

// newApply ucloud firewall apply
func newApply(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewGrantFirewallRequest()
	resourceIDs := []string{}
	fwID := ""
	cmd := &cobra.Command{
		Use:     "apply",
		Short:   "Applay firewall to ucloud service",
		Long:    "Applay firewall to ucloud service",
		Example: "ucloud firewall apply --fw-id firewall-xxx --resource-id uhost-xxx --resource-type uhost",
		Run: func(c *cobra.Command, args []string) {
			req.FWId = sdk.String(ctx.PickResourceID(fwID))
			results := []cli.OpResultRow{}
			for _, id := range resourceIDs {
				req.ResourceId = sdk.String(id)
				_, err := client.GrantFirewall(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] applied to %s[%s]\n", fwID, *req.ResourceType, id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "apply", Status: "Applied"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&fwID, "fw-id", "", "Required. Resource ID of firewall to apply to some ucloud resource")
	req.ResourceType = flags.String("resource-type", "", "Required. Resource type of resource to be applied firewall. Range 'uhost','unatgw','upm','hadoophost','fortresshost','udhost','udockhost','dbaudit'.")
	flags.StringSliceVar(&resourceIDs, "resource-id", nil, "Resource ID of resources to be applied firewall")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetFlagValues(cmd, "resource-type", "uhost", "unatgw", "upm", "hadoophost", "fortresshost", "udhost", "udockhost", "dbaudit")
	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("resource-type")

	return cmd
}

// newCopy ucloud firewall copy
func newCopy(ctx *cli.Context) *cobra.Command {
	srcFirewall := ""
	srcRegion := ""
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateFirewallRequest()
	cmd := &cobra.Command{
		Use:     "copy",
		Short:   "Copy firewall",
		Long:    "Copy firewall",
		Example: "ucloud firewall copy --src-fw firewall-xxx --target-region cn-bj2 --name test",
		Run: func(c *cobra.Command, args []string) {
			fwID := ctx.PickResourceID(srcFirewall)
			firewall, err := getFirewall(ctx, fwID, *req.ProjectId, srcRegion)

			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.Tag = sdk.String(firewall.Tag)
			req.Remark = sdk.String(firewall.Remark)
			for _, r := range firewall.Rule {
				rstr := fmt.Sprintf("%s|%s|%s|%s|%s", r.ProtocolType, r.DstPort, r.SrcIP, r.RuleAction, r.Priority)
				req.Rule = append(req.Rule, rstr)
			}
			resp, err := client.CreateFirewall(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] created from %s\n", resp.FWId, srcFirewall)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.FWId, Action: "copy", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringVar(&srcFirewall, "src-fw", "", "Required. ResourceID or name of source firewall")
	req.Name = flags.String("name", "", "Required. Name of new firewall")
	flags.StringVar(&srcRegion, "region", ctx.DefaultRegion(), "Optional. Current region, used to fetch source firewall")
	req.Region = flags.String("target-region", ctx.DefaultRegion(), "Optional. Copy firewall to target region")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	command.SetCompletion(cmd, "src-fw", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, srcRegion)
	})
	command.SetCompletion(cmd, "target-region", ctx.RegionList)
	command.SetCompletion(cmd, "region", ctx.RegionList)

	cmd.MarkFlagRequired("src-fw")
	cmd.MarkFlagRequired("name")

	return cmd
}

// newDelete ucloud firewall delete
func newDelete(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDeleteFirewallRequest()
	ids := []string{}
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete firewall by resource ids or names",
		Long:    "Delete firewall by resource ids or names",
		Example: "ucloud firewall delete --fw-id firewall-xxx",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, id := range ids {
				rid := ctx.PickResourceID(id)
				req.FWId = sdk.String(rid)
				_, err := client.DeleteFirewall(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: rid, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "fw-id", nil, "Required. Resource IDs of firewall to delete")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("fw-id")
	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

// newResource ucloud firewall resource
func newResource(ctx *cli.Context) *cobra.Command {
	fwID := ""
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeFirewallResourceRequest()
	cmd := &cobra.Command{
		Use:   "resource",
		Short: "List resources that has been applied the firewall",
		Long:  "List resources that has been applied the firewall",
		Run: func(c *cobra.Command, args []string) {
			req.FWId = sdk.String(ctx.PickResourceID(fwID))
			resp, err := client.DescribeFirewallResource(req)
			if err != nil {
				ctx.HandleError(err)
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
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&fwID, "fw-id", "", "Required. Resource ID of firewall")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.Offset = flags.Int("offset", 0, "Optional. Offset")
	req.Limit = flags.Int("limit", 50, "Optional. Limit")

	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")

	return cmd
}

// newUpdate ucloud firewall update
func newUpdate(ctx *cli.Context) *cobra.Command {
	fwIDs := []string{}
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewUpdateFirewallAttributeRequest()
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update firewall attribute, such as name,group and remark.",
		Long:    "Update firewall attribute, such as name,group and remark.",
		Example: `ucloud firewall update --fw-id firewall-2xxxx/test2 --name test_update.1 --remark "this is a remark"`,
		Run: func(c *cobra.Command, args []string) {
			if *req.Name == "" && *req.Tag == "" && *req.Remark == "" {
				fmt.Fprintln(ctx.Err(), "Error: name, group and remark can't be all empty")
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
			results := []cli.OpResultRow{}
			for _, id := range fwIDs {
				rid := ctx.PickResourceID(id)
				req.FWId = sdk.String(rid)
				_, err := client.UpdateFirewallAttribute(req)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				fmt.Fprintf(ctx.ProgressWriter(), "firewall[%s] updated\n", id)
				results = append(results, cli.OpResultRow{ResourceID: rid, Action: "update", Status: "Updated"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&fwIDs, "fw-id", nil, "Required. Resource ID of firewalls")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Region, see 'ucloud region'")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")
	req.Name = flags.String("name", "", "Name of firewall")
	req.Tag = flags.String("group", "", "Group of firewall")
	req.Remark = flags.String("remark", "", "Remark of firewall")

	command.SetCompletion(cmd, "fw-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("fw-id")

	return cmd
}
