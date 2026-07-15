package nlb

import (
	"fmt"

	nlbsdk "github.com/ucloud/ucloud-sdk-go/services/nlb"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// productName is the single source of truth for the product's command name and
// its resource-id flag (`--nlb-id`).
const productName = "nlb"

// resourceIDFlag is the NLB instance resource-id flag, named after the product.
const resourceIDFlag = productName + "-id" // "nlb-id"

// getAllNLB returns every NLB instance in the active region/project, paging
// through the DescribeNetworkLoadBalancers result set.
func getAllNLB(ctx *cli.Context, project, region string) ([]nlbsdk.NetworkLoadBalancer, error) {
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDescribeNetworkLoadBalancersRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)

	list := []nlbsdk.NetworkLoadBalancer{}
	for offset, limit := 0, 100; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeNetworkLoadBalancers(req)
		if err != nil {
			return nil, err
		}
		list = append(list, resp.NLBs...)
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list, nil
}

// getAllNLBIDNames returns "nlbId/name" completion candidates for --nlb-id.
func getAllNLBIDNames(ctx *cli.Context, project, region string) []string {
	list, err := getAllNLB(ctx, project, region)
	if err != nil {
		return nil
	}
	idNames := make([]string, 0, len(list))
	for _, n := range list {
		idNames = append(idNames, fmt.Sprintf("%s/%s", n.NLBId, n.Name))
	}
	return idNames
}

// getAllListeners returns the listeners of a given NLB instance.
func getAllListeners(ctx *cli.Context, nlbID, project, region string) ([]nlbsdk.Listener, error) {
	if nlbID == "" {
		return nil, fmt.Errorf("nlb-id can't be empty")
	}
	client := cli.NewServiceClient(ctx, nlbsdk.NewClient)
	req := client.NewDescribeNLBListenersRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	req.NLBId = sdk.String(cli.PickResourceID(nlbID))
	resp, err := client.DescribeNLBListeners(req)
	if err != nil {
		return nil, err
	}
	return resp.Listeners, nil
}

// getAllListenerIDNames returns "listenerId/name" completion candidates.
func getAllListenerIDNames(ctx *cli.Context, nlbID, project, region string) []string {
	listeners, err := getAllListeners(ctx, nlbID, project, region)
	if err != nil {
		return nil
	}
	idNames := make([]string, 0, len(listeners))
	for _, l := range listeners {
		idNames = append(idNames, fmt.Sprintf("%s/%s", l.ListenerId, l.Name))
	}
	return idNames
}

// getAllTargetIDNames returns the target ids attached to a listener, in
// "targetId/resource" form for --target-id completion.
func getAllTargetIDNames(ctx *cli.Context, nlbID, listenerID, project, region string) []string {
	listeners, err := getAllListeners(ctx, nlbID, project, region)
	if err != nil {
		return nil
	}
	wantListener := cli.PickResourceID(listenerID)
	idNames := []string{}
	for _, l := range listeners {
		if wantListener != "" && l.ListenerId != wantListener {
			continue
		}
		for _, t := range l.Targets {
			label := t.ResourceId
			if label == "" {
				label = t.ResourceIP
			}
			idNames = append(idNames, fmt.Sprintf("%s/%s", t.Id, label))
		}
	}
	return idNames
}

// getAllVPCIDNames returns "vpcId/name" candidates via the VPC SDK service
// package. Cross-product completion uses the peer SDK directly, never imports
// products/vpc (§8, check-product rule 1).
func getAllVPCIDNames(ctx *cli.Context, project, region string) []string {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeVPCRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	resp, err := client.DescribeVPC(req)
	if err != nil {
		return nil
	}
	idNames := make([]string, 0, len(resp.DataSet))
	for _, v := range resp.DataSet {
		idNames = append(idNames, fmt.Sprintf("%s/%s", v.VPCId, v.Name))
	}
	return idNames
}

// getAllSubnetIDNames returns "subnetId/name" candidates, optionally scoped to
// a VPC, via the VPC SDK service package.
func getAllSubnetIDNames(ctx *cli.Context, vpcID, project, region string) []string {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	if vpcID != "" {
		req.VPCId = sdk.String(cli.PickResourceID(vpcID))
	}
	resp, err := client.DescribeSubnet(req)
	if err != nil {
		return nil
	}
	idNames := make([]string, 0, len(resp.DataSet))
	for _, s := range resp.DataSet {
		idNames = append(idNames, fmt.Sprintf("%s/%s", s.SubnetId, s.SubnetName))
	}
	return idNames
}

// derefStr safely dereferences a *string bound by a flag.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
