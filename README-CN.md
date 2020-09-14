[English](./README.md) | 简体中文
##  UCloud CLI 

![](./docs/_static/ucloud_cli_demo.gif)

UCloud CLI为管理UCloud平台上的资源和服务提供了一致性的操作接口，它使用[ucloud-sdk-go](https://github.com/ucloud/ucloud-sdk-go)调用[UCloud OpenAPI](https://docs.ucloud.cn/api/summary/overview)，从而实现对资源和服务的操作，兼容Linux, macOS和Windows平台 https://docs.ucloud.cn/developer/cli/index

## 在macOS或Linux平台安装UCloud-CLI

**通过Homebrew安装(在macOS平台上推荐此方式)**

[Homebrew](https://docs.brew.sh/Installation) 是macOS平台上非常流行的包管理工具，您可以通过如下命令轻松安装或升级UCloud-CLI

安装UCloud-CLI
```
brew install ucloud
```

升级到最新版本

```
brew upgrade ucloud
```

如果安装过程中遇到错误，请先执行如下命令更新Homebrew

```
brew update
```

如果问题依然存在，执行如下命令获取更多帮助

```
brew doctor
```

**基于源代码编译(需要本地安装golang)**

如果您已经安装了git和golang在您的平台上，您可以使用如下命令下载源代码并编译

```
git clone https://github.com/ucloud/ucloud-cli.git
cd ucloud-cli
make install
```

升级到最新版本
```
cd /path/to/ucloud-cli
git pull
make install
```

**下载已编译好的二进制可执行文件(Linux上如果选不到非常方便的安装方式，推荐用此办法安装)**

打开ucloud-cli的[发布页面](https://github.com/ucloud/ucloud-cli/releases)，找到适合您平台的ucloud-cli压缩包。点击链接进行下载，下载后，通过比对sha256摘要来检验下载文件未被劫持，然后把ucloud-cli可执行文件解压到$PATH环境变量包含的目录，操作命令如下：

举个例子
```
curl -OL https://github.com/ucloud/ucloud-cli/releases/download/0.1.23/ucloud-cli-linux-0.1.23-amd64.tgz
echo "b480f8621e8d0bd2c121221857029320eb49be708f4d7cb1b197cdc58b071c09 *ucloud-cli-linux-0.1.23-amd64.tgz" | shasum -c //检查下载的tar包是否被劫持，从发布页面获取sha256摘要
tar zxf ucloud-cli-linux-0.1.23-amd64.tgz -C /usr/local/bin/
```

## 在Windows平台上安装UCloud-CLI

**基于源代码编译**

从UCloud-CLI的[发布页面](https://github.com/ucloud/ucloud-cli/releases)下载源代码并解压，您也可以通过git下载源代码，打开Git Bash， 执行命令```git clone https://github.com/ucloud/ucloud-cli.git```。
切换到源代码所在的目录，编译源代码（执行命令 ```go build -mod=vendor -o ucloud.exe```），然后把可执行文件ucloud.exe所在目录添加到PATH环境变量中，具体操作可参看[文档](https://www.java.com/en/download/help/path.xml)
配置完成后，打开终端（cmd或power shell），执行命令```ucloud --version```检查是否安装成功。


**下载二进制可执行文件**

打开ucloud-cli的[发布页面](https://github.com/ucloud/ucloud-cli/releases)，找到适合您平台的ucloud-cli压缩包。点击链接进行下载并解压，然后把可执行文件ucloud.exe所在目录添加到PATH环境变量中，添加环境变量的操作可参考[文档](https://www.java.com/en/download/help/path.xml)

## 在Docker容器中使用UCloud-CLI
如果您已安装Docker, 通过如下命令拉取已打包UCloud-CLI的镜像。镜像打包[Dockerfile](./Dockerfile)
```
docker pull uhub.service.ucloud.cn/ucloudcli/ucloud-cli:source-code
```

基于此镜像创建容器
```
docker run --name ucloud-cli -it -d uhub.service.ucloud.cn/ucloudcli/ucloud-cli:source-code
```
连接到容器，开始使用UCloud-CLI
```
docker exec -it ucloud-cli zsh
```

## 开启命令补全（bash或zsh shell）

UCloud-CLI支持命令自动补全，开启后，您只需要输入命令的部分字符，然后敲击Tab键即可自动补全命令的其余字符。

**Bash shell** 

把如下代码添加到文件~/.bash_profile 或 ~/.bashrc中，然后source <~/.bash_profile|~/.bashrc>，或打开一个新终端，命令补全即生效

```
complete -C $(which ucloud) ucloud
```

**Zsh shell** 

把如下代码添加到文件~/.zshrc中，然后source ~/.zshrc，或打开一个新终端，命令补全即生效

```
autoload -U +X bashcompinit && bashcompinit
complete -F $(which ucloud) ucloud
```
Zsh内置命令bashcompinit有可能在某些操作系统中不生效，如果以上脚本不生效，尝试用如下脚本替换
```
_ucloud() {
        read -l;
        local cl="$REPLY";
        read -ln;
        local cp="$REPLY";
        reply=(`COMP_SHELL=zsh COMP_LINE="$cl" COMP_POINT="$cp" ucloud`)
}

compctl -K _ucloud ucloud
```


## 初始化配置
UCloud CLI支持多个命名配置，这些配置存储在本地文件config.json和credential.json中，位于~/.ucloud目录。
您可以使用```ucloud config add ```命令添加多个配置，使用--profile指定配置名称，或者直接在本地文件config.json和credential.json中添加配置。
在本地没有已生效的配置的情况下，```ucloud init```命令会添加一个配置并命名为default，此命令尽可能简化了配置过程，适合第一次使用UCloud CLI的时候初始化配置。

总共有10个配置项
- Profile: 配置名称, 此名称不允许重复。执行命令时可以被参数--profile覆盖
- Active: 标识此配置是否生效，生效的配置只有一个
- ProjectID: 默认项目ID，执行命令时可以被参数--project-id覆盖
- Region: 默认地域，执行命令时可以被参数--region覆盖
- Zone: 默认可用区，执行命令时可以被参数--zone覆盖
- BaseURL: 默认的UCloud Open API地址，执行命令时可以被参数--base-url覆盖
- Timeout: 默认的请求API超时时间，单位秒，执行命令是可以被参数--timeout覆盖
- PublicKey: 账户公钥，执行命令时可以被参数--public-key覆盖
- PrivateKey: 账户私钥，执行命令是可以被参数--private-key覆盖
- MaxRetryTimes: 默认最大的API请求失败重试次数，只对幂等API生效，所谓幂API等是指不会因为多次调用而产生副作用，比如释放EIP(ReleaseEIP)，执行命令时可以被参数--max-retry-times覆盖

添加或修改配置的命令如下

首次使用，初始化配置
```
$ ucloud init
```
查看所有配置
```
$ ucloud config list

Profile  Active  ProjectID   Region  Zone       BaseURL                 Timeout  PublicKey           PrivateKey          MaxRetryTimes
default  true    org-oxjwoi  cn-bj2  cn-bj2-05  https://api.ucloud.cn/  15       YSQGIZrL*****nCRQ=  jtma2eqQ*****+Avms  3
uweb     false   org-bdks4e  cn-bj2  cn-bj2-05  https://api.ucloud.cn/  15       4E9UU0Vh*****PWQ==  694581ea*****a0d45  3
```

添加配置
```
$ ucloud config add --profile <new-profie-name>  --public-key xxx --private-key xxx
```

修改某个配置的配置项

```
$ ucloud config update --profile xxx --region cn-sh2
```

更多信息，请参考命令帮助
```
$ ucloud config --help
```

## 举例说明

用UCloud CLI在尼日利亚创建数据中心创建一台主机并绑定一个外网IP，然后配置GlobalSSH加速，加速中国大陆到目的主机的SSH登陆

首先，创建云主机
```
$ ucloud uhost create --cpu 1 --memory-gb 1 --password **** --image-id uimage-fya3qr

uhost[uhost-zbuxxxx] is initializing...done
```

*备注* 

执行以下命令查看创建主机命令的各参数含义

```
$ ucloud uhost create --help
```

其次，申请一个EIP，然后绑定到刚刚创建的主机上
Secondly, we're going to allocate an EIP and then bind it to the uhost created above.

```
$ ucloud eip allocate --bandwidth-mb 1
allocate EIP[eip-xxx] IP:106.75.xx.xx  Line:BGP

$ ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx
bind EIP[eip-xxx] with uhost[uhost-xxx]
```

以上操作也可以用一个命令完成
```
$ ucloud uhost create --cpu 1 --memory-gb 1 --password **** --image-id uimage-fya3qr --create-eip-bandwidth-mb 1
```

配置GlobalSSH，然后通过GlobalSSH登陆主机

```
$ ucloud gssh create --location Washington --target-ip 152.32.140.92
gssh[uga-0psxxx] created

$ ssh root@152.32.140.92.ipssh.net
root@152.32.140.92.ipssh.net's password: password of the uhost instance
```

使用"ucloud api"命令调用任意API，根据API文档把某个API的参数依次填入。此命令比较特殊，不支持--public-key,--private-key,--debug,--profile,--timeout-sec等公共参数，如果要开启debug模式，可以设置环境变量$UCLOUD_CLI_DEBUG=on

```
$ ucloud api --Action <APIName>  --Param1 <value> --Param2 <value> ...
```
或者把API参数写到JSON文件中，举例如下
```
$ ucloud api --local-file ./create_uhost.json

//create_uhost.json文件内容
{
    "Action":"CreateUHostInstance",
    "Region":"cn-bj2",
    "Zone":"cn-bj2-02",
    "ImageId":"uimage-gk2x3x",
    "NetworkInterface": [{
        "EIP":{
            "Bandwidth":1,
            "OperatorName":"Bgp",
            "PayMode": "Bandwidth"
        }
    }],
    "LoginMode":"Password",
    "Password":"dGVzdGx4ajEy",
    "CPU":1,
    "Memory":2048,
    "Disks":[
        {
            "Size":20,
            "Type":"LOCAL_NORMAL",
            "IsBoot":"true"
        }
    ]
}
```