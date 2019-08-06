module github.com/lixiaojun629/cobra

go 1.12

require (
	github.com/cpuguy83/go-md2man v1.0.10
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	gopkg.in/yaml.v2 v2.2.2
)

replace (
	github.com/spf13/cobra v0.0.3 => github.com/lixiaojun629/cobra v0.0.5
	github.com/spf13/pflag v1.0.3 => github.com/lixiaojun629/pflag v1.0.5
)
