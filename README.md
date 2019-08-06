##  UCloud CLI 

![](./docs/_static/ucloud_cli_demo.gif)

The UCloud CLI provides a unified command line interface to UCloud services. It works on Golang SDK based on UCloud OpenAPI and supports Linux, macOS and Windows. 
https://docs.ucloud.cn/software/cli/index

## Installing ucloud-cli on macOS or Linux

**Using Homebrew(Preferred)**

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

**Building from source**

If you have installed git and golang on your platform, you can fetch the source code of ucloud cli from github and complie it by yourself.

```
git clone https://github.com/ucloud/ucloud-cli.git
cd ucloud-cli
make install
```

**Downloading binary release**

Vist the [releases page](https://github.com/ucloud/ucloud-cli/releases) of ucloud cli, and find the appropriate archive for your operating system and architecture.
Download the archive , check the shasum256 hashcode and extract it to your $PATH

For example
```
curl -OL https://github.com/ucloud/ucloud-cli/releases/download/0.1.20/ucloud-cli-macosx-0.1.20-amd64.tgz
echo "6953232b20f3474973cf234218097006a2e0d1d049c115da6c0e09c103762d4d *ucloud-cli-macosx-0.1.20-amd64.tgz" | shasum -c
tar zxf ucloud-cli-macosx-0.1.20-amd64.tgz -C /usr/local/bin/
```

## Installing ucloud cli on Windows

**Building from source**

Download the source code of ucloud cli from [releases page](https://github.com/ucloud/ucloud-cli/releases). You can also download it by running ```git clone https://github.com/ucloud/ucloud-cli.git```
Ensure you have git installed, because go will download packages using git. Go to the directory of the source code, and then compile the source code by running "go build -o ucloud.exe"
After that add ucloud.exe to your environment variable PATH. You could follow [this document](https://www.java.com/en/download/help/path.xml) if you don't know how to do. 
Open CMD Terminal and run ```ucloud --version ``` to test installation. 


**Downloading binary release**

Vist the [releases page](https://github.com/ucloud/ucloud-cli/releases) of ucloud cli, and find the appropriate archive for your operating system and architecture.
Download the archive , and extract it. Add binary file ucloud.exe to your environment variable PATH following [this document](https://www.java.com/en/download/help/path.xml)

## Using ucloud cli in a Docker container
If you have installed docker on your platform, pull the docker image embeded ucloud cli by follow command.
```
docker pull uhub.service.ucloud.cn/ucloudcli/ucloud-cli:0.1.20
```

The docker image was built by the Dockerfile as follow.
```
FROM ubuntu:18.04
RUN apt-get update && apt-get install wget -y
RUN wget https://github.com/ucloud/ucloud-cli/releases/download/0.1.20/ucloud-cli-linux-0.1.20-amd64.tgz
RUN tar -zxf ucloud-cli-linux-0.1.20-amd64.tgz -C /usr/local/bin/
RUN echo "complete -C $(which ucloud) ucloud" >> ~/.bashrc #enable auto completion
```

Create a docker container named ucloud-cli using the docker image your have pulled by following command.

```
docker run --name ucloud-cli -it -d uhub.service.ucloud.cn/ucloudcli/ucloud-cli:0.1.20
```
Run bash command in ucloud-cli container, and then you could play with ucloud cli.

```
docker exec -it ucloud-cli bash
```

## Enabling Shell Auto-Completion for bash or zsh shell user.

UCloud CLI also has auto-completion support. It can be set up so that if you partially type a command and then press TAB, the rest of the command is automatically filled in.

**Bash shell** 

Add following scripts to  ~/.bash_profile or ~/.bashrc 

```
complete -C $(which ucloud) ucloud
```

**Zsh shell** 

Add following scripts to ~/.zshrc 

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

Run the command below to get started and configure ucloud-cli. The private key and public key will be saved automatically and locally to directory ~/.ucloud.
You can delete the directory whenever you want.

```
$ ucloud init
```

To reset the configurations, run:

```
$ ucloud config
```

For more information, run:

```
$ ucloud config --help
```

## For example

I want to create a uhost in Nigeria (region: air-nigeria) and bind a public IP, and then configure GlobalSSH to accelerate efficiency of SSH service beyond China mainland.

Firstly, create an uhost instance:

```
$ ucloud uhost create --cpu 1 --memory 1 --password **** --image-id uimage-fya3qr

uhost[uhost-zbuxxxx] is initializing...done
```

*Note* 

Run command below to get details about the parameters of creating new uhost instance.

```
$ ucloud uhost create --help
```

Secondly, we're going to allocate an EIP and then bind it to the uhost created above.

```
$ ucloud eip allocate --line International --bandwidth 1
allocate EIP[eip-xxx] IP:106.75.xx.xx  Line:BGP

$ ucloud eip bind --eip-id eip-xxx --resource-id uhost-xxx
bind EIP[eip-xxx] with uhost[uhost-xxx]
```

Configure the GlobalSSH to the uhost instance and login the instance via GlobalSSH

```
$ ucloud gssh create --location Washington --target-ip 152.32.140.92
gssh[uga-0psxxx] created

$ ssh root@152.32.140.92.ipssh.net
root@152.32.140.92.ipssh.net's password: password of the uhost instance
```
