module github.com/ucloud/ucloud-cli

go 1.12

require (
	github.com/kr/pretty v0.1.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/ucloud/ucloud-sdk-go v0.21.21
	golang.org/x/sys v0.0.0-20190412213103-97732733099d
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace (
	github.com/spf13/cobra v0.0.3 => github.com/lixiaojun629/cobra v0.0.10
	github.com/spf13/pflag v1.0.3 => github.com/lixiaojun629/pflag v1.0.5
)
