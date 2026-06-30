// Copyright © 2018 NAME HERE tony.li@ucloud.cn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/pkg/ui"
	"github.com/ucloud/ucloud-sdk-go/ucloud/log"
)

var global = &base.Global

// NewCmdRoot 创建rootCmd rootCmd represents the base command when called without any subcommands
func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "ucloud",
		Short:             "UCloud CLI v" + base.Version,
		Long:              `UCloud CLI - manage UCloud resources and developer workflow`,
		DisableAutoGenTag: true,
		// PersistentPreRun runs the per-invocation auth/config init for the
		// executing command. Replaces the fork's OnInitialize(func(*cobra.Command))
		// (upstream OnInitialize takes func() and can't receive the command). It is
		// inherited by all subcommands (none override PersistentPreRun), so it runs
		// before every runnable command as the old OnInitialize did. The `api`
		// command keeps bypassing this via the direct-Run path in Execute().
		PersistentPreRun: func(c *cobra.Command, args []string) {
			initialize(c)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if global.Version {
				base.Cxt.Printf("ucloud cli %s\n", base.Version)
			} else if global.Completion {
				NewCmdCompletion().Run(cmd, args)
			} else if global.Config {
				base.ListAggConfig(global.JSON)
			} else if global.Signup {
				NewCmdSignup().Run(cmd, args)
			} else {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	cmd.PersistentFlags().BoolVarP(&global.Debug, "debug", "d", false, "Running in debug mode")
	cmd.PersistentFlags().BoolVarP(&global.JSON, "json", "j", false, "Print result in JSON format whenever possible")
	cmd.PersistentFlags().StringVar(&global.Output, "output", "", "Output format: table, json, or yaml. Defaults to json when stdout is not a TTY, else table")
	cmd.PersistentFlags().StringVarP(&global.Profile, "profile", "p", global.Profile, "Specifies the configuration for the operation")
	cmd.Flags().BoolVarP(&global.Version, "version", "v", false, "Display version")
	cmd.Flags().BoolVar(&global.Completion, "completion", false, "Turn on auto completion according to the prompt")
	cmd.Flags().BoolVar(&global.Config, "config", false, "Display configuration")
	cmd.Flags().BoolVar(&global.Signup, "signup", false, "Launch UCloud sign up page in browser")

	command.SetPersistentCompletion(cmd, "profile", func() []string { return base.AggConfigListIns.GetProfileNameList() })
	command.SetPersistentCompletion(cmd, "output", func() []string { return []string{"json", "table", "yaml"} })
	cmd.SetHelpTemplate(helpTmpl)
	cmd.SetUsageTemplate(usageTmpl)
	resetHelpFunc(cmd)

	return cmd
}

const helpTmpl = `Usage:{{if .Runnable}}

  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:

  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:

  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if .IsAvailableCommand}}

  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:

{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:

{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// 概要帮助信息模板
const usageTmpl = `Usage:{{if .Runnable}}
 {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}} [command] {{if $size:=len .Commands}}
 {{"command may be" | printf "%-20s"}} {{range $index,$cmd:= .Commands}}{{if .IsAvailableCommand}}{{$cmd.Name}}{{if gt $size  (add $index 1)}} | {{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableFlags}}
 {{"flags may be" | printf "%-20s"}} {{flagNames .Flags}}

Use "{{.CommandPath}} --help" for details.{{end}}
`

func newSchemaCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "__schema",
		Short:  "Print a machine-readable schema of all commands (for tools/AI)",
		Hidden: true,
		Run: func(c *cobra.Command, args []string) {
			out, err := cli.RenderSchemaJSON(c.Root())
			if err != nil {
				base.HandleError(err)
				return
			}
			fmt.Fprintln(base.Cxt.GetWriter(), out)
		},
	}
}

// productCtx is the cli.Context shared by all product commands. Its output
// format is finalized in initialize() (PersistentPreRun) after cobra parses
// --output; buildContext() runs at tree-construction time, before flag parsing,
// so the format it computes is provisional.
var productCtx *cli.Context

func addChildren(root *cobra.Command) {
	addPlatformCommands(root)
	productCtx = buildContext()
	addProductCommands(root, registeredProducts(), productCtx)
	applyGlobalOverrideFlags(root)
}

// addPlatformCommands registers all built-in platform commands onto root.
// The set and order of AddCommand calls must stay identical to preserve
// the command-tree golden (hack/snapshot/testdata/cmdtree.golden).
func addPlatformCommands(root *cobra.Command) {
	out := base.Cxt.GetWriter()
	root.AddCommand(NewCmdInit())
	root.AddCommand(NewCmdAuth())
	root.AddCommand(NewCmdDoc(out))
	root.AddCommand(NewCmdConfig())
	root.AddCommand(NewCmdRegion(out))
	root.AddCommand(NewCmdProject())
	root.AddCommand(NewCmdUHost())
	root.AddCommand(NewCmdUPHost())
	root.AddCommand(NewCmdUImage())
	root.AddCommand(NewCmdSubnet())
	root.AddCommand(NewCmdVpc())
	root.AddCommand(NewCmdFirewall())
	root.AddCommand(NewCmdDisk())
	root.AddCommand(NewCmdEIP())
	root.AddCommand(NewCmdBandwidth())
	root.AddCommand(NewCmdUDPN(out))
	root.AddCommand(NewCmdULB())
	root.AddCommand(NewCmdGssh())
	root.AddCommand(NewCmdPathx())
	root.AddCommand(NewCmdRedis())
	root.AddCommand(NewCmdMemcache())
	root.AddCommand(NewCmdExt())
	root.AddCommand(NewCmdAPI(out))
	root.AddCommand(NewCmdSignature())
	root.AddCommand(newSchemaCmd())
}

// addProductCommands registers product-package commands onto root.
// Each cli.Product contributes one top-level cobra command. This runs
// after addPlatformCommands so product commands sort after platform ones
// when cobra.EnableCommandSorting is false.
func addProductCommands(root *cobra.Command, products []cli.Product, ctx *cli.Context) {
	for _, p := range products {
		root.AddCommand(p.NewCommand(ctx))
	}
}

// applyGlobalOverrideFlags adds the five per-invocation override flags to
// every top-level command that is not in the exempt list. Running this after
// both addPlatformCommands and addProductCommands ensures product commands
// also receive the flags.
func applyGlobalOverrideFlags(root *cobra.Command) {
	for _, c := range root.Commands() {
		if c.Name() != "init" && c.Name() != "gendoc" && c.Name() != "config" && c.Name() != "auth" {
			c.PersistentFlags().StringVar(&global.PublicKey, "public-key", global.PublicKey, "Set public-key to override the public-key in local config file")
			c.PersistentFlags().StringVar(&global.PrivateKey, "private-key", global.PrivateKey, "Set private-key to override the private-key in local config file")
			c.PersistentFlags().StringVar(&global.BaseURL, "base-url", "", "Set base-url to override the base-url in local config file")
			c.PersistentFlags().IntVar(&global.Timeout, "timeout-sec", 0, "Set timeout-sec to override the timeout-sec in local config file")
			c.PersistentFlags().IntVar(&global.MaxRetryTimes, "max-retry-times", -1, "Set max-retry-times to override the max-retry-times in local config file")
		}
	}
}

// buildContext constructs the platform-level cli.Context from base globals
// and the cmd-package completion providers. Safe to call both under Execute
// (post-InitConfig) and AddChildrenForSnapshot (stubbed values).
func buildContext() *cli.Context {
	return cli.NewContext(cli.Deps{
		In:          os.Stdin,
		Out:         os.Stdout,
		Err:         os.Stderr,
		Format:      decideOutputFormat(os.Stdout),
		Config:      base.ConfigIns,
		RegionList:  getRegionList,
		ZoneList:    getZoneList,
		ProjectList: getProjectList,
		AllRegions:  getAllRegions,
	})
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Phase 3 脱敏扩面：panic 路径兜底，避免 panic 消息（可能含 token/header）原样落到 stderr
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, base.Redact(fmt.Sprintf("panic: %v", r)))
			os.Exit(1)
		}
	}()
	cmd := NewCmdRoot()
	if base.InCloudShell {
		err := base.InitConfigInCloudShell()
		if err != nil {
			base.HandleError(err)
			return
		}
	}
	base.InitConfig()
	mode := os.Getenv("UCLOUD_CLI_DEBUG")
	if mode == "on" || global.Debug {
		base.ClientConfig.LogLevel = log.DebugLevel
		base.BizClient = base.NewClient(base.ClientConfig, base.AuthCredential, base.ConfigIns)
	}

	addChildren(cmd)

	targetCmd, flags, err := cmd.Find(os.Args[1:])
	if err == nil {
		if targetCmd.Use == "api" {
			targetCmd.Run(targetCmd, flags)
			return
		}
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// usageTmpl uses the `add` template function (the command-list separator).
	// The forked cobra registered it; upstream cobra (C2) does not, so without
	// this the usage template fails to parse ("function \"add\" not defined")
	// and panics whenever it renders — e.g. on any required-flag error. Register
	// it once here so usage rendering works for every command.
	cobra.AddTemplateFunc("add", func(a, b int) int { return a + b })
	// usageTmpl also used pflag's fork-only FlagSet.FlagNames; upstream pflag
	// has no such method, so the template errored at render time. Provide an
	// equivalent template func that lists the flag names.
	cobra.AddTemplateFunc("flagNames", func(fs *pflag.FlagSet) string {
		var names []string
		fs.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
		return strings.Join(names, ", ")
	})

	//-1表示不覆盖配置文件中的MaxRetryTimes参数
	global.MaxRetryTimes = -1
	for idx, arg := range os.Args {
		if arg == "--profile" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.Profile = os.Args[idx+1]
		}
		if arg == "--public-key" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.PublicKey = os.Args[idx+1]
		}
		if arg == "--private-key" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.PrivateKey = os.Args[idx+1]
		}
		if arg == "--base-url" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			global.BaseURL = os.Args[idx+1]
		}
		if arg == "--timeout-sec" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			sec, err := strconv.Atoi(os.Args[idx+1])
			if err != nil {
				fmt.Printf("parse timeout-sec failed: %v\n", err)
			} else {
				global.Timeout = sec
			}
		}
		if arg == "--max-retry-times" && len(os.Args) > idx+1 && os.Args[idx+1] != "" {
			times, err := strconv.Atoi(os.Args[idx+1])
			if err != nil {
				fmt.Printf("parse max-retry-times failed: %v\n", err)
			} else {
				global.MaxRetryTimes = times
			}
		}
	}
	cobra.EnableCommandSorting = false
}

func resetHelpFunc(cmd *cobra.Command) {
	for _, a := range os.Args {
		if a == "-h" {
			cmd.SetHelpTemplate(usageTmpl)
		}
	}
}

func initialize(cmd *cobra.Command) {
	// Finalize the product output format now that cobra has parsed --output.
	// buildContext() ran before flag parsing, so the format it set was
	// provisional (always JSON for non-TTY stdout, ignoring an explicit
	// --output). Recompute it here so `--output table` etc. take effect.
	if productCtx != nil {
		productCtx.SetFormat(decideOutputFormat(os.Stdout))
	}

	flags := cmd.Flags()
	project, err := flags.GetString("project-id")
	if err == nil {
		base.ClientConfig.ProjectId = project
	}

	region, err := flags.GetString("region")
	if err == nil {
		base.ClientConfig.Region = region
	}

	zone, err := flags.GetString("zone")
	if err == nil {
		base.ClientConfig.Zone = zone
	}

	if isAuthSkippedCmd(cmd) {
		return
	}
	if base.InCloudShell {
		return
	}

	if base.ConfigIns.AuthMode == base.AuthModeOAuth {
		// AP-1：oauth 凭据缺失/失效 → stderr + 非零退出（不复制下方 aksk 路径的 exit 0 反模式）
		isTTY := base.IsStdinTTY()
		if msg, ok := base.CheckOAuthRunnable(base.ConfigIns, isTTY); !ok {
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}
		if err := base.EnsureFreshToken(base.ConfigIns, base.AggConfigListIns); err != nil {
			fmt.Fprintln(os.Stderr, base.OAuthRefreshFailedHint(base.ConfigIns.Profile, isTTY, err))
			os.Exit(1)
		}
		// 刷新可能换了 token，重建 client 让新 Bearer 生效。
		// GetBizClient 会重建 ClientConfig 并硬编码 FatalLevel，若 Execute 已按
		// UCLOUD_CLI_DEBUG 设了 DebugLevel，需在重建后恢复（SDK logger 在
		// NewClient 时捕获 LogLevel，必须再 rebuild 一次才生效）。
		debugOn := base.ClientConfig.LogLevel == log.DebugLevel
		bc, err := base.GetBizClient(base.ConfigIns)
		if err != nil {
			base.HandleError(err)
		} else {
			base.BizClient = bc
		}
		if debugOn {
			base.ClientConfig.LogLevel = log.DebugLevel
			base.BizClient = base.NewClient(base.ClientConfig, base.AuthCredential, base.ConfigIns)
		}
		return
	}

	// 既有 AK/SK 检查，原样保留（CRITICAL 回归约束：行为与文案零变化）
	if base.ConfigIns.PrivateKey == "" {
		base.Cxt.Println("private-key is empty. Execute command 'ucloud init|config' to configure it or run 'ucloud config list' to check your configurations")
		os.Exit(0)
	}
	if base.ConfigIns.PublicKey == "" {
		base.Cxt.Println("public-key is empty. Execute command 'ucloud init|config' to configure it or run 'ucloud config list' to check your configurations")
		os.Exit(0)
	}
}

// decideOutputFormat resolves the effective output format: explicit --output
// wins; then legacy --json; otherwise JSON for non-TTY stdout, Table for TTY.
func decideOutputFormat(out io.Writer) cli.OutputFormat {
	switch strings.ToLower(global.Output) {
	case "json":
		return cli.OutputJSON
	case "yaml":
		return cli.OutputYAML
	case "table":
		return cli.OutputTable
	}
	if global.JSON {
		return cli.OutputJSON
	}
	if ui.IsTTY(out) {
		return cli.OutputTable
	}
	return cli.OutputJSON
}

// isAuthSkippedCmd 启动凭据检查跳过清单（D7：login/logout/help/version/config/init）
func isAuthSkippedCmd(cmd *cobra.Command) bool {
	if cmd.Parent() == nil {
		return true // root 命令本身（--version/--config/help），与历史行为一致
	}
	switch cmd.Name() {
	case "config", "init", "version", "login", "logout", "help", "auth", "__schema":
		return true
	}
	if cmd.Parent() != nil && (cmd.Parent().Name() == "config" || cmd.Parent().Name() == "auth") {
		return true
	}
	return false
}
