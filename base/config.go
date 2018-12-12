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
)

const configFile = "config.json"

//Version 版本号
const Version = "0.1.6"

//ConfigPath 配置文件路径

//ConfigInstance 配置实例, 程序加载时生成
var ConfigInstance = &Config{}

//ClientConfig 创建sdk client参数
var ClientConfig *sdk.Config

//Credential 创建sdk client参数
var Credential *auth.Credential

// Config 全局配置
type Config struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Region     string `json:"region"`
	Zone       string `json:"zone"`
	ProjectID  string `json:"project_id"`
}

//ConfigPublicKey 输入公钥
func (p *Config) ConfigPublicKey() error {
	Cxt.Print("Your public-key:")
	_, err := fmt.Scanf("%s\n", &p.PublicKey)
	p.PublicKey = strings.TrimSpace(p.PublicKey)
	Credential.PublicKey = p.PublicKey
	p.SaveConfig()
	if err != nil {
		Cxt.Println(err)
	}
	return err
}

//ConfigPrivateKey 输入私钥
func (p *Config) ConfigPrivateKey() error {
	Cxt.Print("Your private-key:")
	_, err := fmt.Scanf("%s\n", &p.PrivateKey)
	p.PrivateKey = strings.TrimSpace(p.PrivateKey)
	Credential.PrivateKey = p.PrivateKey
	p.SaveConfig()
	if err != nil {
		Cxt.Println(err)
	}
	return err
}

//ConfigRegion 输入默认Region
func (p *Config) ConfigRegion() error {
	p.LoadConfig()
	Cxt.Print("Default region:")
	_, err := fmt.Scanf("%s\n", &p.Region)
	if err != nil {
		Cxt.PrintErr(err)
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
	Cxt.Print("Default project-id:")
	_, err := fmt.Scanf("%s\n", &p.ProjectID)
	if err != nil {
		Cxt.Println(err)
		return err
	}
	p.ProjectID = strings.TrimSpace(p.ProjectID)
	ClientConfig.ProjectId = p.ProjectID
	p.SaveConfig()
	return nil
}

//GetClientConfig 用来生成sdkClient
func (p *Config) GetClientConfig(isDebug bool) *sdk.Config {
	p.LoadConfig()
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
func (p *Config) GetCredential() *auth.Credential {
	p.LoadConfig()
	return &auth.Credential{
		PublicKey:  p.PublicKey,
		PrivateKey: p.PrivateKey,
	}
}

//ListConfig 查看配置
func (p *Config) ListConfig(json bool) error {
	tmpConfig := *p
	tmpConfig.PrivateKey = MosaicString(tmpConfig.PrivateKey, 8, 5)
	tmpConfig.PublicKey = MosaicString(tmpConfig.PublicKey, 8, 5)

	if json {
		return PrintJSON(tmpConfig)
	}
	PrintTable([]Config{tmpConfig}, []string{"PublicKey", "PrivateKey", "Region", "Zone", "ProjectID"})
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
	fileFullPath := GetConfigPath() + "/" + configFile
	err = ioutil.WriteFile(fileFullPath, bytes, 0600)
	return err
}

//LoadConfig 从本地文件加载配置
func (p *Config) LoadConfig() error {
	fileFullPath := GetConfigPath() + "/" + configFile
	if _, err := os.Stat(fileFullPath); os.IsNotExist(err) {
		p = new(Config)
	} else {
		content, err := ioutil.ReadFile(fileFullPath)
		if err != nil {
			return err
		}
		json.Unmarshal(content, p)
	}
	return nil
}

//LoadUserInfo 从~/.ucloud/user.json加载用户信息
func LoadUserInfo() (*uaccount.UserInfo, error) {
	fileFullPath := GetConfigPath() + "/user.json"
	if _, err := os.Stat(fileFullPath); os.IsNotExist(err) {
		return new(uaccount.UserInfo), nil
	}
	content, err := ioutil.ReadFile(fileFullPath)
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

func init() {
	ConfigInstance.LoadConfig()
	timeout, _ := time.ParseDuration("15s")
	ClientConfig = &sdk.Config{
		BaseUrl:   "https://api.ucloud.cn/",
		Timeout:   timeout,
		UserAgent: fmt.Sprintf("UCloud CLI v%s", Version),
		LogLevel:  1,
	}

	Credential = &auth.Credential{
		PublicKey:  ConfigInstance.PublicKey,
		PrivateKey: ConfigInstance.PrivateKey,
	}

	//sdkClient 用于上报数据
	SdkClient = sdk.NewClient(ClientConfig, Credential)

	//bizClient 用于调用业务接口
	BizClient = NewClient(ClientConfig, Credential)
}
