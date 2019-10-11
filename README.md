English | [简体中文](./README-CN.md)

##  UCloud CLI 

![](./docs/_static/ucloud_cli_demo.gif)

The UCloud CLI provides a unified command line interface to UCloud services. It works on [ucloud-sdk-go](https://github.com/ucloud/ucloud-sdk-go) based on UCloud OpenAPI and supports Linux, macOS and Windows. 
https://docs.ucloud.cn/developer/cli/index

## Installing ucloud-cli on macOS or Linux

**Using Homebrew(Recommended on macOS)**

The [Homebrew](https://docs.brew.sh/Installation) package manager may be used on macOS and Linux.
It could install ucloud-cli and its dependencies automatically by running command below.

```
brew install ucloud
```

If you have installed ucloud-cli already and want to upgrade to the latest version, just run:

```
brew upgrade ucloud
```

If you come across some errors during the installation via homebrew, please update the homebrew first and try again.

```
brew update
```

If the error is still unresolved, try the following command for help.

```
brew doctor
```

**Building from source(Recommended if you have golang installed)**

If you have installed git and golang on your platform, you can fetch the source code of ucloud cli from github and complie it by yourself.

```
git clone https://github.com/ucloud/ucloud-cli.git
cd ucloud-cli
make install
```

Upgrade to latest version

```
cd ucloud-cli
git pull
make install
```

**Downloading binary release(Recommended on Linux)**

Visit the [releases page](https://github.com/ucloud/ucloud-cli/releases) of ucloud cli, and find the appropriate archive for your operating system and architecture.
Download the archive , check the shasum256 hashcode and extract it to your $PATH

For example
```
curl -OL https://github.com/ucloud/ucloud-cli/releases/download/0.1.22/ucloud-cli-linux-0.1.22-amd64.tgz
echo "efbfb6d36d99f692b1f9cc7c9e3858047bb7b4fca6205c454098267e660b41d9 *ucloud-cli-linux-0.1.22-amd64.tgz" | shasum -c //check shasum to verify whether the downloaded tarball was hijacked. get the shasum from release page
tar zxf ucloud-cli-linux-0.1.22-amd64.tgz -C /usr/local/bin/
```

## Installing ucloud cli on Windows

**Building from source**

Download the source code of ucloud cli from [releases page](https://github.com/ucloud/ucloud-cli/releases) and extract it. You can also download it by running ```git clone https://github.com/ucloud/ucloud-cli.git```
Go to the directory of the source code, and then compile the source code by running ```go build -mod=vendor -o ucloud.exe```
After that add ucloud.exe to your environment variable PATH. You could follow [this document](https://www.java.com/en/download/help/path.xml) if you don't know how to do. 
Open CMD Terminal and run ```ucloud --version ``` to test installation. 


**Downloading binary release**

Vist the [releases page](https://github.com/ucloud/ucloud-cli/releases) of ucloud cli, and find the appropriate archive for your operating system and architecture.
Download the archive , and extract it. Add binary file ucloud.exe to your environment variable PATH following [this document](https://www.java.com/en/download/help/path.xml)

## Using ucloud cli in a Docker container
If you have installed docker on your platform, pull the docker image embedded ucloud cli by follow command. Lookup Dockerfile from [here](./Dockerfile)
```
docker pull uhub.service.ucloud.cn/ucloudcli/ucloud-cli:source-code
```

Create a docker container named ucloud-cli using the docker image your have pulled by following command.

```
docker run --name ucloud-cli -it -d uhub.service.ucloud.cn/ucloudcli/ucloud-cli:source-code
```
Run bash command in ucloud-cli container, and then you could play with ucloud cli.

```
docker exec -it ucloud-cli zsh
```

## Enabling Shell Auto-Completion for bash or zsh shell user.

UCloud CLI also has auto-completion support. It can be set up so that if you partially type a command and then press TAB, the rest of the command is automatically filled in.

**Bash shell** 

Add following scripts to  ~/.bash_profile or ~/.bashrc and then restart your terminal or run ```source <~/.bash_profile|~/.bashrc>```

```
complete -C $(which ucloud) ucloud
```

**Zsh shell** 

Add following scripts to ~/.zshrc and then restart your terminal or run ```source ~/.zshrc```

```
autoload -U +X bashcompinit && bashcompinit
complete -F $(which ucloud) ucloud
```
Zsh builtin command bashcompinit may not work on some platform. If the scripts don't work on your OS, try following scripts
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


## Setup configuration

The UCloud CLI supports using any of multiple named profiles that are stored in config.json and credential.json files which located in ~/.ucloud. 
You can configure additional profiles by using ```ucloud config add``` with the --profile flag, or by adding entries to the config.json and credential.json files.
ucloud init will add profile named default if you do not have an active profile, and it does its best to reduce configuration items for first-time use of ucloud-cli.

There are 10 configuration items

- Profile: name of the profile, duplicated names are not allowed. It can be override by --profile flag
- Active: Whether to take effect, Only one profile is active
- ProjectID: ID of default project, and it can be override by --project-id flag
- Region: default region, it can be override by --region flag
- Zone: default zone, it can be override by --zone flag
- BaseURL: default url of UCloud Open API, it can be override by --base-url flag
- Timeout: default timeout value of querying UCloud Open API, unit second. It can be override by --timeout flag
- PublicKey: public key of your account. It can be override by --public-key flag
- PrivateKey: private key of your account. It can be override by --private-key flag
- MaxRetryTimes: default max retry times for failed API request. It only works for idempotent APIs which can be called many times without side effect, for example 'ReleaseEIP', and it can be override by --max-retry-times flag

Run the command below to get started and configure ucloud-cli.
```
$ ucloud init
```
List all profiles (for example)
```
$ ucloud config list

Profile  Active  ProjectID   Region  Zone       BaseURL                 Timeout  PublicKey           PrivateKey          MaxRetryTimes
default  true    org-oxjwoi  cn-bj2  cn-bj2-05  https://api.ucloud.cn/  15       YSQGIZrL*****nCRQ=  jtma2eqQ*****+Avms  3
uweb     false   org-bdks4e  cn-bj2  cn-bj2-05  https://api.ucloud.cn/  15       4E9UU0Vh*****PWQ==  694581ea*****a0d45  3
```

Add additional profiles
```
$ ucloud config add --profile <new-profie-name>  --public-key xxx --private-key xxx
```

To change configuration items of specified profile, run:

```
$ ucloud config update --profile xxx --region cn-sh2
```

For more information, run:

```
$ ucloud config --help
```

## For example

I want to create a uhost in Nigeria (region: air-nigeria) and bind a public IP, and then configure GlobalSSH to accelerate efficiency of SSH service beyond China mainland.

Firstly, create an uhost instance:

```
$ ucloud uhost create --cpu 1 --memory-gb 1 --password **** --image-id uimage-fya3qr

uhost[uhost-zbuxxxx] is initializing...done
```

*Note* 

Run command below to get details about the parameters of creating new uhost instance.

```
$ ucloud uhost create --help
```

Secondly, we're going to allocate an EIP and then bind it to the uhost created above.

```
$ ucloud eip allocate --bandwidth-mb 1
allocate EIP[eip-xxx] IP:106.75.xx.xx  Line:BGP

$ ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx
bind EIP[eip-xxx] with uhost[uhost-xxx]
```

The operations above also can be done by one command
```
$ ucloud uhost create --cpu 1 --memory-gb 1 --password **** --image-id uimage-fya3qr --create-eip-bandwidth-mb 1
```

Configure the GlobalSSH to the uhost instance and login the instance via GlobalSSH

```
$ ucloud gssh create --location Washington --target-ip 152.32.140.92
gssh[uga-0psxxx] created

$ ssh root@152.32.140.92.ipssh.net
root@152.32.140.92.ipssh.net's password: password of the uhost instance
```
