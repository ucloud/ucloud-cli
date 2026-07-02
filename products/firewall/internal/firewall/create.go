package firewall

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/unet"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

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
