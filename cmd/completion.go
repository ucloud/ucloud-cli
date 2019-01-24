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
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/base"
)

// NewCmdCompletion ucloud completion
func NewCmdCompletion() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Print the description of how to enable auto completion",
		Long:  "Print the description of how to enable auto completion",
		Run: func(cmd *cobra.Command, args []string) {
			shell, ok := os.LookupEnv("SHELL")
			if ok {
				if strings.HasSuffix(shell, "bash") {
					bashCompletion(cmd)
				} else if strings.HasSuffix(shell, "zsh") {
					zshCompletion(cmd)
				} else {
					fmt.Println("Unknow shell: %", shell)
				}
			} else {
				fmt.Println("Lookup shell failed")
			}
		},
	}
	return completionCmd
}

func bashCompletion(cmd *cobra.Command) {
	platform := runtime.GOOS
	if platform == "darwin" {
		fmt.Println(`Please append 'complete -C /usr/local/bin/ucloud ucloud' to file '~/.bash_profile'
If the following scripts are included in '~/.bash_profile', please remove it. Those scripts used to auto complete words before ucloud cli v0.1.3"

if [ -f $(brew --prefix)/etc/bash_completion ]; then
  . $(brew --prefix)/etc/bash_completion
fi
source ~/.ucloud/ucloud.sh`)

	} else if platform == "linux" {
		fmt.Println(`Please append 'complete -C /usr/local/bin/ucloud ucloud' to file '~/.bashrc'
If the following scripts are included in '~/.bashrc', please remove it. Those scripts used to auto complete words before ucloud cli v0.1.3"

if [ -f /etc/bash_completion ]; then
  . /etc/bash_completion
fi

source ~/.ucloud/ucloud.sh`)
	}
}

func zshCompletion(cmd *cobra.Command) {
	fmt.Println(`Please append the following scripts to file '~/.zshrc'.
autoload -U +X bashcompinit && bashcompinit
complete -F /usr/local/bin/ucloud ucloud`)
	fmt.Println("If the following scripts are included in '~/.bash_profile' or '~/.bashrc', please remove it. The scripts used to auto complete words before ucloud cli v0.1.3")
	fmt.Printf("fpath=(~/%s $fpath)\n", base.ConfigPath)
	fmt.Println("autoload -U +X compinit && compinit")
}

func getBashVersion() (version string, err error) {
	lookupBashVersion := exec.Command("bash", "-version")
	out, err := lookupBashVersion.Output()
	if err != nil {
		base.Cxt.PrintErr(err)
	}

	// Example
	// $ bash -version
	// GNU bash, version 3.2.57(1)-release (x86_64-apple-darwin17)
	// Copyright (C) 2007 Free Software Foundation, Inc.
	versionStr := string(out)
	re := regexp.MustCompile("(\\d)\\.\\d\\.")
	strs := re.FindAllStringSubmatch(versionStr, -1)
	if len(strs) >= 1 {
		result := strs[0]
		if len(result) >= 2 {
			version = result[1]
		}
	}
	if version == "" {
		err = fmt.Errorf("lookup bash version failed")
	}
	return
}
