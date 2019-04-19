package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

//ConfigFile filename
const ConfigFile = "config.json"

//CredentialFile filename
const CredentialFile = "credential.json"

//DefaultTimeoutSec default timeout for requesting api, 15s
const DefaultTimeoutSec = 15

//DefaultBaseURL location of api server
const DefaultBaseURL = "https://api.ucloud.cn/"

//DefaultProfile name of default profile
const DefaultProfile = "default"

//Version 版本号
const Version = "0.1.15"

//ConfigIns 配置实例, 程序加载时生成
var ConfigIns = &AggConfig{}

//ClientConfig 创建sdk client参数
var ClientConfig *sdk.Config

//AuthCredential 创建sdk client参数
var AuthCredential *auth.Credential

//Global 全局flag
var Global GlobalFlag

//GlobalFlag 几乎所有接口都需要的参数，例如 region zone projectID
type GlobalFlag struct {
	Debug      bool
	JSON       bool
	Version    bool
	Completion bool
	Config     bool
	Signup     bool
}

//CLIConfig cli_config element
type CLIConfig struct {
	ProjectID string `json:"project_id"`
	Region    string `json:"region"`
	Zone      string `json:"zone"`
	BaseURL   string `json:"base_url"`
	Timeout   int    `json:"timeout_sec"`
	Profile   string `json:"profile"`
	Active    bool   `json:"active"` //是否生效
}

//CredentialConfig credential element
type CredentialConfig struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Profile    string `json:"profile"`
}

//AggConfig 聚合配置 config+credential
type AggConfig struct {
	Profile    string `json:"profile"`
	Active     bool   `json:"active"`
	ProjectID  string `json:"project_id"`
	Region     string `json:"region"`
	Zone       string `json:"zone"`
	BaseURL    string `json:"base_url"`
	Timeout    int    `json:"timeout_sec"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

//ConfigPublicKey 输入公钥
func (p *AggConfig) ConfigPublicKey() error {
	Cxt.Print("Your public-key:")
	_, err := fmt.Scanf("%s\n", &p.PublicKey)
	if err != nil {
		Cxt.Println(err)
		return err
	}
	p.PublicKey = strings.TrimSpace(p.PublicKey)
	AuthCredential.PublicKey = p.PublicKey
	return nil
}

//ConfigPrivateKey 输入私钥
func (p *AggConfig) ConfigPrivateKey() error {
	Cxt.Print("Your private-key:")
	_, err := fmt.Scanf("%s\n", &p.PrivateKey)
	if err != nil {
		Cxt.Println(err)
		return err
	}
	p.PrivateKey = strings.TrimSpace(p.PrivateKey)
	AuthCredential.PrivateKey = p.PrivateKey
	return nil
}

//GetClientConfig 用来生成sdkClient
func (p *AggConfig) GetClientConfig(isDebug bool) *sdk.Config {
	p.LoadActiveAggConfig()
	clientConfig := &sdk.Config{
		Region:    p.Region,
		ProjectId: p.ProjectID,
		BaseUrl:   ClientConfig.BaseUrl,
		Timeout:   ClientConfig.Timeout,
		UserAgent: ClientConfig.UserAgent,
		LogLevel:  ClientConfig.LogLevel,
	}
	if isDebug == true {
		clientConfig.LogLevel = 5
	}
	return clientConfig
}

//GetCredential 用来生成SDkClient
func (p *AggConfig) GetCredential() *auth.Credential {
	p.LoadActiveAggConfig()
	return &auth.Credential{
		PublicKey:  p.PublicKey,
		PrivateKey: p.PrivateKey,
	}
}

func (p *AggConfig) copyToCLIConfig(target *CLIConfig) {
	target.BaseURL = p.BaseURL
	target.Timeout = p.Timeout
	target.ProjectID = p.ProjectID
	target.Region = p.Region
	target.Zone = p.Zone
	target.Active = p.Active
	target.Profile = p.Profile
}

func (p *AggConfig) copyToCredentialConfig(target *CredentialConfig) {
	target.Profile = p.Profile
	target.PrivateKey = p.PrivateKey
	target.PublicKey = p.PublicKey
}

//Save 保存配置到本地文件，如果本地已有此配置，则更新，否则添加
func (p *AggConfig) Save() error {
	configs, err := readCLIConfigs()
	if err != nil {
		return fmt.Errorf("read config fail | %v", err)
	}
	configMap := make(map[string]*CLIConfig, 0)
	for idx := range configs {
		configMap[configs[idx].Profile] = &configs[idx]
		//确保只有一个配置有效(Active:true)
		if p.Active && p.Profile != configs[idx].Profile {
			configs[idx].Active = false
		}
	}
	if target, ok := configMap[p.Profile]; ok {
		p.copyToCLIConfig(target)
	} else {
		if p.PrivateKey == "" || p.PublicKey == "" {
			return fmt.Errorf("private-key and public_key are required for new profile")
		}
		_target := &CLIConfig{}
		p.copyToCLIConfig(_target)
		configs = append(configs, *_target)
	}

	err = WriteJSONFile(configs, ConfigFile)
	if err != nil {
		return err
	}

	credentials, err := readCredentials()
	if err != nil {
		return fmt.Errorf("read credentials fail | %v", err)
	}

	credentialMap := make(map[string]*CredentialConfig)
	for idx := range credentials {
		credentialMap[credentials[idx].Profile] = &credentials[idx]
	}
	if target, ok := credentialMap[p.Profile]; ok {
		p.copyToCredentialConfig(target)
	} else {
		_target := &CredentialConfig{}
		p.copyToCredentialConfig(_target)
		credentials = append(credentials, *_target)
	}

	return WriteJSONFile(credentials, CredentialFile)
}

//LoadActiveAggConfig 从本地文件加载有效配置
func (p *AggConfig) LoadActiveAggConfig() error {
	configs, err := readCLIConfigs()
	if err != nil {
		return fmt.Errorf("read config failed | %v", err)
	}
	credentials, err := readCredentials()
	if err != nil {
		return fmt.Errorf("read credential failed | %v", err)
	}
	var currConfig *CLIConfig
	for _, config := range configs {
		if config.Active {
			currConfig = &config
			break
		}
	}
	if currConfig == nil {
		return fmt.Errorf("no active config found, run 'ucloud config list' to check")
	}
	var currCredential *CredentialConfig
	for _, credential := range credentials {
		if credential.Profile == currConfig.Profile {
			currCredential = &credential
			break
		}
	}
	if currCredential == nil {
		return fmt.Errorf("no availavle credential")
	}

	p.Profile = currConfig.Profile
	p.PrivateKey = currCredential.PrivateKey
	p.PublicKey = currCredential.PublicKey
	p.ProjectID = currConfig.ProjectID
	p.Region = currConfig.Region
	p.Zone = currConfig.Zone
	p.BaseURL = currConfig.BaseURL
	p.Timeout = currConfig.Timeout
	p.Active = currConfig.Active

	return nil
}

//DeleteAggConfigByProfile 从本地文件中删除此配置
func DeleteAggConfigByProfile(profile string) error {
	configs, err := readCLIConfigs()
	if err != nil {
		return fmt.Errorf("read config fail | %v", err)
	}
	for idx, c := range configs {
		if c.Profile == profile {
			configs = append(configs[:idx], configs[idx+1:]...)
		}
	}
	err = WriteJSONFile(configs, ConfigFile)
	if err != nil {
		return err
	}

	credentials, err := readCredentials()
	if err != nil {
		return fmt.Errorf("read credential fail | %v", err)
	}
	for idx, c := range credentials {
		if c.Profile == profile {
			credentials = append(credentials[:idx], credentials[idx+1:]...)
		}
	}
	return WriteJSONFile(credentials, CredentialFile)
}

//GetProfileNameList 获取所有profiles 用于ucloud config --profile 补全
func GetProfileNameList() []string {
	list, err := readCredentials()
	if err != nil {
		return nil
	}
	profiles := []string{}
	for _, item := range list {
		profiles = append(profiles, item.Profile)
	}
	return profiles
}

//GetAggConfigList get all profile config
func GetAggConfigList() ([]AggConfig, error) {
	configs, err := readCLIConfigs()
	if err != nil {
		return nil, fmt.Errorf("read cli config failed | %v", err)
	}
	credentials, err := readCredentials()
	if err != nil {
		return nil, fmt.Errorf("read credential failed | %v", err)
	}
	credentialMap := map[string]*CredentialConfig{}
	for idx, c := range credentials {
		credentialMap[c.Profile] = &credentials[idx]
	}
	list := []AggConfig{}
	for _, c := range configs {
		if credentail, ok := credentialMap[c.Profile]; ok {
			aggc := AggConfig{
				Profile:    c.Profile,
				ProjectID:  c.ProjectID,
				Region:     c.Region,
				Zone:       c.Zone,
				BaseURL:    c.BaseURL,
				Timeout:    c.Timeout,
				Active:     c.Active,
				PrivateKey: credentail.PrivateKey,
				PublicKey:  credentail.PublicKey,
			}
			list = append(list, aggc)
		}
	}
	return list, nil
}

//ListAggConfig ucloud --config + ucloud config list
func ListAggConfig(json bool) {
	aggConfigs, err := GetAggConfigList()
	if err != nil {
		HandleError(err)
		return
	}
	for idx, ac := range aggConfigs {
		aggConfigs[idx].PrivateKey = MosaicString(ac.PrivateKey, 8, 5)
		aggConfigs[idx].PublicKey = MosaicString(ac.PublicKey, 8, 5)
	}
	if json {
		PrintJSON(aggConfigs, os.Stdout)
	} else {
		PrintTableS(aggConfigs)
	}
}

//GetAggConfigByProfile get config of specify profile
func GetAggConfigByProfile(profile string) (*AggConfig, error) {
	configs, err := readCLIConfigs()
	if err != nil {
		return nil, fmt.Errorf("read cli config failed | %v", err)
	}
	credentials, err := readCredentials()
	if err != nil {
		return nil, fmt.Errorf("read credential failed | %v", err)
	}
	var targetConfig *CLIConfig
	for _, config := range configs {
		if config.Profile == profile {
			targetConfig = &config
			break
		}
	}
	var targetCredential *CredentialConfig
	for _, c := range credentials {
		if c.Profile == profile {
			targetCredential = &c
			break
		}
	}
	retConfig := &AggConfig{
		Profile: profile,
	}
	if targetConfig != nil {
		retConfig.ProjectID = targetConfig.ProjectID
		retConfig.Region = targetConfig.Region
		retConfig.Zone = targetConfig.Zone
		retConfig.BaseURL = targetConfig.BaseURL
		retConfig.Timeout = targetConfig.Timeout
		retConfig.Active = targetConfig.Active
	}
	if targetCredential != nil {
		retConfig.PrivateKey = targetCredential.PrivateKey
		retConfig.PublicKey = targetCredential.PublicKey
	}
	return retConfig, nil
}

//LoadUserInfo 从~/.ucloud/user.json加载用户信息
func LoadUserInfo() (*uaccount.UserInfo, error) {
	filePath := GetConfigPath() + "/user.json"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("user.json is not exist")
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var user uaccount.UserInfo
	err = json.Unmarshal(content, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func readCLIConfigs() ([]CLIConfig, error) {
	filePath := GetConfigPath() + "/" + ConfigFile
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []CLIConfig{CLIConfig{Profile: DefaultProfile, Timeout: DefaultTimeoutSec, BaseURL: DefaultBaseURL, Active: true}}, nil
	}
	byts, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	configs := make([]CLIConfig, 0)
	err = json.Unmarshal(byts, &configs)
	if err != nil {
		return nil, fmt.Errorf("file path:%s | %v", filePath, err)
	}
	return configs, nil
}

func readCredentials() ([]CredentialConfig, error) {
	filePath := GetConfigPath() + "/" + CredentialFile
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []CredentialConfig{CredentialConfig{Profile: DefaultProfile}}, nil
	}
	byts, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	credentials := make([]CredentialConfig, 0)
	err = json.Unmarshal(byts, &credentials)
	if err != nil {
		return nil, fmt.Errorf("file path:%s | %v", filePath, err)
	}
	return credentials, nil
}

//OldConfig 0.1.7以及之前版本的配置struct
type OldConfig struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Region     string `json:"region"`
	Zone       string `json:"zone"`
	ProjectID  string `json:"project_id"`
}

//Load 从本地文件加载配置
func (p *OldConfig) Load() error {
	fileFullPath := GetConfigPath() + "/" + ConfigFile
	if _, err := os.Stat(fileFullPath); os.IsNotExist(err) {
		p = new(OldConfig)
	} else {
		content, err := ioutil.ReadFile(fileFullPath)
		if err != nil {
			return err
		}
		json.Unmarshal(content, p)
	}
	return nil
}

func adaptOldConfig() error {
	oc := &OldConfig{}
	err := oc.Load()
	if err != nil {
		return err
	}
	ac := &AggConfig{
		Profile:    DefaultProfile,
		ProjectID:  oc.ProjectID,
		Region:     oc.Region,
		Zone:       oc.Zone,
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeoutSec,
		Active:     true,
		PrivateKey: oc.PrivateKey,
		PublicKey:  oc.PublicKey,
	}
	filePath := GetConfigPath() + "/" + ConfigFile
	err = os.Rename(filePath, filePath+".old")
	if err != nil {
		return err
	}
	return ac.Save()
}

func init() {
	err := ConfigIns.LoadActiveAggConfig()
	if err != nil {
		aerr := adaptOldConfig()
		if aerr != nil {
			HandleError(aerr)
		} else {
			err := ConfigIns.LoadActiveAggConfig()
			if err != nil {
				HandleError(fmt.Errorf("LoadConfig failed | %v", err))
			}
		}
	}

	timeout, err := time.ParseDuration(fmt.Sprintf("%ds", ConfigIns.Timeout))
	if err != nil {
		HandleError(fmt.Errorf("parse timeout:%ds failed | %v", ConfigIns.Timeout, err))
	}

	ClientConfig = &sdk.Config{
		BaseUrl:   ConfigIns.BaseURL,
		Timeout:   timeout,
		UserAgent: fmt.Sprintf("UCloud-CLI/%s", Version),
		LogLevel:  log.FatalLevel,
	}

	AuthCredential = &auth.Credential{
		PublicKey:  ConfigIns.PublicKey,
		PrivateKey: ConfigIns.PrivateKey,
	}

	//bizClient 用于调用业务接口
	BizClient = NewClient(ClientConfig, AuthCredential)
}
