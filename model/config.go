package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ucloud/ucloud-cli/util"
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
	"github.com/ucloud/ucloud-sdk-go/service/uaccount/types"
)

const configFile = "config.json"

//ConfigPath 配置文件路径

//ConfigInstance 配置实例, 程序加载时生成
var ConfigInstance = &Config{}

//ClientConfig 创建sdk client参数
var ClientConfig *sdk.ClientConfig

//Credential 创建sdk client参数
var Credential *auth.Credential

// Config 全局配置
type Config struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Region     string `json:"region"`
	ProjectID  string `json:"project_id"`
}

//ConfigPublicKey 输入公钥
func (p *Config) ConfigPublicKey() error {
	context.Print("Your public-key:")
	_, err := fmt.Scanf("%s\n", &p.PublicKey)
	p.PublicKey = strings.TrimSpace(p.PublicKey)
	Credential.PublicKey = p.PublicKey
	p.SaveConfig()
	if err != nil {
		context.Println(err)
	}
	return err
}

//ConfigPrivateKey 输入私钥
func (p *Config) ConfigPrivateKey() error {
	context.Print("Your private-key:")
	_, err := fmt.Scanf("%s\n", &p.PrivateKey)
	p.PrivateKey = strings.TrimSpace(p.PrivateKey)
	Credential.PrivateKey = p.PrivateKey
	p.SaveConfig()
	if err != nil {
		context.Println(err)
	}
	return err
}

//ConfigRegion 输入默认Region
func (p *Config) ConfigRegion() error {
	p.LoadConfig()
	context.Print("Default region:")
	_, err := fmt.Scanf("%s\n", &p.Region)
	if err != nil {
		context.Println(err)
		return err
	}
	p.Region = strings.TrimSpace(p.Region)
	ClientConfig.Region = p.Region
	p.SaveConfig()
	return nil
}

//ConfigProjectID 输入默认ProjectID
func (p *Config) ConfigProjectID() error {
	p.LoadConfig()
	context.Print("Default project-id:")
	_, err := fmt.Scanf("%s\n", &p.ProjectID)
	if err != nil {
		context.Println(err)
		return err
	}
	p.ProjectID = strings.TrimSpace(p.ProjectID)
	ClientConfig.ProjectId = p.ProjectID
	p.SaveConfig()
	return nil
}

//GetClientConfig 用来生成sdkClient
func (p *Config) GetClientConfig(isDebug bool) *sdk.ClientConfig {
	p.LoadConfig()
	clientConfig := &sdk.ClientConfig{
		Region:    p.Region,
		ProjectId: p.ProjectID,
		BaseUrl:   "https://api.ucloud.cn/",
		Timeout:   10 * time.Second,
		LogLevel:  1,
	}
	if isDebug == true {
		clientConfig.LogLevel = 5
	}
	return clientConfig
}

//GetCredential 用来生成SDkClient
func (p *Config) GetCredential() *auth.Credential {
	p.LoadConfig()
	return &auth.Credential{
		PublicKey:  p.PublicKey,
		PrivateKey: p.PrivateKey,
	}
}

//ListConfig 查看配置
func (p *Config) ListConfig() error {

	tmpConfig := *p
	tmpConfig.PrivateKey = util.MosaicString(tmpConfig.PrivateKey, 8, 5)
	tmpConfig.PublicKey = util.MosaicString(tmpConfig.PublicKey, 8, 5)

	bytes, err := json.MarshalIndent(tmpConfig, "", "  ")
	if err != nil {
		return err
	}
	context.Println(string(bytes))
	return nil
}

//ClearConfig 清空配置
func (p *Config) ClearConfig() error {
	p = &Config{}
	return p.SaveConfig()
}

//SaveConfig 保存配置到本地文件，以后可以直接使用
func (p *Config) SaveConfig() error {
	bytes, err := json.Marshal(p)
	if err != nil {
		return err
	}
	fileFullPath := util.GetConfigPath() + "/" + configFile
	err = ioutil.WriteFile(fileFullPath, bytes, 0600)
	return err
}

//LoadConfig 从本地文件加载配置
func (p *Config) LoadConfig() error {
	fileFullPath := util.GetConfigPath() + "/" + configFile
	content, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return err
	}
	json.Unmarshal(content, p)
	return nil
}

//LoadUserInfo 从~/.ucloud/user.json加载用户信息
func LoadUserInfo() (*types.UserInfo, error) {
	fileFullPath := util.GetConfigPath() + "/user.json"
	content, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return nil, err
	}
	var user types.UserInfo
	err = json.Unmarshal(content, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func init() {
	ConfigInstance.LoadConfig()
	ClientConfig = &sdk.ClientConfig{
		Region:    ConfigInstance.Region,
		ProjectId: ConfigInstance.ProjectID,
		BaseUrl:   "https://api.ucloud.cn/",
		LogLevel:  1,
	}

	Credential = &auth.Credential{
		PublicKey:  ConfigInstance.PublicKey,
		PrivateKey: ConfigInstance.PrivateKey,
	}

}
