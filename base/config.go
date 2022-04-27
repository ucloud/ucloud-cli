package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

//ConfigFilePath path of config.json
var ConfigFilePath = fmt.Sprintf("%s/%s", GetConfigDir(), "config.json")

//CredentialFilePath path of credential.json
var CredentialFilePath = fmt.Sprintf("%s/%s", GetConfigDir(), "credential.json")

var CredentialFilePathInCloudShell = os.Getenv("CLOUD_SHELL_CREDENTIAL_FILE")

//LocalFileMode file mode of $HOME/ucloud/*
const LocalFileMode os.FileMode = 0600

//DefaultTimeoutSec default timeout for requesting api, 15s
const DefaultTimeoutSec = 15

//DefaultMaxRetryTimes default timeout for requesting api, 15s
const DefaultMaxRetryTimes = 3

//DefaultBaseURL location of api server
const DefaultBaseURL = "https://api.ucloud.cn/"

//DefaultProfile name of default profile
const DefaultProfile = "default"

//Version 版本号
const Version = "0.1.38"

var UserAgent = fmt.Sprintf("UCloud-CLI/%s", Version)

var InCloudShell = os.Getenv("CLOUD_SHELL") == "true"

//ConfigIns 配置实例, 程序加载时生成
var ConfigIns = &AggConfig{
	Profile:       DefaultProfile,
	BaseURL:       DefaultBaseURL,
	Timeout:       DefaultTimeoutSec,
	MaxRetryTimes: sdk.Int(DefaultMaxRetryTimes),
}

//AggConfigListIns 配置列表, 进程启动时从本地文件加载
var AggConfigListIns = &AggConfigManager{}

//ClientConfig 创建sdk client参数
var ClientConfig *sdk.Config

//AuthCredential 创建sdk client参数
var AuthCredential *CredentialConfig

//BizClient 用于调用业务接口
var BizClient *Client

//Global 全局flag
var Global GlobalFlag

//GlobalFlag 几乎所有接口都需要的参数，例如 region zone projectID
type GlobalFlag struct {
	Debug         bool
	JSON          bool
	Version       bool
	Completion    bool
	Config        bool
	Signup        bool
	Profile       string
	PublicKey     string
	PrivateKey    string
	BaseURL       string
	Timeout       int
	MaxRetryTimes int
}

//CLIConfig cli_config element
type CLIConfig struct {
	ProjectID      string `json:"project_id"`
	Region         string `json:"region"`
	Zone           string `json:"zone"`
	BaseURL        string `json:"base_url"`
	Timeout        int    `json:"timeout_sec"`
	Profile        string `json:"profile"`
	Active         bool   `json:"active"` //是否生效
	MaxRetryTimes  *int   `json:"max_retry_times"`
	AgreeUploadLog bool   `json:"agree_upload_log"`
}

//CredentialConfig credential element
type CredentialConfig struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Cookie     string `json:"cookie"`
	CSRFToken  string `json:"csrf_token"`
	Profile    string `json:"profile"`
}

//AggConfig 聚合配置 config+credential
type AggConfig struct {
	Profile        string `json:"profile"`
	Active         bool   `json:"active"`
	ProjectID      string `json:"project_id"`
	Region         string `json:"region"`
	Zone           string `json:"zone"`
	BaseURL        string `json:"base_url"`
	Timeout        int    `json:"timeout_sec"`
	PublicKey      string `json:"public_key"`
	PrivateKey     string `json:"private_key"`
	Cookie         string `json:"cookie"`
	CSRFToken      string `json:"csrf_token"`
	MaxRetryTimes  *int   `json:"max_retry_times"`
	AgreeUploadLog bool   `json:"agree_upload_log"`
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

//ConfigBaseURL 输入BaseURL
func (p *AggConfig) ConfigBaseURL() error {
	fmt.Printf("Default base-url(%s):", DefaultBaseURL)
	_, err := fmt.Scanf("%s\n", &p.BaseURL)
	if err != nil {
		return err
	}
	p.BaseURL = strings.TrimSpace(p.BaseURL)
	if len(p.BaseURL) == 0 {
		p.BaseURL = DefaultBaseURL
	}
	return nil
}

//ConfigUploadLog agree upload log or not
func (p *AggConfig) ConfigUploadLog() error {
	var input string
	fmt.Print("Do you agree to upload log in local file ~/.ucloud/cli.log to help ucloud-cli get better(yes|no):")
	_, err := fmt.Scanf("%s\n", &input)
	if err != nil {
		HandleError(err)
		return err
	}

	if str := strings.ToLower(input); str == "y" || str == "ye" || str == "yes" {
		p.AgreeUploadLog = true
	}
	return nil
}

//GetClientConfig 用来生成sdkClient
func (p *AggConfig) GetClientConfig(isDebug bool) *sdk.Config {
	clientConfig := &sdk.Config{
		Region:    p.Region,
		ProjectId: p.ProjectID,
		BaseUrl:   ClientConfig.BaseUrl,
		Timeout:   ClientConfig.Timeout,
		UserAgent: ClientConfig.UserAgent,
		LogLevel:  ClientConfig.LogLevel,
	}
	if isDebug == true {
		clientConfig.LogLevel = log.DebugLevel
	}
	return clientConfig
}

//GetCredential 用来生成SDkClient
func (p *AggConfig) GetCredential() *auth.Credential {
	return &auth.Credential{
		PublicKey:  p.PublicKey,
		PrivateKey: p.PrivateKey,
	}
}

func (p *AggConfig) copyToCLIConfig(target *CLIConfig) {
	target.Profile = p.Profile
	target.BaseURL = p.BaseURL
	target.Timeout = p.Timeout
	target.ProjectID = p.ProjectID
	target.Region = p.Region
	target.Zone = p.Zone
	target.Active = p.Active
	target.MaxRetryTimes = p.MaxRetryTimes
	target.AgreeUploadLog = p.AgreeUploadLog
}

func (p *AggConfig) copyToCredentialConfig(target *CredentialConfig) {
	target.Profile = p.Profile
	target.PrivateKey = p.PrivateKey
	target.PublicKey = p.PublicKey
	target.Cookie = p.Cookie
	target.CSRFToken = p.CSRFToken
}

//AggConfigManager 配置管理
type AggConfigManager struct {
	activeProfile string
	configs       map[string]*AggConfig
	configFile    *os.File
	credFile      *os.File
}

//NewAggConfigManager create instance
func NewAggConfigManager(cfgFile, credFile *os.File) (*AggConfigManager, error) {
	manager := &AggConfigManager{
		configs:    make(map[string]*AggConfig),
		configFile: cfgFile,
		credFile:   credFile,
	}

	err := manager.Load()
	if err != nil {
		if !os.IsNotExist(err) {
			return manager, err
		}

		aerr := adaptOldConfig()
		if aerr != nil {
			HandleError(fmt.Errorf("adapt to old config failed: %v", aerr))
			return manager, aerr
		}

		err := manager.Load()
		if err != nil {
			HandleError(fmt.Errorf("retry to load cli config failed: %v", err))
			return manager, err
		}
	}
	return manager, nil
}

//Append config to list, override if already exist the same profile
func (p *AggConfigManager) Append(config *AggConfig) error {
	if _, ok := p.configs[config.Profile]; ok {
		return fmt.Errorf("profile [%s] exists already", config.Profile)
	}

	if config.Active && config.Profile != p.activeProfile {
		if ac, ok := p.configs[p.activeProfile]; ok {
			ac.Active = false
		}
		p.activeProfile = config.Profile
	}
	p.configs[config.Profile] = config
	return p.Save()
}

//UpdateAggConfig  update AggConfig append if not exist
func (p *AggConfigManager) UpdateAggConfig(config *AggConfig) error {
	if _, ok := p.configs[config.Profile]; !ok {
		return p.Append(config)
	}

	if config.Active && config.Profile != p.activeProfile {
		if ac, ok := p.configs[p.activeProfile]; ok {
			ac.Active = false
		}
		p.activeProfile = config.Profile
	}
	return p.Save()
}

//Load AggConfigList from local file  $HOME/.ucloud/config.json+credential.json
func (p *AggConfigManager) Load() error {
	configs, err := p.parseCLIConfigs()
	if err != nil {
		return fmt.Errorf("read config failed: %v", err)
	}
	credentials, err := p.parseCredentials()
	if err != nil {
		return fmt.Errorf("read credential failed: %v", err)
	}

	//key: profile , value: CLIConfig
	configMap := make(map[string]*CLIConfig)
	for _, config := range configs {
		c := config
		configMap[config.Profile] = &c
		if config.Active {
			p.activeProfile = config.Profile
		}
	}
	credMap := make(map[string]*CredentialConfig)
	for _, cred := range credentials {
		c := cred
		credMap[cred.Profile] = &c
	}

	for profile, config := range configMap {
		cred, ok := credMap[profile]
		if !ok {
			LogError("profile: %s don't exist in credential")
			continue
		}

		p.configs[profile] = &AggConfig{
			PrivateKey:     cred.PrivateKey,
			PublicKey:      cred.PublicKey,
			Cookie:         cred.Cookie,
			CSRFToken:      cred.CSRFToken,
			Profile:        config.Profile,
			ProjectID:      config.ProjectID,
			Region:         config.Region,
			Zone:           config.Zone,
			BaseURL:        config.BaseURL,
			Timeout:        config.Timeout,
			Active:         config.Active,
			MaxRetryTimes:  config.MaxRetryTimes,
			AgreeUploadLog: config.AgreeUploadLog,
		}
	}

	if p.activeProfile == "" && len(configMap) > 0 {
		return fmt.Errorf("no active config found, run 'ucloud config list' to check")
	}
	if _, ok := credMap[p.activeProfile]; p.activeProfile != "" && !ok {
		return fmt.Errorf("profile %s's credential don't exist, run 'ucloud config list' to check", p.activeProfile)
	}

	return nil
}

type CredHeader struct {
	Key   string
	Value []string
}

type project struct {
	ProjectId   string
	ProjectName string
}

type region struct {
	Region string
	Zone   string
}

func NewInCloudShell() (*AggConfigManager, error) {
	credFile, err := os.OpenFile(CredentialFilePathInCloudShell, os.O_RDONLY, LocalFileMode)
	if err != nil {
		return nil, fmt.Errorf("open credential file error: %w", err)
	}
	data, err := ioutil.ReadAll(credFile)
	if err != nil {
		return nil, fmt.Errorf("read from credential file error: %w", err)
	}
	var creds []CredHeader
	err = json.Unmarshal(data, &creds)
	if err != nil {
		return nil, fmt.Errorf("unmarshal credential file error: %w", err)
	}

	var cookie string
	var tokenMap map[string]string
	for _, header := range creds {
		key := strings.ToLower(header.Key)
		if key == "cookie" {
			cookie = header.Value[0]
			tokenMap, err = parseCookie(header.Value[0])
		}
	}
	if err != nil {
		return nil, err
	}
	email := tokenMap["U_USER_EMAIL"]
	email = strings.ReplaceAll(email, ".", "_")
	email = strings.ReplaceAll(email, "@", "_")
	projectKey := fmt.Sprintf("c_project_%s", email)
	regionKey := fmt.Sprintf("c_last_region_%s", email)
	var proj project
	var reg region
	if _, ok := tokenMap[projectKey]; ok {
		err = json.Unmarshal([]byte(tokenMap[projectKey]), &proj)
		if err != nil {
			return nil, err
		}
	} else {
		id, name, err := getDefaultProject(cookie, tokenMap["CSRF_TOKEN"])
		if err != nil {
			return nil, fmt.Errorf("query default project error: %w", err)
		}
		proj.ProjectId = id
		proj.ProjectName = name
	}
	if _, ok := tokenMap[regionKey]; ok {
		err = json.Unmarshal([]byte(tokenMap[regionKey]), &reg)
		if err != nil {
			return nil, err
		}
	} else {
		region, zone, err := getDefaultRegion(cookie, tokenMap["CSRF_TOKEN"])
		if err != nil {
			return nil, fmt.Errorf("query default region error: %w", err)
		}
		reg.Region = region
		reg.Zone = zone
	}

	ac := &AggConfig{
		Cookie:        cookie,
		Profile:       DefaultProfile,
		Active:        true,
		BaseURL:       DefaultBaseURL,
		ProjectID:     proj.ProjectId,
		Region:        reg.Region,
		Zone:          reg.Zone,
		MaxRetryTimes: sdk.Int(DefaultMaxRetryTimes),
		CSRFToken:     tokenMap["CSRF_TOKEN"],
		Timeout:       DefaultTimeoutSec,
	}

	aggConfigs := make(map[string]*AggConfig, 0)
	aggConfigs[DefaultProfile] = ac

	return &AggConfigManager{
		activeProfile: DefaultProfile,
		configs:       aggConfigs,
	}, nil
}

func parseCookie(str string) (map[string]string, error) {
	items := strings.Split(str, ";")
	tokenMap := make(map[string]string, 0)
	for _, str := range items {
		strs := strings.SplitN(str, "=", 2)
		if len(strs) == 2 {
			v, err := url.QueryUnescape(strings.TrimSpace(strs[1]))
			if err != nil {
				return tokenMap, err
			}
			tokenMap[strings.TrimSpace(strs[0])] = v
		}
	}
	return tokenMap, nil
}

//Save configs to local file
func (p *AggConfigManager) Save() error {
	var clics []*CLIConfig
	var credcs []*CredentialConfig
	for _, aggConfig := range p.configs {
		cliConfig := &CLIConfig{}
		aggConfig.copyToCLIConfig(cliConfig)
		clics = append(clics, cliConfig)

		credConfig := &CredentialConfig{}
		aggConfig.copyToCredentialConfig(credConfig)
		credcs = append(credcs, credConfig)
	}
	aerr := WriteJSONFile(clics, p.configFile.Name())
	berr := WriteJSONFile(credcs, p.credFile.Name())

	if aerr != nil && berr != nil {
		return fmt.Errorf("save cli config failed: %v | save credentail failed: %v", aerr, berr)
	}
	if aerr != nil {
		return fmt.Errorf("save cli config failed: %v", aerr)
	}
	if berr != nil {
		return fmt.Errorf("save cerdentail failed: %v", berr)
	}
	return nil
}

//DeleteByProfile 从AggConfigList和本地文件中删除此配置
func (p *AggConfigManager) DeleteByProfile(profile string) error {
	if _, ok := p.configs[profile]; !ok {
		return fmt.Errorf("profile: %s is not exist", profile)
	}

	ac := p.configs[profile]
	if ac.Active {
		return fmt.Errorf("can't delete active profile")
	}

	delete(p.configs, profile)

	err := p.Save()
	if err != nil {
		return fmt.Errorf("delete profile %s failed: %v", profile, err)
	}
	return nil
}

//GetProfileNameList 获取所有profiles 用于ucloud config --profile 补全
func (p *AggConfigManager) GetProfileNameList() []string {
	profiles := []string{}
	for _, item := range p.configs {
		profiles = append(profiles, item.Profile)
	}
	return profiles
}

//GetAggConfigList get all profile config
func (p *AggConfigManager) GetAggConfigList() []AggConfig {
	configs := []AggConfig{}
	for _, cfg := range p.configs {
		configs = append(configs, *cfg)
	}
	return configs
}

//GetAggConfigByProfile get config of specify profile
func (p *AggConfigManager) GetAggConfigByProfile(profile string) (*AggConfig, bool) {
	if ac, ok := p.configs[profile]; ok {
		return ac, true
	}
	return nil, false
}

//GetActiveAggConfig get active agg config
func (p *AggConfigManager) GetActiveAggConfig() (*AggConfig, error) {
	if ac, ok := p.configs[p.activeProfile]; ok {
		return ac, nil
	}
	return nil, fmt.Errorf("active profile not found. see 'ucloud config list'")
}

//GetActiveAggConfigName get active config name
func (p *AggConfigManager) GetActiveAggConfigName() string {
	if ac, ok := p.configs[p.activeProfile]; ok {
		return ac.Profile
	}
	return ""
}

func (p *AggConfigManager) parseCLIConfigs() ([]CLIConfig, error) {
	var configs []CLIConfig
	rawConfig, err := ioutil.ReadAll(p.configFile)
	if err != nil {
		return nil, err
	}
	if len(rawConfig) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(rawConfig, &configs)
	if err != nil {
		return nil, fmt.Errorf("parse cli config faild: %v", err)
	}
	//特殊处理未配置max_retry_times的情况，v0.1.21之前硬编码重试次数为3
	for idx := range configs {
		if configs[idx].MaxRetryTimes == nil {
			configs[idx].MaxRetryTimes = sdk.Int(DefaultMaxRetryTimes)
		}
	}
	return configs, nil
}

func (p *AggConfigManager) parseCredentials() ([]CredentialConfig, error) {
	var credentials []CredentialConfig
	rawCred, err := ioutil.ReadAll(p.credFile)
	if err != nil {
		return nil, err
	}

	if len(rawCred) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(rawCred, &credentials)
	if err != nil {
		return nil, fmt.Errorf("parse credential failed: %v", err)
	}
	return credentials, nil
}

//ListAggConfig ucloud --config + ucloud config list
func ListAggConfig(json bool) {
	aggConfigs := AggConfigListIns.GetAggConfigList()
	for idx, ac := range aggConfigs {
		aggConfigs[idx].PrivateKey = MosaicString(ac.PrivateKey, 8, 5)
		aggConfigs[idx].PublicKey = MosaicString(ac.PublicKey, 8, 5)
	}
	if json {
		err := PrintJSON(aggConfigs, os.Stdout)
		if err != nil {
			HandleError(err)
		}
	} else {
		PrintTable(aggConfigs, []string{"Profile", "Active", "ProjectID", "Region", "Zone", "BaseURL", "Timeout", "PublicKey", "PrivateKey", "MaxRetryTimes", "AgreeUploadLog"})
	}
}

//LoadUserInfo 从~/.ucloud/user.json加载用户信息
func LoadUserInfo() (*uaccount.UserInfo, error) {
	filePath := GetConfigDir() + "/user.json"
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

//GetUserInfo from local file and remote api
func GetUserInfo() (*uaccount.UserInfo, error) {
	user, err := LoadUserInfo()
	if err == nil {
		return user, nil
	}

	req := BizClient.NewGetUserInfoRequest()
	resp, err := BizClient.GetUserInfo(req)

	if err != nil {
		return nil, err
	}

	if len(resp.DataSet) == 1 {
		user = &resp.DataSet[0]
		bytes, err := json.Marshal(user)
		if err != nil {
			return nil, err
		}
		fileFullPath := GetConfigDir() + "/user.json"
		err = ioutil.WriteFile(fileFullPath, bytes, 0600)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("GetUserInfo DataSet length: %d", len(resp.DataSet))
	}
	return user, nil
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
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		p = new(OldConfig)
		return nil
	}

	content, err := ioutil.ReadFile(ConfigFilePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, p)
	if err != nil {
		return err
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
		Profile:       DefaultProfile,
		ProjectID:     oc.ProjectID,
		Region:        oc.Region,
		Zone:          oc.Zone,
		BaseURL:       DefaultBaseURL,
		Timeout:       DefaultTimeoutSec,
		Active:        true,
		PrivateKey:    oc.PrivateKey,
		PublicKey:     oc.PublicKey,
		MaxRetryTimes: sdk.Int(DefaultMaxRetryTimes),
	}
	err = os.Rename(ConfigFilePath, ConfigFilePath+".old")
	if err != nil {
		return err
	}
	return AggConfigListIns.Append(ac)
}

//GetBizClient 初始化BizClient
func GetBizClient(ac *AggConfig) (*Client, error) {
	timeout, err := time.ParseDuration(fmt.Sprintf("%ds", ac.Timeout))
	if err != nil {
		err = fmt.Errorf("parse timeout %ds failed: %v", ac.Timeout, err)
	}
	ClientConfig = &sdk.Config{
		BaseUrl:    ac.BaseURL,
		Timeout:    timeout,
		UserAgent:  UserAgent,
		LogLevel:   log.FatalLevel,
		Region:     ac.Region,
		ProjectId:  ac.ProjectID,
		MaxRetries: *ac.MaxRetryTimes,
	}
	AuthCredential = &CredentialConfig{
		PublicKey:  ac.PublicKey,
		PrivateKey: ac.PrivateKey,
		Cookie:     ac.Cookie,
		CSRFToken:  ac.CSRFToken,
	}
	return NewClient(ClientConfig, AuthCredential), err
}

func InitConfigInCloudShell() error {
	configFile, err := os.OpenFile(ConfigFilePath, os.O_CREATE|os.O_RDONLY, LocalFileMode)
	if err != nil {
		return err
	}
	credFile, err := os.OpenFile(CredentialFilePath, os.O_CREATE|os.O_RDONLY, LocalFileMode)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(credFile)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		var credConfigs []CredentialConfig
		err = json.Unmarshal(data, &credConfigs)
		if err != nil {
			return err
		}
		if len(credConfigs) > 0 {
			cred := credConfigs[0]
			if cred.Cookie != "" && cred.CSRFToken != "" {
				return nil
			}
		}
	}

	AggConfigM, err := NewInCloudShell()
	if err != nil {
		return err
	}

	AggConfigM.credFile = credFile
	AggConfigM.configFile = configFile
	ins, err := AggConfigM.GetActiveAggConfig()
	if err != nil {
		return err
	}
	ConfigIns = ins
	bc, err := GetBizClient(ConfigIns)
	if err != nil {
		return err
	}
	BizClient = bc
	return AggConfigM.Save()
}

//InitConfig 初始化配置
func InitConfig() {
	configFile, err := os.OpenFile(ConfigFilePath, os.O_CREATE|os.O_RDONLY, LocalFileMode)
	if err != nil && !os.IsNotExist(err) {
		HandleError(err)
	}
	credFile, err := os.OpenFile(CredentialFilePath, os.O_CREATE|os.O_RDONLY, LocalFileMode)
	if err != nil && !os.IsNotExist(err) {
		HandleError(err)
	}

	AggConfigListIns, err = NewAggConfigManager(configFile, credFile)
	if err != nil {
		LogError(err.Error())
		return
	}

	var ins *AggConfig
	if Global.Profile == "" {
		ins, err = AggConfigListIns.GetActiveAggConfig()
		if err != nil && len(AggConfigListIns.GetAggConfigList()) != 0 {
			HandleError(err)
		}
	} else {
		ins, _ = AggConfigListIns.GetAggConfigByProfile(Global.Profile)
	}

	if ins != nil {
		ConfigIns = ins
	}

	mergeConfigIns(ConfigIns)
	logCmd()

	bc, err := GetBizClient(ConfigIns)
	if err != nil {
		HandleError(err)
	} else {
		BizClient = bc
	}
}

func mergeConfigIns(ins *AggConfig) {
	if Global.BaseURL != "" {
		ins.BaseURL = Global.BaseURL
	}
	if Global.Timeout != 0 {
		ins.Timeout = Global.Timeout
	}
	if Global.MaxRetryTimes != -1 {
		ins.MaxRetryTimes = sdk.Int(Global.MaxRetryTimes)
	}

	if Global.PublicKey != "" && Global.PrivateKey != "" {
		ins.PrivateKey = Global.PrivateKey
		ins.PublicKey = Global.PublicKey
	}
}

func init() {
	//配置日志
	err := initLog()
	if err != nil {
		fmt.Println(err)
	}
}
