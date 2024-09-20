module github.com/ucloud/ucloud-cli

go 1.19

require (
	github.com/fatih/color v1.13.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/ucloud/ucloud-sdk-go v0.22.25
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150
)

require (
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.8.0 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace (
	github.com/spf13/cobra v0.0.3 => github.com/lixiaojun629/cobra v0.0.10
	github.com/spf13/pflag v1.0.3 => github.com/lixiaojun629/pflag v1.0.5
)
