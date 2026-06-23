package cli

import (
	"github.com/ucloud/ucloud-cli/base"
	ucloud "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// NewServiceClient returns an authed SDK service client for the active profile.
// It reuses base's exact credential + handler path (BuildCredential + AttachHandlers),
// so oauth and AK/SK profiles share one code path (§9: no auth regression). ctor is
// e.g. udb.NewClient.
//
// ctx is kept in the signature per the platform contract — products call
// cli.NewServiceClient(ctx, udb.NewClient). It is intentionally unused for now.
//
// Go methods cannot have type parameters, so this is a package-level generic
// function rather than a *Context method.
func NewServiceClient[T ucloud.ServiceClient](ctx *Context, ctor func(*ucloud.Config, *auth.Credential) T) T {
	c := ctor(base.ClientConfig, base.BuildCredential())
	base.AttachHandlers(c)
	return c
}
