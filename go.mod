module github.com/ucloud/ucloud-cli

go 1.12

require (
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/ucloud/ucloud-sdk-go v0.7.3
)

replace (
	github.com/spf13/cobra v0.0.3 => github.com/lixiaojun629/cobra v0.0.5
	github.com/spf13/pflag v1.0.3 => github.com/lixiaojun629/pflag v1.0.5
)
