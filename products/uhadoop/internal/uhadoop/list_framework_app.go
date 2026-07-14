package uhadoop

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// frameworkAppResponse mirrors the real API response for
// ListUHadoopFrameworkAppByUseCase. The SDK's UseCases type has MustHas as
// string and is missing the Apps field, but the API returns both as []string.
// We bypass the SDK-generated response type and unmarshal into our own struct.
type frameworkAppResponse struct {
	response.CommonBase
	AppConfigSet []frameworkAppConfigVersion `json:"AppConfigSet"`
}

type frameworkAppConfigVersion struct {
	Framework        string                `json:"Framework"`
	FrameworkVersion string                `json:"FrameworkVersion"`
	HadoopVersion    string                `json:"HadoopVersion"`
	ReleaseVersion   string                `json:"ReleaseVersion"`
	UseCases         []frameworkUseCaseRaw `json:"UseCases"`
}

type frameworkUseCaseRaw struct {
	ClusterCase string              `json:"ClusterCase"`
	Apps        []string            `json:"Apps"`
	MustHas     []string            `json:"MustHas"`
	AppVersion  []frameworkAppEntry `json:"AppVersion"`
}

type frameworkAppEntry struct {
	AppName    string `json:"AppName"`
	AppVersion string `json:"AppVersion"`
}

// newListFrameworkApp ucloud uhadoop list-framework-app
func newListFrameworkApp(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewListUHadoopFrameworkAppByUseCaseRequest()
	cmd := &cobra.Command{
		Use:          "list-framework-app",
		Short:        "List UHadoop framework apps by use case",
		Long:         `List available UHadoop frameworks and their applications organized by use case`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if req.Region == nil || *req.Region == "" {
				ctx.HandleError(fmt.Errorf("--region is required"))
				return
			}
			if req.Zone == nil || *req.Zone == "" {
				ctx.HandleError(fmt.Errorf("--zone is required"))
				return
			}

			var resp frameworkAppResponse
			// Use InvokeAction directly with our custom response struct because
			// the SDK's typed response has MustHas as string (wrong: API returns []string).
			err := client.InvokeAction("ListUHadoopFrameworkAppByUseCase", req, &resp)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			listFrameworkApps(ctx, resp.AppConfigSet)
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
}

func listFrameworkApps(ctx *cli.Context, appConfigs []frameworkAppConfigVersion) {
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(appConfigs)
		return
	}

	list := make([]frameworkRow, 0)
	for _, ac := range appConfigs {
		for _, uc := range ac.UseCases {
			var apps []string
			var versions []string
			for _, av := range uc.AppVersion {
				apps = append(apps, av.AppName)
				versions = append(versions, av.AppName+"#"+av.AppVersion)
			}
			list = append(list, frameworkRow{
				Framework:        ac.Framework,
				FrameworkVersion: ac.FrameworkVersion,
				ReleaseVersion:   ac.ReleaseVersion,
				HadoopVersion:    ac.HadoopVersion,
				UseCase:          uc.ClusterCase,
				Apps:             strings.Join(apps, ","),
				Versions:         strings.Join(versions, ","),
				MustHas:          strings.Join(uc.MustHas, ","),
			})
		}
	}
	ctx.PrintList(list)
}

type frameworkRow struct {
	Framework        string
	FrameworkVersion string
	ReleaseVersion   string
	HadoopVersion    string
	UseCase          string
	Apps             string
	Versions         string
	MustHas          string
}
