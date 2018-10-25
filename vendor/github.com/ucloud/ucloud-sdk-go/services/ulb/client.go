package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type ULBClient struct {
	client *ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *ULBClient {
	client := ucloud.NewClient(config, credential)
	return &ULBClient{
		client: client,
	}
}
