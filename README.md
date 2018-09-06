## UCloud Command Line Interface 

The UCloud Command Line Interface is a tool to manage your UCloud services. It's built on the [UCloud API](https://docs.ucloud.cn/api/summary/index).

### Install UCloud CLI

You can install UCloud CLI by downloading executable binary file or building from the source code by yourself.

##### Download binary file
Archive links:
[Mac](http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-macosx-0.1.1-amd64.tgz)
[Linux](http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-linux-0.1.1-amd64.tgz)
[Windows](http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-windows-0.1.1-amd64.zip)

SHA-256 checksum
```
165f1ce4d413bf92e2794efe2722678eb80990602b81fd6e501d0c5f6bbf30bb ucloud-cli-linux-0.1.1-amd64.tgz
e174c2ef268f4b653062d0e1331bf642134a0fafbb745b407969a194d7c1bc0c ucloud-cli-macosx-0.1.1-amd64.tgz
75ff8741d9348881b3d992701590bc27f9278f207a3fb9a12ef0edfab19058d2 ucloud-cli-windows-0.1.1-amd64.zip
```

Download the binary file and extract to /usr/local/bin directory or add it to the $PATH. Take macOS as an example.
```
$ curl -o ucloud-cli.tgz http://ucloud-sdk.ufile.ucloud.com.cn/ucloud-cli-macosx-0.1.1-amd64.tgz
$ echo "e174c2ef268f4b653062d0e1331bf642134a0fafbb745b407969a194d7c1bc0c *ucloud-cli-macosx-0.1.1-amd64.tgz" | shasum -a 256 -c
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

Remove the executable file /user/local/bin/ucloud and the directory $HOME/.ucloud