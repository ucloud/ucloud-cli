package uhost

import (
	"fmt"
	"strings"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	cliconst "github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// completion.go holds the cross-product completion-data fetchers that uhost's
// flags need (--vpc-id, --subnet-id, --firewall-id, --bind-eip, --image-id).
// Each is a self-contained SDK call COPIED from the originating product's cmd
// file (NOT imported — products stay boundary-isolated), with base.BizClient
// swapped for cli.NewServiceClient. Request-logging is dropped; completion funcs
// must stay silent.

// getAllVPCIns mirrors cmd/vpc.go getAllVPCIns.
func getAllVPCIns(ctx *cli.Context, project, region string) ([]vpc.VPCInfo, error) {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeVPCRequest()
	req.ProjectId = &project
	req.Region = &region
	resp, err := client.DescribeVPC(req)
	if err != nil {
		return nil, err
	}
	return resp.DataSet, nil
}

// getAllVPCIdNames mirrors cmd/vpc.go getAllVPCIdNames (--vpc-id completion).
func getAllVPCIdNames(ctx *cli.Context, project, region string) []string {
	vpcInsList, err := getAllVPCIns(ctx, project, region)
	list := []string{}
	if err != nil {
		return nil
	}
	for _, vpc := range vpcInsList {
		list = append(list, fmt.Sprintf("%s/%s", vpc.VPCId, vpc.Name))
	}
	return list
}

// getAllSubnets mirrors cmd/vpc.go getAllSubnets.
func getAllSubnets(ctx *cli.Context, vpcID, project, region string) ([]vpc.SubnetInfo, error) {
	client := cli.NewServiceClient(ctx, vpc.NewClient)
	req := client.NewDescribeSubnetRequest()
	req.ProjectId = sdk.String(cli.PickResourceID(project))
	req.Region = sdk.String(region)
	if vpcID != "" {
		req.VPCId = sdk.String(cli.PickResourceID(vpcID))
	}
	subnets := []vpc.SubnetInfo{}
	for limit, offset := 50, 0; ; offset += limit {
		req.Limit = sdk.Int(limit)
		req.Offset = sdk.Int(offset)
		resp, err := client.DescribeSubnet(req)
		if err != nil {
			ctx.HandleError(err)
			return nil, err
		}
		subnets = append(subnets, resp.DataSet...)
		if limit+offset >= resp.TotalCount {
			break
		}
	}
	return subnets, nil
}

// getAllSubnetIDNames mirrors cmd/vpc.go getAllSubnetIDNames (--subnet-id completion).
func getAllSubnetIDNames(ctx *cli.Context, vpcID, project, region string) []string {
	subnets, err := getAllSubnets(ctx, vpcID, project, region)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, s := range subnets {
		list = append(list, fmt.Sprintf("%s/%s", s.SubnetId, s.SubnetName))
	}
	return list
}

// getAllFirewallIns lists all firewalls in project/region, paging by 100.
// Copied self-contained from cmd/firewall_compat.go (base.BizClient →
// cli.NewServiceClient).
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

// getFirewallIDNames returns "FWId/Name" completion candidates (--firewall-id).
// Copied self-contained from cmd/firewall_compat.go.
func getFirewallIDNames(ctx *cli.Context, project, region string) (idNames []string) {
	list, err := getAllFirewallIns(ctx, project, region)
	if err != nil {
		return
	}
	for _, f := range list {
		idNames = append(idNames, f.FWId+"/"+f.Name)
	}
	return
}

// getAllEip returns "EIPId/ip1,ip2" completion candidates filtered by states and
// paymodes (nil filter = no filter). Copied self-contained from
// cmd/eip_compat.go; uses the package-local fetchAllEip (helpers.go).
func getAllEip(ctx *cli.Context, projectID, region string, states, paymodes []string) []string {
	list, err := fetchAllEip(ctx, projectID, region)
	if err != nil {
		return nil
	}
	strs := []string{}
	for _, item := range list {
		rightState := false
		if states == nil {
			rightState = true
		} else {
			for _, s := range states {
				if item.Status == s {
					rightState = true
				}
			}
		}

		rightPayMode := false
		if paymodes == nil {
			rightPayMode = true
		} else {
			for _, m := range paymodes {
				if item.PayMode == m {
					rightPayMode = true
				}
			}
		}
		if !rightPayMode || !rightState {
			continue
		}

		ips := []string{}
		for _, ip := range item.EIPAddr {
			ips = append(ips, ip.IP)
		}
		strs = append(strs, item.EIPId+"/"+strings.Join(ips, ","))
	}
	return strs
}

// getImageList returns "ImageId/ImageName" completion candidates filtered by
// states + imageType (--image-id completion on create). Copied self-contained
// from cmd/image_compat.go (base.BizClient → cli.NewServiceClient on the uhost
// SDK, which serves DescribeImage).
func getImageList(ctx *cli.Context, states []string, imageType, project, region, zone string) []string {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeImageRequest()
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	req.Limit = sdk.Int(1000)
	if imageType != cliconst.IMAGE_ALL {
		req.ImageType = sdk.String(imageType)
	}
	resp, err := client.DescribeImage(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, image := range resp.ImageSet {
		for _, s := range states {
			if image.State == s {
				list = append(list, image.ImageId+"/"+image.ImageName)
			}
		}
	}
	return list
}

// describeImageByID returns the image-feature-probe describe func, closing over
// ctx + project/region/zone. Copied self-contained from cmd/image_compat.go;
// used by create to probe an image's HotPlug/CloudInit features and by
// create-image's poller. Returns *uhostsdk.UHostImageSet.
func describeImageByID(ctx *cli.Context, project, region, zone string) func(imageID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(imageID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
		req := client.NewDescribeImageRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.ImageId = sdk.String(imageID)
		req.ProjectId = sdk.String(project)
		req.Region = sdk.String(region)
		req.Zone = sdk.String(zone)
		req.Limit = sdk.Int(50)
		resp, err := client.DescribeImage(req)
		if err != nil {
			return nil, err
		}
		if len(resp.ImageSet) < 1 {
			return nil, nil
		}
		return &resp.ImageSet[0], nil
	}
}
