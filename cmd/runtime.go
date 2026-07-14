package cmd

import (
	"fmt"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
	"github.com/ucloud/ucloud-cli/pkg/command"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type runtimeState struct {
	Configs    *platform.AggConfigManager
	Config     *platform.AggConfig
	SDKConfig  *sdk.Config
	Credential *platform.CredentialConfig
}

var activeRuntime *runtimeState
var runtimeAutoStub = true

func buildRuntimeFromBaseGlobals() *runtimeState {
	return &runtimeState{
		Configs:    platform.AggConfigListIns,
		Config:     platform.ConfigIns,
		SDKConfig:  platform.ClientConfig,
		Credential: platform.AuthCredential,
	}
}

func ensureRuntime() *runtimeState {
	if activeRuntime == nil {
		activeRuntime = buildRuntimeFromBaseGlobals()
	}
	if runtimeAutoStub && activeRuntime.SDKConfig == nil {
		activeRuntime.SDKConfig = &sdk.Config{BaseUrl: platform.DefaultBaseURL}
		platform.ClientConfig = activeRuntime.SDKConfig
	}
	if runtimeAutoStub && activeRuntime.Credential == nil {
		activeRuntime.Credential = &platform.CredentialConfig{}
		platform.AuthCredential = activeRuntime.Credential
	}
	return activeRuntime
}

func setActiveRuntimeFromBaseGlobals() {
	runtimeAutoStub = true
	activeRuntime = buildRuntimeFromBaseGlobals()
}

func runtimeDefaults() command.Defaults {
	rt := ensureRuntime()
	if rt == nil || rt.Config == nil {
		return command.Defaults{}
	}
	return command.Defaults{Region: rt.Config.Region, Zone: rt.Config.Zone, ProjectID: rt.Config.ProjectID}
}

func runtimeClientConfig() *sdk.Config {
	rt := ensureRuntime()
	if rt == nil {
		return nil
	}
	if !runtimeAutoStub && rt.SDKConfig == nil {
		panic("cmd runtime disabled for snapshot completion")
	}
	return rt.SDKConfig
}

func runtimeCredential() *auth.Credential {
	rt := ensureRuntime()
	if rt == nil {
		return platform.BuildCredentialFrom(nil)
	}
	return platform.BuildCredentialFrom(rt.Credential)
}

func attachRuntimeHandlers(sc sdk.ServiceClient) {
	rt := ensureRuntime()
	if rt == nil {
		platform.AttachHandlersWith(sc, nil, nil, nil)
		return
	}
	platform.AttachHandlersWith(sc, rt.Credential, rt.Config, rt.Configs)
}

func newServiceClient[T sdk.ServiceClient](ctor func(*sdk.Config, *auth.Credential) T) T {
	rt := ensureRuntime()
	if rt == nil || rt.SDKConfig == nil {
		panic("cmd runtime is not initialized")
	}
	client := ctor(rt.SDKConfig, platform.BuildCredentialFrom(rt.Credential))
	platform.AttachHandlersWith(client, rt.Credential, rt.Config, rt.Configs)
	return client
}

func newServiceClientForConfig[T sdk.ServiceClient](cfg *platform.AggConfig, ctor func(*sdk.Config, *auth.Credential) T) (T, error) {
	var zero T
	sdkConfig, credConfig, err := platform.BuildClientRuntime(cfg)
	if sdkConfig == nil {
		return zero, fmt.Errorf("build sdk config failed")
	}
	client := ctor(sdkConfig, platform.BuildCredentialFrom(credConfig))
	rt := ensureRuntime()
	var manager *platform.AggConfigManager
	if rt != nil {
		manager = rt.Configs
	}
	platform.AttachHandlersWith(client, credConfig, cfg, manager)
	return client, err
}
