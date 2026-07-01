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
	"fmt"
	"strings"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
)

// uhost_compat.go is a cross-batch transition shim. uhost migrated to
// products/uhost (Part 6), but cmd/ext.go (the `ext uhost switch-eip` command,
// still in package cmd — platform/ext, batch2) calls two cmd-local uhost
// helpers:
//   - describeUHostByID: cmd/ext.go ~:75 (lookup uhost before switching EIP)
//   - getUhostList:      cmd/ext.go ~:201 (--uhost-id completion)
// These copies keep the SAME names+signatures as the originals in the deleted
// cmd/uhost.go (base.BizClient path), so package cmd keeps compiling. Remove this
// file once cmd/ext.go migrates (batch2).

// describeUHostByID looks up a single uhost by id (cmd-local, base.BizClient).
// Same signature as the original cmd/uhost.go describeUHostByID.
func describeUHostByID(uhostID, projectID, region, zone string) (interface{}, error) {
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	req.UHostIds = []string{uhostID}
	req.ProjectId = &projectID
	req.Region = &region
	req.Zone = &zone

	resp, err := base.BizClient.DescribeUHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.UHostSet) < 1 {
		return nil, fmt.Errorf("uhost [%s] does not exist", uhostID)
	}

	return &resp.UHostSet[0], nil
}

// getUhostList returns "UHostId/Name" completion candidates filtered by states
// (nil = all). Same signature as the original cmd/uhost.go getUhostList.
func getUhostList(states []string, project, region, zone string) []string {
	req := base.BizClient.NewDescribeUHostInstanceRequest()
	req.ProjectId = sdk.String(project)
	req.Region = sdk.String(region)
	req.Zone = sdk.String(zone)
	req.Limit = sdk.Int(50)
	resp, err := base.BizClient.DescribeUHostInstance(req)
	if err != nil {
		//todo runtime log
		return nil
	}
	list := []string{}
	for _, host := range resp.UHostSet {
		if states != nil {
			for _, s := range states {
				if host.State == s {
					list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
				}
			}
		} else {
			list = append(list, host.UHostId+"/"+strings.Replace(host.Name, " ", "-", -1))
		}
	}
	return list
}
