package firewall

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
				ctx.HandleError(fmt.Errorf("flags rules and rules-file can't be both empty"))
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
