package mysql

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBConf ucloud udb conf
func newUDBConf(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conf",
		Short: "List and manipulate configuration files of MySQL instances",
		Long:  "List and manipulate configuration files of MySQL instances",
	}
	cmd.AddCommand(newUDBConfList(ctx))
	cmd.AddCommand(newUDBConfDescribe(ctx))
	cmd.AddCommand(newUDBConfClone(ctx))
	cmd.AddCommand(newUDBConfUpload(ctx))
	cmd.AddCommand(newUDBConfUpdate(ctx))
	cmd.AddCommand(newUDBConfDelete(ctx))
	cmd.AddCommand(newUDBConfApply(ctx))
	cmd.AddCommand(newUDBConfDownload(ctx))
	return cmd
}

// UDBConfRow 表格行
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

// newUDBConfList ucloud mysql conf list
func newUDBConfList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configuartion files of MySQL instances",
		Long:  "List configuartion files of MySQL instances",
		Run: func(c *cobra.Command, args []string) {
			if *req.GroupId == 0 {
				req.GroupId = nil
			}
			resp, err := client.DescribeUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
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
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindOffset(cmd, req)
	ctx.BindLimit(cmd, req)
	req.GroupId = flags.Int("conf-id", 0, "Optional. Configuration identifier for the configuration to be described")
	req.ClassType = sdk.String("sql")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getConfIDList(ctx, *req.ClassType, *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// UDBConfParamRow 参数配置展示表格行
type UDBConfParamRow struct {
	Key   string
	Value string
}

// confDescribeRow mirrors base.DescribeTableRow for the conf describe view.
type confDescribeRow struct {
	Attribute string
	Content   string
}

// newUDBConfDescribe ucloud udb conf describe
func newUDBConfDescribe(ctx *cli.Context) *cobra.Command {
	var confID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	req.RegionFlag = sdk.Bool(false)
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display details about a configuration file of MySQL instance",
		Long:  "Display details about a configuration file of MySQL instance",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupId = &id
			resp, err := client.DescribeUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.DataSet) != 1 {
				fmt.Fprintf(ctx.Out(), "Error, conf-id[%d] may not be exist\n", req.GroupId)
				return
			}
			conf := resp.DataSet[0]
			attrs := []confDescribeRow{
				{Attribute: "ConfID", Content: strconv.Itoa(conf.GroupId)},
				{Attribute: "DBVersion", Content: conf.DBTypeId},
				{Attribute: "Name", Content: conf.GroupName},
				{Attribute: "Description", Content: conf.Description},
				{Attribute: "Modifiable", Content: strconv.FormatBool(conf.Modifiable)},
				{Attribute: "Zone", Content: conf.Zone},
			}
			fmt.Fprintln(ctx.Out(), "Attributes:")
			ctx.PrintList(attrs)

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
			fmt.Fprintln(ctx.Out(), "\nParameters:")
			ctx.PrintList(params)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Requried. Configuration identifier for the configuration to be described")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return getConfIDList(ctx, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// newUDBConfClone ucloud udb conf clone
func newUDBConfClone(ctx *cli.Context) *cobra.Command {
	var srcConfID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewCreateUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Create configuration file by cloning existed configuration",
		Long:  "Create configuration file by cloning existed configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(srcConfID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if *req.DBTypeId == "" {
				confIns, err := getConfByID(ctx, id, *req.ProjectId, *req.Region, *req.Zone)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				req.DBTypeId = sdk.String(confIns.DBTypeId)
			}
			req.SrcGroupId = &id
			resp, err := client.CreateUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.Out(), "conf[%d] created\n", resp.GroupId)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.DBTypeId = flags.String("db-version", "", fmt.Sprintf("Required. Version of DB. Accept values:%s", strings.Join(dbVersionList, ", ")))
	req.GroupName = flags.String("name", "", "Required. Name of configuration. It's length should be between 6 and 63")
	req.Description = flags.String("description", " ", "Optional. Description of the configuration to clone")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)
	flags.StringVar(&srcConfID, "src-conf-id", "", "Optional. The ConfID of source configuration which to be cloned from")

	command.SetFlagValues(cmd, "db-version", dbVersionList...)
	command.SetCompletion(cmd, "src-conf-id", func() []string {
		return getConfIDList(ctx, "sql", *req.ProjectId, *req.Region, *req.Zone)
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

// newUDBConfUpload ucloud udb conf upload
func newUDBConfUpload(ctx *cli.Context) *cobra.Command {
	var file string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewUploadUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Create configuration file by uploading local DB configuration file",
		Long:  "Create configuration file by uploading local DB configuration file",
		Run: func(c *cobra.Command, args []string) {
			content, err := cli.ReadFile(file)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if l := len(*req.GroupName); l < 6 || l > 63 {
				fmt.Fprintln(ctx.Out(), "Error, length of name shoud be between 6 and 63")
				return
			}
			req.Content = sdk.String(base64.StdEncoding.EncodeToString([]byte(content)))
			req.ParamGroupTypeId = sdk.Int(10)
			resp, err := client.UploadUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.Out(), "conf[%d] uploaded\n", resp.GroupId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&file, "conf-file", "", "Required. Path of local configuration file")
	req.DBTypeId = flags.String("db-version", "", fmt.Sprintf("Required. Version of DB. Accept values:%s", strings.Join(dbVersionList, ", ")))
	req.GroupName = flags.String("name", "", "Required. Name of configuration. It's length should be between 6 and 63")
	req.Description = flags.String("description", " ", "Optional. Description of the configuration to clone")
	// flags.StringVar(&subtype, "db-type", "", fmt.Sprintf("Optional. DB type. Accept values: %s", strings.Join(subtypeList, ", ")))
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-file")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("db-version")
	// cmd.MarkFlagRequired("db-type")

	command.SetFlagValues(cmd, "db-version", dbVersionList...)
	// command.SetFlagValues(cmd, "db-type", subtypeList...)
	command.SetCompletion(cmd, "conf-file", func() []string {
		return getFileList("")
	})
	return cmd
}

// newUDBConfUpdate ucloud udb conf update
func newUDBConfUpdate(ctx *cli.Context) *cobra.Command {
	var confID, key, value, file string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewUpdateUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update parameters of DB's configuration",
		Long:  "Update parameters of DB's configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupId = &id

			if key != "" && value != "" {
				req.Key = &key
				req.Value = &value
				_, err := client.UpdateUDBParamGroup(req)
				if err != nil {
					ctx.HandleError(err)
				} else {
					fmt.Printf("conf[%s]'sparameter[%s = %s] updated\n", confID, key, value)
				}
			}
			if file != "" {
				params, err := parseParam(file)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				for _, p := range params {
					req.Key = sdk.String(p.Key)
					req.Value = sdk.String(p.Value)
					_, err := client.UpdateUDBParamGroup(req)
					if err != nil {
						fmt.Printf("conf[%s]'sparameter[%s = %s] failed\n", confID, p.Key, p.Value)
						ctx.HandleError(err)
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

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration to update")
	flags.StringVar(&key, "key", "", "Optional. Key of parameter")
	flags.StringVar(&value, "value", "", "Optional. Value of parameter")
	flags.StringVar(&file, "file", "", "Optional. Path of file in which each parameter occupies one line with format 'key = value'")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getModifiableConfIDList(ctx, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "file", func() []string {
		return getFileList("")
	})

	cmd.MarkFlagRequired("conf-id")
	return cmd
}

// newUDBConfDelete ucloud udb conf delete
func newUDBConfDelete(ctx *cli.Context) *cobra.Command {
	var confID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete configuration of udb by conf-id",
		Long:  "Delete configuration of udb by conf-id",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			req.GroupId = &id
			_, err = client.DeleteUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.Out(), "conf[%s] deleted\n", confID)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of the configuration to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	command.SetCompletion(cmd, "conf-id", func() []string {
		return getModifiableConfIDList(ctx, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}

// newUDBConfApply ucloud udb conf apply
func newUDBConfApply(ctx *cli.Context) *cobra.Command {
	var confID string
	var udbIDs []string
	var restart, yes, async bool

	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewChangeUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply configuration for UDB instances",
		Long:  "Apply configuration for UDB instances",
		Run: func(c *cobra.Command, args []string) {
			req.GroupId = sdk.String(ctx.PickResourceID(confID))
			for _, idname := range udbIDs {
				req.DBId = sdk.String(ctx.PickResourceID(idname))
				_, err := client.ChangeUDBParamGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.Out(), "conf[%s] has applied for udb[%s]\n", confID, idname)
				if !restart {
					continue
				}
				ok := ctx.Confirm(yes, fmt.Sprintf("udb[%s] is about to restart, do you want to continue?", idname))
				if !ok {
					continue
				}
				restartReq := client.NewRestartUDBInstanceRequest()
				restartReq.Region = req.Region
				restartReq.Zone = req.Zone
				restartReq.ProjectId = req.ProjectId
				restartReq.DBId = req.DBId
				_, err = client.RestartUDBInstance(restartReq)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if async {
					fmt.Fprintf(ctx.Out(), "udb[%s] is restarting\n", idname)
				} else {
					text := fmt.Sprintf("udb[%s] is restarting", idname)
					ctx.Poller(describeUdbByID(ctx)).Spoll(*req.DBId, text, []string{status.UDB_FAIL, status.UDB_RUNNING})
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
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")
	cmd.MarkFlagRequired("udb-id")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getModifiableConfIDList(ctx, "", *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}

// newUDBConfDownload ucloud udb conf download
func newUDBConfDownload(ctx *cli.Context) *cobra.Command {
	var confID string
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewExtractUDBParamGroupRequest()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download UDB configuration",
		Long:  "Download UDB configuration",
		Run: func(c *cobra.Command, args []string) {
			id, err := strconv.Atoi(ctx.PickResourceID(confID))
			if err != nil {
				ctx.HandleError(err)
				return
			}

			req.GroupId = &id
			resp, err := client.ExtractUDBParamGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprint(ctx.Out(), resp.Content)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&confID, "conf-id", "", "Required. ConfID of configuration to download")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("conf-id")

	command.SetCompletion(cmd, "conf-id", func() []string {
		return getConfIDList(ctx, "sql", *req.ProjectId, *req.Region, *req.Zone)
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

func getConfByID(ctx *cli.Context, confID int, project, region, zone string) (*udb.UDBParamGroupSet, error) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	req.GroupId = &confID
	resp, err := client.DescribeUDBParamGroup(req)
	if err != nil {
		return nil, err
	}
	if len(resp.DataSet) != 1 {
		return nil, fmt.Errorf("conf-id[%d] may not exist", *req.GroupId)
	}
	return &resp.DataSet[0], nil
}

func getConfList(ctx *cli.Context, dbType, project, region, zone string) ([]udb.UDBParamGroupSet, error) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBParamGroupRequest()
	req.ClassType = &dbType
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	list := []udb.UDBParamGroupSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeUDBParamGroup(req)
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

func getModifiableConfIDList(ctx *cli.Context, dbType, project, region, zone string) []string {
	confs, err := getConfList(ctx, dbType, project, region, zone)
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

func getConfIDList(ctx *cli.Context, dbType, project, region, zone string) []string {
	confs, err := getConfList(ctx, dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, conf := range confs {
		list = append(list, fmt.Sprintf("%d/%s", conf.GroupId, conf.GroupName))
	}
	return list
}
