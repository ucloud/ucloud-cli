## UCloud Command Line Interface 

The UCloud Command Line Interface is a tool to manage your UCloud services. It's built on the [UCloud API](https://docs.ucloud.cn/api/summary/index).

### Install UCloud CLI

You can install UCloud CLI by Homebrew/Linuxbrew, downloading executable binary file or building from the source code by yourself.

##### Homebrew(recommended)

You can use [Homebrew](https://brew.sh/) on macOS or [Linuxbrew](http://linuxbrew.sh/) on Linux. After installing Homebrew or Linuxbrew,just type the following command to complete the installation.
```
brew install ucloud
```

##### Download binary file
Archive links:
[Mac](http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-macosx-0.1.2-amd64.tgz)
[Linux](http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-linux-0.1.2-amd64.tgz)
[Windows](http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-windows-0.1.2-amd64.zip)

SHA-256 checksum
```
19b7a0803fc41ee689797a36fd67b288e993c383edf6087f56825a4d5bb17875 ucloud-cli-linux-0.1.2-amd64.tgz
ecc787f4045ea14d583801cd0cfa746be357d50756c2cf0ba879e405c2325d1c ucloud-cli-macosx-0.1.2-amd64.tgz
f48058ac96bb0283b18c660f0350eedba49d03a753775b0a2773b2081698b3f3 ucloud-cli-windows-0.1.2-amd64.zip
```

Download the binary file and extract to /usr/local/bin directory or add it to the $PATH. Take macOS as an example.
```
$ curl -o ucloud-cli.tgz http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-macosx-0.1.2-amd64.tgz
$ echo "ecc787f4045ea14d583801cd0cfa746be357d50756c2cf0ba879e405c2325d1c *ucloud-cli-macosx-0.1.2-amd64.tgz" | shasum -a 256 -c
$ tar -zxf ucloud-cli.tgz
$ cp ucloud /usr/local/bin
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

### Config UCloud CLI

After install the cli, run 'ucloud config' to complete the cli configuration following the tips. Local settings will be saved in directory $HOME/.ucloud
Command 'ucloud ls --object [region|project]' display all the regions and projects. You can change the default region and prject by runing 'ucloud config set [region|project] xxx'.
Execute 'ucloud config --help' for more information.

### Uninstall UCloud CLI

Remove the executable file /usr/local/bin/ucloud and the directory $HOME/.ucloud
