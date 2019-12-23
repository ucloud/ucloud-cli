package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// ULBClient is the client of ULB
type ULBClient struct {
	*ucloud.Client
}

// NewClient will return a instance of ULBClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *ULBClient {
	client := ucloud.NewClient(config, credential)
	return &ULBClient{
		client,
	}
}
