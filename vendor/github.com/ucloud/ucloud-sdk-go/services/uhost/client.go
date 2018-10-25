package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UHostClient struct {
	client *ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UHostClient {
	client := ucloud.NewClient(config, credential)
	return &UHostClient{
		client: client,
	}
}
