package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/ucloud/ucloud-cli/base"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
)

//RegionService Region 服务
type RegionService interface {
	GetDefaultRegion() (string, string, error)
	FetchRegion() (map[string][]string, error)
	GetRegionList() []string
	GetZoneList(string) []string
	GetUserInfo() (*uaccount.UserInfo, error)
}

type regionService struct {
	client *base.Client
}

//NewRegionService create regionService
func NewRegionService(client *base.Client) RegionService {
	return &regionService{client}
}

func (rs *regionService) GetDefaultRegion() (string, string, error) {
	req := rs.client.NewGetRegionRequest()
	resp, err := rs.client.GetRegion(req)
	if err != nil {
		return "", "", err
	}
	for _, region := range resp.Regions {
		if region.IsDefault {
			return region.Region, region.Zone, nil
		}
	}
	return "", "", errors.New("No default region")
}

func (rs *regionService) FetchRegion() (map[string][]string, error) {
	req := rs.client.NewGetRegionRequest()
	resp, err := rs.client.GetRegion(req)
	if err != nil {
		return nil, err
	}
	regionMap := make(map[string][]string)
	for _, region := range resp.Regions {
		regionMap[region.Region] = append(regionMap[region.Region], region.Zone)
	}
	return regionMap, nil
}

func (rs *regionService) GetRegionList() []string {
	regionMap, err := rs.FetchRegion()
	if err != nil {
		return nil
	}
	list := []string{}
	for region := range regionMap {
		list = append(list, region)
	}
	return list
}

func (rs *regionService) GetZoneList(region string) []string {
	regionMap, err := rs.FetchRegion()
	if err != nil {
		return nil
	}
	list := []string{}
	if region == "" {
		for _, zones := range regionMap {
			list = append(list, zones...)
		}
	} else {
		list = regionMap[region]
	}
	return list
}

func (rs *regionService) GetUserInfo() (*uaccount.UserInfo, error) {
	req := rs.client.NewGetUserInfoRequest()
	var userInfo uaccount.UserInfo
	resp, err := rs.client.GetUserInfo(req)

	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) == 1 {
		userInfo = resp.DataSet[0]
		bytes, err := json.Marshal(userInfo)
		if err != nil {
			return nil, err
		}
		fileFullPath := base.GetConfigPath() + "/user.json"
		err = ioutil.WriteFile(fileFullPath, bytes, 0600)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("GetUserInfo DataSet length: %d", len(resp.DataSet))
	}
	return &userInfo, nil
}
