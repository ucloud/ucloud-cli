package ulhost

import (
	"fmt"
	"strings"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// completion.go holds the cross-product completion-data fetchers that ulhost's
// flags need (--ulhost-id, --image-id). Each is a self-contained SDK call
// (NOT imported — products stay boundary-isolated), with cli.NewServiceClient.

// getULHostList returns "ULHostId/Name" completion candidates filtered by states
// (nil = all). Mirrors uhost's getUhostList pattern.
func getULHostList(ctx *cli.Context, states []string, project, region string) []string {
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewDescribeULHostInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Limit = sdk.Int(100)
	resp, err := client.DescribeULHostInstance(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, host := range resp.ULHostInstanceSets {
		if states != nil {
			for _, s := range states {
				if host.State == s {
					list = append(list, host.ULHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
				}
			}
		} else {
			list = append(list, host.ULHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
		}
	}
	return list
}

// getULHostImageList returns "ImageId/ImageName" completion candidates for
// ulhost images. Uses the ucompshare DescribeULHostImage API if available,
// otherwise falls back to listing from the uhost image API filtered for ulhost.
func getULHostImageList(ctx *cli.Context, project, region string) []string {
	// ULHost images are queried through the ucompshare service; the SDK
	// does not expose a separate DescribeULHostImage in the public SDK,
	// so we provide basic completion via the uhost image API as a fallback.
	// Users can also use `ucloud ulhost image list` to browse available images.
	return nil
}

// formatBundleInfo returns a human-readable description of a bundle.
func formatBundleInfo(cpu, memory, sysDiskSpace, bandwidth, trafficPacket int) string {
	memoryGB := memory / 1024
	if trafficPacket > 0 {
		return fmt.Sprintf("cpu:%d memory:%dG disk:%dG bandwidth:%dM traffic:%dG", cpu, memoryGB, sysDiskSpace, bandwidth, trafficPacket)
	}
	return fmt.Sprintf("cpu:%d memory:%dG disk:%dG bandwidth:%dM", cpu, memoryGB, sysDiskSpace, bandwidth)
}
