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

// eip_compat.go is a cross-batch transition shim. The eip product moved to
// products/eip (Part 4), but several commands still in package cmd call the
// old cmd-local eip helpers. These copies keep package cmd compiling until
// those consumers migrate. Logic is identical to the originals in the deleted
// cmd/eip.go (base.BizClient path).
//
// Consumers (kept until each migrates):
//   - getAllEip:  cmd/ulb.go, cmd/bandwidth.go, cmd/globalssh.go (batch2)
//   - sbindEIP:   cmd/ext.go (platform, batch2)
//   - bindEIP:    cmd/ulb.go (batch2)
//   - unbindEIP:  cmd/ext.go (platform, batch2)
//
// Private deps getEIPIDbyIP/fetchAllEip are included because the four public
// helpers above call them. getEIPLine stays in cmd/util.go (consumed by
// cmd/ulb.go), so it is intentionally NOT duplicated here.
//
// Remove this file once all the consumers above have migrated (batch2).

import (
	"fmt"
	"net"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

func getEIPIDbyIP(ip net.IP, projectID, region string) (string, error) {
	eipList, err := fetchAllEip(projectID, region)
	if err != nil {
		return "", err
	}
	for _, eip := range eipList {
		for _, addr := range eip.EIPAddr {
			if addr.IP == ip.String() {
				return eip.EIPId, nil
			}
		}
	}
	return "", fmt.Errorf("IP[%s] not exist", ip.String())
}

func fetchAllEip(projectID, region string) ([]unet.UnetEIPSet, error) {
	req := base.BizClient.NewDescribeEIPRequest()
	list := []unet.UnetEIPSet{}
	req.ProjectId = sdk.String(projectID)
	req.Region = sdk.String(region)
	for offset, step := 0, 100; ; offset += step {
		req.Offset = &offset
		req.Limit = &step
		resp, err := base.BizClient.DescribeEIP(req)
		if err != nil {
			return nil, err
		}
		for i, size := 0, len(resp.EIPSet); i < size; i++ {
			list = append(list, resp.EIPSet[i])
		}
		if resp.TotalCount <= offset+step {
			break
		}
	}
	return list, nil
}

// states,paymodes 为nil时，不作为过滤条件
func getAllEip(projectID, region string, states, paymodes []string) []string {
	list, err := fetchAllEip(projectID, region)
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

func bindEIP(resourceID, resourceType, eipID, projectID, region *string) {
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ip, *projectID, *region)
		if err != nil {
			base.HandleError(err)
		} else {
			*eipID = id
		}
	}
	req := base.BizClient.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(base.PickResourceID(*eipID))
	req.ProjectId = sdk.String(base.PickResourceID(*projectID))
	req.Region = region
	_, err := base.BizClient.BindEIP(req)
	if err != nil {
		base.HandleError(err)
	} else {
		base.Cxt.Printf("bind EIP[%s] with %s[%s]\n", *req.EIPId, *req.ResourceType, *req.ResourceId)
	}
}

func sbindEIP(resourceID, resourceType, eipID, projectID, region *string) ([]string, error) {
	logs := make([]string, 0)
	ip := net.ParseIP(*eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ip, *projectID, *region)
		if err != nil {
			base.HandleError(err)
		} else {
			*eipID = id
		}
	}
	req := base.BizClient.NewBindEIPRequest()
	req.ResourceId = resourceID
	req.ResourceType = resourceType
	req.EIPId = sdk.String(base.PickResourceID(*eipID))
	req.ProjectId = sdk.String(base.PickResourceID(*projectID))
	req.Region = region
	logs = append(logs, fmt.Sprintf("api: BindEIP, request: %v", base.ToQueryMap(req)))
	_, err := base.BizClient.BindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("bind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("bind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}

func unbindEIP(resourceID, resourceType, eipID, projectID, region string) ([]string, error) {
	logs := make([]string, 0)
	eipID = base.PickResourceID(eipID)
	ip := net.ParseIP(eipID)
	if ip != nil {
		id, err := getEIPIDbyIP(ip, projectID, region)
		if err != nil {
			base.HandleError(err)
		} else {
			eipID = id
		}
	}
	req := base.BizClient.NewUnBindEIPRequest()
	req.ResourceId = &resourceID
	req.ResourceType = &resourceType
	req.EIPId = &eipID
	req.ProjectId = sdk.String(base.PickResourceID(projectID))
	req.Region = &region
	logs = append(logs, fmt.Sprintf("api: UnBindEIP, request: %v", base.ToQueryMap(req)))
	_, err := base.BizClient.UnBindEIP(req)
	if err != nil {
		logs = append(logs, fmt.Sprintf("unbind eip failed: %v", err))
		return logs, err
	}
	logs = append(logs, fmt.Sprintf("unbind eip[%s] with %s[%s] successfully", *req.EIPId, *req.ResourceType, *req.ResourceId))
	return logs, nil
}
