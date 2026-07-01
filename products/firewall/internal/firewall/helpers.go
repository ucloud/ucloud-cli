package firewall

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// parseRulesFromFile reads a rules file, one rule per line. Verbatim from
// cmd/firewall.go.
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

// getFirewall finds a firewall by resource id or name. Ported from
// cmd/firewall.go (base.BizClient → cli.NewServiceClient).
func getFirewall(ctx *cli.Context, fwNameID, project, region string) (*unet.FirewallDataSet, error) {
	var firewall *unet.FirewallDataSet
	list, err := getAllFirewallIns(ctx, project, region)
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

// getAllFirewallIns lists all firewalls in project/region, paging by 100.
// Ported from cmd/firewall.go (base.BizClient → cli.NewServiceClient).
func getAllFirewallIns(ctx *cli.Context, project, region string) ([]unet.FirewallDataSet, error) {
	client := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewDescribeFirewallRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	list := []unet.FirewallDataSet{}
	for offset, limit := 0, 100; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeFirewall(req)
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
