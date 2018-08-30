## UCloud Command Line Interface 

The UCloud Command Line Interface is a tool to manage your UCloud services. It's built on the [UCloud API](https://docs.ucloud.cn/api/summary/index).


### Install UCloud CLI：

You can install UCloud CLI by downloading executable binary file or building from the source code by yourself.

* Download binary file:
[Mac](https://github.com/ucloud/ucloud-cli/releases/download/0.1.1/ucloud-cli-macosx-0.1.1-amd64.tgz) [Linux](https://github.com/ucloud/ucloud-cli/releases/download/0.1.1/ucloud-cli-linux-0.1.1-amd64.tgz)
[Windows](https://github.com/ucloud/ucloud-cli/releases/download/0.1.1/ucloud-cli-windows-0.1.1-amd64.zip)

Download the binary file and extract to /usr/local/bin directory or add it to the $PATH

* Build from source code

If you have installed golang, run the following commands to install the UCloud CLI.

```
$ mkdir -p $GOPATH/src/github.com/ucloud
$ cd $GOPATH/src/github.com/ucloud
$ git clone http://github.com/ucloud/ucloud-cli.git
$ cd ucloud-cli
$ make install
```

### Config UCloud CLI

After install the cli, run 'ucloud config' to complete the cli configuration following the tips.


