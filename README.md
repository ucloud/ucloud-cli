##  <u>ucloud-cli 
  
- website: https://www.ucloud.cn/

![](http://cli-ucloud-logo.sg.ufileos.com/ucloud.png)

The ucloud-cli provides a unified command line interface to manage Ucloud services. It works through Golang SDK based on UCloud OpenAPI and support Linux, macOS, and Windows. 

## Installation

The easiest way to install ucloud-cli is to use home-brew for Linux and macOS users. This will install the package as well as all dependencies.

```
$ brew install ucloud
```

If you have the ucloud-cli installed and want to upgrade to the latest version you can run:

```
$ brew upgrade ucloud
```

**Note**

If you come across error during the installation via home-brew, you may update the management package first.

```
$ brew update
```

**Build from the source code**

For windows users, suggest build from the source code which require install Golang first. This also works for Linux and macOS.

```
$ mkdir -p $GOPATH/src/github.com/ucloud
$ cd $GOPATH/src/github.com/ucloud
$ git clone https://github.com/ucloud/ucloud-cli.git
$ cd ucloud-cli
$ make install
```

## Command Completion

The ucloud-cli include command completion feature and need configure it manually. Add following scripts to  ~/.bash_profile or ~/.bashrc 

```
complete -C /usr/local/bin/ucloud ucloud
```

**Zsh shell** please add following scripts to ~/.zshrc 

```
autoload -U +X bashcompinit && bashcompinit
complete -F /usr/local/bin/ucloud ucloud
```

## Getting Started

Run the command to get started and configure ucloud-cli follow the steps. The public & private keys will be saved automatically and locally to directory ~/.ucloud.
You can delete the directory whenever you want.

```
$ ucloud init
```

To reset the configurations, run the command:

```
$ ucloud config
```

To learn the usage and flags, run the command:

```
$ ucloud help
```

## Example

Taking configure globalssh to uhost instance as an example, which will acceleare the instance SSH management efficiency (TCP 22 as default):

```
$ ucloud gssh create --location Washington --target-ip 128.14.225.161
```
