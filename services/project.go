package services

import (
	"fmt"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
)

//ProjectService is a service of ucloud project
type ProjectService interface {
	GetDefaultProject() (string, string, error)
	GetProjectList() []string
}

type projectService struct {
	client *base.Client
}

//NewProjectService create project service
func NewProjectService(client *base.Client) ProjectService {
	return &projectService{client}
}

// func (ps *projectService) ListProject() error {
// 	req := &uaccount.GetProjectListRequest{}
// 	resp, err := base.BizClient.GetProjectList(req)
// 	if err != nil {
// 		return err
// 	}
// 	if resp.RetCode != 0 {
// 		return base.HandleBizError(resp)
// 	}
// 	if global.JSON {
// 		base.PrintJSON(resp.ProjectSet)
// 	} else {
// 		base.PrintTable(resp.ProjectSet, []string{"ProjectId", "ProjectName"})
// 	}
// 	return nil
// }

func (ps *projectService) GetProjectList() []string {
	req := &uaccount.GetProjectListRequest{}
	resp, err := base.BizClient.GetProjectList(req)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, p := range resp.ProjectSet {
		list = append(list, p.ProjectId+"/"+p.ProjectName)
	}
	return list
}

func (ps *projectService) GetDefaultProject() (string, string, error) {
	req := ps.client.NewGetProjectListRequest()

	resp, err := ps.client.GetProjectList(req)
	if err != nil {
		return "", "", err
	}
	for _, project := range resp.ProjectSet {
		if project.IsDefault {
			return project.ProjectId, project.ProjectName, nil
		}
	}
	return "", "", fmt.Errorf("No default project")
}
