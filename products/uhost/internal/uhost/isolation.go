package uhost

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	cliconst "github.com/ucloud/ucloud-cli/model/cli"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newIsolationGroup ucloud uhost isolation-group
// Mirrors cmd/uhost.go NewCmdIsolation (AddCommand order: list, create, delete).
func newIsolationGroup(ctx *cli.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "isolation-group",
		Short: "List and manipulate isolation group of uhost",
		Long:  "List and manipulate isolation group of uhost",
	}
	cmd.AddCommand(newIsolationList(ctx))
	cmd.AddCommand(newIsolationCreate(ctx))
	cmd.AddCommand(newIsolationDelete(ctx))
	return cmd
}

// newIsolationCreate ucloud uhost isolation-group create
func newIsolationCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewCreateIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create isolation group instance",
		Long:  "Create isolation group instance",
		Run: func(c *cobra.Command, args []string) {
			re := regexp.MustCompile(cliconst.REGEXP_NAME)
			if !re.Match([]byte(*req.GroupName)) {
				ctx.LogError(fmt.Sprintf("group-name %s is invalid! Length 1~63, only English,Chinese,number and '-_.' are allowed", *req.GroupName))
				return
			}
			resp, err := client.CreateIsolationGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "isolation group %s created\n", resp.GroupId)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.GroupId, Action: "create", Status: "Created"})
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupName = flags.String("group-name", "", "Required. Name of isolation group. Length 1~63, only English,Chinese,number and '-_.' are allowed")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Remark = flags.String("remark", "", "Optional. Remark ok isolation group")

	cmd.MarkFlagRequired("group-name")
	return cmd
}

// newIsolationDelete ucloud uhost isolation-group delete
func newIsolationDelete(ctx *cli.Context) *cobra.Command {
	var ids []string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDeleteIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete isolation group instances",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range ids {
				id := ctx.PickResourceID(idname)
				req.GroupId = &id
				_, err := client.DeleteIsolationGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "isolation group %s deleted\n", idname)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&ids, "group-id", nil, "Required. Resource ID of isolation groups to be deleted")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("group-id")
	command.SetCompletion(cmd, "group-id", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

// newIsolationList ucloud uhost isolation-group list
func newIsolationList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewDescribeIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List isolation group of uhost",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.DescribeIsolationGroup(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			var list []isolationGroupRow
			for _, group := range resp.IsolationGroupSet {
				row := isolationGroupRow{
					ResourceID: group.GroupId,
					Name:       group.GroupName,
					Remark:     group.Remark,
				}
				var zones []string
				for _, item := range group.SpreadInfoSet {
					zones = append(zones, fmt.Sprintf("%s:%d", item.Zone, item.UHostCount))
				}
				row.UHostCount = strings.Join(zones, " ")
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.GroupId = flags.String("group-id", "", "Optional. Resource ID of isolation group to describe")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)

	command.SetCompletion(cmd, "group-id", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

// newLeaveIsolationGroup ucloud uhost leave-isolation-group
func newLeaveIsolationGroup(ctx *cli.Context) *cobra.Command {
	var uhostIds []string
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewLeaveIsolationGroupRequest()
	cmd := &cobra.Command{
		Use:   "leave-isolation-group",
		Short: "Detach uhost from its isolation group",
		Run: func(c *cobra.Command, args []string) {
			results := []cli.OpResultRow{}
			for _, idname := range uhostIds {
				id := ctx.PickResourceID(idname)
				any, err := describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(id, nil)
				if err != nil {
					ctx.LogError(fmt.Sprintf("fetch uhost %s failed: %v", idname, err))
					continue
				}
				ins, ok := any.(*uhostsdk.UHostInstanceSet)
				if !ok {
					ctx.LogError(fmt.Sprintf("uhost %s may not exist", idname))
					continue
				}
				if ins.IsolationGroup == "" {
					fmt.Fprintf(ctx.ProgressWriter(), "uhost %s doesn't attached any isolation group\n", idname)
					continue
				}
				req.GroupId = sdk.String(ins.IsolationGroup)
				req.UHostId = &id
				_, err = client.LeaveIsolationGroup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(ctx.ProgressWriter(), "uhost %s detached from isolation group %s\n", idname, ins.IsolationGroup)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "leave-isolation-group", Status: "Detached"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	flags.StringSliceVar(&uhostIds, "uhost-id", nil, "Required. Resource ID of uhosts to be detech from its isolation group")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	ctx.BindZone(cmd, req)
	cmd.MarkFlagRequired("uhost-id")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, nil, *req.ProjectId, *req.Region, *req.Zone)
	})
	return cmd
}
