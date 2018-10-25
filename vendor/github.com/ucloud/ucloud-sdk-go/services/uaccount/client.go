package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UAccountClient struct {
	client *ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UAccountClient {
	client := ucloud.NewClient(config, credential)
	return &UAccountClient{
		client: client,
	}
}
