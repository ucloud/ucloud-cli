module github.com/ucloud/ucloud-cli

go 1.12

require (
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/ucloud/ucloud-sdk-go v0.8.1-beta1
	golang.org/x/sys v0.0.0-20181205085412-a5c9d58dba9a
)

replace (
	github.com/spf13/cobra v0.0.3 => github.com/lixiaojun629/cobra v0.0.6
	github.com/spf13/pflag v1.0.3 => github.com/lixiaojun629/pflag v1.0.5
)
