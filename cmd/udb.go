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
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/model/status"
)

//NewCmdUDBConf ucloud udb conf
func NewCmdUDBConf() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conf",
		Short: "List and manipulate configuration files of MySQL instances",
		Long:  "List and manipulate configuration files of MySQL instances",
	}
	out := base.Cxt.GetWriter()
	cmd.AddCommand(NewCmdUDBConfList(out))
	cmd.AddCommand(NewCmdUDBConfDescribe(out))
	cmd.AddCommand(NewCmdUDBConfClone(out))
	cmd.AddCommand(NewCmdUDBConfUpload(out))
	cmd.AddCommand(NewCmdUDBConfUpdate(out))
	cmd.AddCommand(NewCmdUDBConfDelete(out))
	cmd.AddCommand(NewCmdUDBConfApply(out))
	cmd.AddCommand(NewCmdUDBConfDownload(out))
	return cmd
}

//UDBConfRow 表格行
type UDBConfRow struct {
	ConfID      int
	DBVersion   string
	Name        string
	Description string
	Modifiable  bool
	Zone        string
}

var dbTypeMap = map[string]string{
	"mysql":      "sql",
	"mongodb":    "nosql",
	"postgresql": "postgresql",
	"sqlserver":  "sqlserver",
}

var dbTypeList = []string{"mysql", "mongodb", "postgresql", "sqlserver"}

//NewCmdUDBConfList ucloud mysql conf list
func NewCmdUDBConfList(out io.Writer) *cobra.Command {
	req := base.BizClient.NewDescribeUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuartion files of MySQL instances",
		Long:  "List configuartion files of MySQL instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.GroupId == 0 {
				req.GroupId = nil
			}
			resp, err := base.BizClient.DescribeUDBParamGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			list := []UDBConfRow{}
			for _, ins := range resp.DataSet {
				row := UDBConfRow{
					ConfID:      ins.GroupId,
					Name:        ins.GroupName,
					Zone:        ins.Zone,
					DBVersion:   ins.DBTypeId,
					Description: ins.Description,
					Modifiable:  ins.Modifiable,
				}
				list = append(list, row)
			}
			base.PrintList(list, out)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	bindOffset(req, flags)
	bindLimit(req, flags)
	req.GroupId = flags.Int("conf-id", 0, "Optional. Configuration identifier for the configuration to be described")
	req.ClassType = sdk.String("sql")

	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getConfIDList(*req.ClassType, *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//UDBConfParamRow 参数配置展示表格行
type UDBConfParamRow struct {
	Key   string
	Value string
}

//NewCmdUDBConfDescribe ucloud udb conf describe
func NewCmdUDBConfDescribe(out io.Writer) *cobra.Command {
	var confID string
	req := base.BizClient.NewDescribeUDBParamGroupRequest()
	req.RegionFlag = sdk.Bool(false)
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display details about a configuration file of MySQL instance",
		Long:  "Display details about a configuration file of MySQL instance",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(base.PickResourceID(confID))
			if err != nil {
				base.HandleError(err)
				return
			}
			req.GroupId = &id
			resp, err := base.BizClient.DescribeUDBParamGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			if len(resp.DataSet) != 1 {
				fmt.Fprintf(out, "Error, conf-id[%d] may not be exist\n", req.GroupId)
				return
			}
			conf := resp.DataSet[0]
			attrs := []base.DescribeTableRow{
				base.DescribeTableRow{Attribute: "ConfID", Content: strconv.Itoa(conf.GroupId)},
				base.DescribeTableRow{Attribute: "DBVersion", Content: conf.DBTypeId},
				base.DescribeTableRow{Attribute: "Name", Content: conf.GroupName},
				base.DescribeTableRow{Attribute: "Description", Content: conf.Description},
				base.DescribeTableRow{Attribute: "Modifiable", Content: strconv.FormatBool(conf.Modifiable)},
				base.DescribeTableRow{Attribute: "Zone", Content: conf.Zone},
			}
			fmt.Fprintln(out, "Attributes:")
			base.PrintList(attrs, out)

			params := []UDBConfParamRow{}
			for _, p := range conf.ParamMember {
				if p.Value == "" {
					continue
				}
				row := UDBConfParamRow{
					Key:   p.Key,
					Value: p.Value,
				}
				params = append(params, row)
			}
			fmt.Fprintln(out, "\nParameters:")
			base.PrintList(params, out)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Requried. Configuration identifier for the configuration to be described")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("conf-id")
	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getConfIDList("sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//NewCmdUDBConfClone ucloud udb conf clone
func NewCmdUDBConfClone(out io.Writer) *cobra.Command {
	var srcConfID string
	req := base.BizClient.NewCreateUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Create configuration file by cloning existed configuration",
		Long:  "Create configuration file by cloning existed configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(base.PickResourceID(srcConfID))
			if err != nil {
				base.HandleError(err)
				return
			}
			if *req.DBTypeId == "" {
				confIns, err := getConfByID(id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					base.HandleError(err)
					return
				}
				req.DBTypeId = sdk.String(confIns.DBTypeId)
			}
			req.SrcGroupId = &id
			resp, err := base.BizClient.CreateUDBParamGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "conf[%d] created\n", resp.GroupId)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBTypeId = flags.String("db-version", "", fmt.Sprintf("Required. Version of DB. Accept values:%s", strings.Join(dbVersionList, ", ")))
	req.GroupName = flags.String("name", "", "Required. Name of configuration. It's length should be between 6 and 63")
	req.Description = flags.String("description", " ", "Optional. Description of the configuration to clone")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)
	flags.StringVar(&srcConfID, "src-conf-id", "", "Optional. The ConfID of source configuration which to be cloned from")

	flags.SetFlagValues("db-version", dbVersionList...)
	flags.SetFlagValuesFunc("src-conf-id", func() []string {
		return getConfIDList("sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("src-conf-id")
	return cmd
}

var udbSubtypeMap = map[string]int{
	"unknow":               0,
	"Shardsvr-MMAPv1":      1,
	"Shardsvr-WiredTiger":  2,
	"Configsvr-MMAPv1":     3,
	"Configsvr-WiredTiger": 4,
	"Mongos":               5,
	"Mysql":                10,
	"Postgresql":           20,
}

var subtypeList = []string{"Shardsvr-MMAPv1", "Shardsvr-WiredTiger", "Configsvr-MMAPv1", "Configsvr-WiredTiger", "Mongos", "Mysql", "Postgresql"}

//NewCmdUDBConfUpload ucloud udb conf upload
func NewCmdUDBConfUpload(out io.Writer) *cobra.Command {
	var file string
	req := base.BizClient.NewUploadUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Create configuration file by uploading local DB configuration file",
		Long:  "Create configuration file by uploading local DB configuration file",
		Run: func(c *cobra.Command, args []string) {
			content, err := readFile(file)
			if err != nil {
				base.HandleError(err)
				return
			}
			if l := len(*req.GroupName); l < 6 || l > 63 {
				fmt.Fprintln(out, "Error, length of name shoud be between 6 and 63")
				return
			}
			req.Content = sdk.String(base64.StdEncoding.EncodeToString([]byte(content)))
			req.ParamGroupTypeId = sdk.Int(10)
			resp, err := base.BizClient.UploadUDBParamGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "conf[%d] uploaded\n", resp.GroupId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&file, "conf-file", "", "Required. Path of local configuration file")
	req.DBTypeId = flags.String("db-version", "", fmt.Sprintf("Required. Version of DB. Accept values:%s", strings.Join(dbVersionList, ", ")))
	req.GroupName = flags.String("name", "", "Required. Name of configuration. It's length should be between 6 and 63")
	req.Description = flags.String("description", " ", "Optional. Description of the configuration to clone")
	// flags.StringVar(&subtype, "db-type", "", fmt.Sprintf("Optional. DB type. Accept values: %s", strings.Join(subtypeList, ", ")))
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("conf-file")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("db-version")
	// cmd.MarkFlagRequired("db-type")

	flags.SetFlagValues("db-version", dbVersionList...)
	// flags.SetFlagValues("db-type", subtypeList...)
	flags.SetFlagValuesFunc("conf-file", func() []string {
		return base.GetFileList("")
	})
	return cmd
}

//NewCmdUDBConfUpdate ucloud udb conf update
func NewCmdUDBConfUpdate(out io.Writer) *cobra.Command {
	var confID, key, value, file string
	req := base.BizClient.NewUpdateUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update parameters of DB's configuration",
		Long:  "Update parameters of DB's configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(base.PickResourceID(confID))
			if err != nil {
				base.HandleError(err)
				return
			}
			req.GroupId = &id

			if key != "" && value != "" {
				req.Key = &key
				req.Value = &value
				_, err := base.BizClient.UpdateUDBParamGroup(req)
				if err != nil {
					base.HandleError(err)
				} else {
					fmt.Printf("conf[%s]'sparameter[%s = %s] updated\n", confID, key, value)
				}
			}
			if file != "" {
				params, err := parseParam(file)
				if err != nil {
					base.HandleError(err)
					return
				}
				for _, p := range params {
					req.Key = sdk.String(p.Key)
					req.Value = sdk.String(p.Value)
					_, err := base.BizClient.UpdateUDBParamGroup(req)
					if err != nil {
						fmt.Printf("conf[%s]'sparameter[%s = %s] failed\n", confID, p.Key, p.Value)
						base.HandleError(err)
					} else {
						fmt.Printf("conf[%s]'sparameter[%s = %s] updated\n", confID, p.Key, p.Value)
					}
					fmt.Println("")
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration to update")
	flags.StringVar(&key, "key", "", "Optional. Key of parameter")
	flags.StringVar(&value, "value", "", "Optional. Value of parameter")
	flags.StringVar(&file, "file", "", "Optional. Path of file in which each parameter occupies one line with format 'key = value'")

	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getModifiableConfIDList("", *req.ProjectId, *req.Region, *req.Zone)
	})
	flags.SetFlagValuesFunc("file", func() []string {
		return base.GetFileList("")
	})

	cmd.MarkFlagRequired("conf-id")
	return cmd
}

//NewCmdUDBConfDelete ucloud udb conf delete
func NewCmdUDBConfDelete(out io.Writer) *cobra.Command {
	var confID string
	req := base.BizClient.NewDeleteUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete configuration of udb by conf-id",
		Long:  "Delete configuration of udb by conf-id",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(base.PickResourceID(confID))
			if err != nil {
				base.HandleError(err)
				return
			}
			req.GroupId = &id
			_, err = base.BizClient.DeleteUDBParamGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintf(out, "conf[%s] deleted\n", confID)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of the configuration to delete")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("conf-id")
	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getModifiableConfIDList("", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

// NewCmdUDBConfApply ucloud udb conf apply
func NewCmdUDBConfApply(out io.Writer) *cobra.Command {
	var confID string
	var udbIDs []string
	var restart, yes, async bool

	req := base.BizClient.NewChangeUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply configuration for UDB instances",
		Long:  "Apply configuration for UDB instances",
		Run: func(c *cobra.Command, args []string) {
			req.GroupId = sdk.String(base.PickResourceID(confID))
			for _, idname := range udbIDs {
				req.DBId = sdk.String(base.PickResourceID(idname))
				_, err := base.BizClient.ChangeUDBParamGroup(req)
				if err != nil {
					base.HandleError(err)
					continue
				}
				fmt.Fprintf(out, "conf[%s] has applied for udb[%s]\n", confID, idname)
				if !restart {
					continue
				}
				ok := base.Confirm(yes, fmt.Sprintf("udb[%s] is about to restart, do you want to continue?", idname))
				if !ok {
					continue
				}
				restartReq := base.BizClient.NewRestartUDBInstanceRequest()
				restartReq.Region = req.Region
				restartReq.Zone = req.Zone
				restartReq.ProjectId = req.ProjectId
				restartReq.DBId = req.DBId
				_, err = base.BizClient.RestartUDBInstance(restartReq)
				if err != nil {
					base.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(out, "udb[%s] is restarting\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is restarting", idname)
					poller.Spoll(*req.DBId, text, []string{status.UDB_FAIL, status.UDB_RUNNING})
				}
			}
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of the configuration to be applied")
	flags.StringSliceVar(&udbIDs, "udb-id", nil, "Required. Resource ID of UDB instances to change configuration")
	flags.BoolVar(&restart, "restart-after-apply", true, "Optional. The new configuration will take effect after DB restarts")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation")
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the long-running operation to finish.")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("conf-id")
	cmd.MarkFlagRequired("udb-id")

	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getModifiableConfIDList("", *req.ProjectId, *req.Region, *req.Zone)
	})
	flags.SetFlagValuesFunc("udb-id", func() []string {
		return getUDBIDList(nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

//NewCmdUDBConfDownload ucloud udb conf download
func NewCmdUDBConfDownload(out io.Writer) *cobra.Command {
	var confID string
	req := base.BizClient.NewExtractUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download UDB configuration",
		Long:  "Download UDB configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(base.PickResourceID(confID))
			if err != nil {
				base.HandleError(err)
				return
			}

			req.GroupId = &id
			resp, err := base.BizClient.ExtractUDBParamGroup(req)
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprint(out, resp.Content)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration to download")
	bindRegion(req, flags)
	bindZone(req, flags)
	bindProjectID(req, flags)

	cmd.MarkFlagRequired("conf-id")

	flags.SetFlagValuesFunc("conf-id", func() []string {
		return getConfIDList("sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

type confParam struct {
	Key   string
	Value string
}

func parseParam(filePath string) ([]confParam, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	params := []confParam{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		strs := strings.SplitN(line, "=", 2)
		if len(strs) < 2 {
			continue
		}
		param := confParam{
			Key:   strings.TrimSpace(strs[0]),
			Value: strings.TrimSpace(strs[1]),
		}
		params = append(params, param)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return params, nil
}

func getConfByID(confID int, project, region, zone string) (*udb.UDBParamGroupSet, error) {
	req := base.BizClient.NewDescribeUDBParamGroupRequest()
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	req.GroupId = &confID
	resp, err := base.BizClient.DescribeUDBParamGroup(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) != 1 {
		return nil, fmt.Errorf("conf-id[%d] may not exist", *req.GroupId)
	}
	return &resp.DataSet[0], nil
}

func getConfList(dbType, project, region, zone string) ([]udb.UDBParamGroupSet, error) {
	req := base.BizClient.NewDescribeUDBParamGroupRequest()
	req.ClassType = &dbType
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	list := []udb.UDBParamGroupSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := base.BizClient.DescribeUDBParamGroup(req)
		if err != nil {
			return nil, err
		}
		for _, conf := range resp.DataSet {
			list = append(list, conf)
		}
		if resp.TotalCount <= offset+limit {
			break
		}
	}
	return list, nil
}

func getModifiableConfIDList(dbType, project, region, zone string) []string {
	confs, err := getConfList(dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, conf := range confs {
		if conf.Modifiable == true {
			list = append(list, fmt.Sprintf("%d/%s", conf.GroupId, conf.GroupName))
		}
	}
	return list
}

func getConfIDList(dbType, project, region, zone string) []string {
	confs, err := getConfList(dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, conf := range confs {
		list = append(list, fmt.Sprintf("%d/%s", conf.GroupId, conf.GroupName))
	}
	return list
}
