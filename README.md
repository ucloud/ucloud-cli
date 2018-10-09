## UCloud Command Line Interface 

The UCloud Command Line Interface is a tool to manage your UCloud services. It's built on the [UCloud API](https://docs.ucloud.cn/api/summary/index).

### Install UCloud CLI

You can install UCloud CLI by Homebrew/Linuxbrew, downloading executable binary file or building from the source code by yourself.

##### Homebrew(recommended)

You can use [Homebrew](https://brew.sh/) on macOS or [Linuxbrew](http://linuxbrew.sh/) on Linux. After installing Homebrew or Linuxbrew,just type the following command to complete the installation.
```
brew install ucloud
```
##### Build from source code

If you have installed golang, run the following commands to install the UCloud CLI.

```
$ mkdir -p $GOPATH/src/github.com/ucloud
$ cd $GOPATH/src/github.com/ucloud
$ git clone https://github.com/ucloud/ucloud-cli.git
$ cd ucloud-cli
$ make install
```

### Uninstall UCloud CLI

Remove the executable file /usr/local/bin/ucloud and the directory $HOME/.ucloud

### Config UCloud CLI

After install the cli, run 'ucloud init' to complete the cli configuration following the tips. Local settings will be saved in directory $HOME/.ucloud

### Auto complete
Run 'ucloud --completion' for help

#### Bash shell 
Please append the following scripts to file ~/.bash_profile or ~/.bashrc.
```
complete -C /usr/local/bin/ucloud ucloud
```

#### Zsh shell
Please append the following scripts to file ~/.zshrc.
```
autoload -U +X bashcompinit && bashcompinit
complete -F /usr/local/bin/ucloud ucloud
```