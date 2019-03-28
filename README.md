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
$ git clone https://github.com/ucloud/ucloud-cli.git
$ cd ucloud-cli
$ make install
```

## Command Completion

The ucloud-cli include command completion feature and need configure it manually. 

**Bash shell** Add following scripts to  ~/.bash_profile or ~/.bashrc 

```
complete -C $(which ucloud) ucloud
```

**Zsh shell** please add following scripts to ~/.zshrc 

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

Taking create uhost in Nigeria (region: air-nigeria) and bind a public IP as an example, then configure GlobalSSH to accelerate the SSH efficiency beyond China mainland.

First to create an uhost instance:

```
$ ucloud uhost create --cpu 1 --memory 1 --password **** --image-id uimage-fya3qr

UHost:[uhost-tr1e] created successfully!
```

*Note* 

Run follow command to get details regarding the parameters to create new uhost instance.

```
$ ucloud uhost create --help
```

And suggest run the command to get the image-id first.

```
$ ucloud image list
```

Secondly, we're going to allocate an EIP and bind to the instance created above.

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
