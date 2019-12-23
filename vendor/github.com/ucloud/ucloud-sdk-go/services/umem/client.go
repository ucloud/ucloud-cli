package umem

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UMemClient is the client of UMem
type UMemClient struct {
	*ucloud.Client
}

// NewClient will return a instance of UMemClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UMemClient {
	client := ucloud.NewClient(config, credential)
	return &UMemClient{
		client,
	}
}
