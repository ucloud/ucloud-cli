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
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-cli/util"
)

// NewCmdCompletion ucloud completion
func NewCmdCompletion() *cobra.Command {
	var desc = `Description:
  On macOS, using bash

  On macOS, you will need to install bash-completion support via Homebrew first:
  $ ucloud completion
  $ brew install bash-completion
  Follow the “caveats” section of brew’s output to add the appropriate bash completion path to your local .bash_profile.
  and then generate bash completion scripts for ucloud 
  
  On Linux, using bash
  
  On Linux, you may need to install the bash-completion package which is not installed by default.
  $ ucloud completion
  $ yum install bash-completion or apt-get install bash-completion
  and then genreate bash completion scripts for ucloud 
  
  Using zsh
  > ucloud completion
  
  Restart session after auto completion
`
	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generates bash/zsh completion scripts",
		Long:  desc,
		Run: func(cmd *cobra.Command, args []string) {
			shell, ok := os.LookupEnv("SHELL")
			if ok {
				if strings.HasSuffix(shell, "bash") {
					bashCompletion(cmd.Parent())
				} else if strings.HasSuffix(shell, "zsh") {
					zshCompletion(cmd.Parent())
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

var darwinBash = `
Install bash-completion with command 'brew install bash-completion', and then append the following scripts to ~/.bash_profile

if [ -f $(brew --prefix)/etc/bash_completion ]; then
  . $(brew --prefix)/etc/bash_completion
fi

source ~/.ucloud/ucloud.sh
`

var linuxBash = `
Ensure your have installed bash-completion, and then append the following scripts to ~/.bashrc

if [ -f /etc/bash_completion ]; then
  . /etc/bash_completion
fi

source ~/.ucloud/ucloud.sh
`

func getBashVersion() (version string, err error) {
	lookupBashVersion := exec.Command("bash", "-version")
	out, err := lookupBashVersion.Output()
	if err != nil {
		context.AppendError(err)
		fmt.Println(err)
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

func bashCompletion(cmd *cobra.Command) {
	home := util.GetHomePath()
	shellPath := home + "/" + util.ConfigPath + "/ucloud.sh"
	cmd.GenBashCompletionFile(shellPath)
	fmt.Printf("Completion scripts has been written to '~/%s/ucloud.sh'\n", util.ConfigPath)

	platform := runtime.GOOS

	if platform == "darwin" {
		fmt.Println(darwinBash)
	} else if platform == "linux" {
		fmt.Println(linuxBash)
	}
}

func zshCompletion(cmd *cobra.Command) {
	home := util.GetHomePath()
	shellPath := home + "/" + util.ConfigPath + "/_ucloud"
	file, err := os.Create(shellPath)
	if err != nil {
		fmt.Println(err)
		context.AppendError(err)
		return
	}
	defer file.Close()

	runCompletionZsh(file, cmd)
	fmt.Printf("Completion scripts was written to '~/%s/_ucloud'\n", util.ConfigPath)

	scripts := fmt.Sprintf("fpath=(~/%s $fpath)\n", util.ConfigPath)
	scripts += "autoload -U +X compinit && compinit"
	fmt.Printf("Please append the following scripts to your ~/.zshrc\n%s\n", scripts)
}

//参考自 k8s.io/kubernetes/pkg/kubectl/cmd/completion.go
func runCompletionZsh(out io.Writer, cmd *cobra.Command) error {
	zsh_head := "#compdef ucloud\n"

	out.Write([]byte(zsh_head))

	zsh_initialization := `
__ucloud_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand

	source "$@"
}

__ucloud_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift

		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__ucloud_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}

__ucloud_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?

	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}

__ucloud_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}

__ucloud_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}

__ucloud_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}

__ucloud_filedir() {
	local RET OLD_IFS w qw

	__ucloud_debug "_filedir $@ cur=$cur"
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi

	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"

	IFS="," __ucloud_debug "RET=${RET[@]} len=${#RET[@]}"

	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__ucloud_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}

__ucloud_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
    	printf %q "$1"
    fi
}

autoload -U +X bashcompinit && bashcompinit

# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi

__ucloud_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__ucloud_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__ucloud_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__ucloud_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__ucloud_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__ucloud_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__ucloud_type/g" \
	<<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zsh_initialization))

	buf := new(bytes.Buffer)
	cmd.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zsh_tail := `
BASH_COMPLETION_EOF
}

__ucloud_bash_source <(__ucloud_convert_bash_to_zsh)
_complete ucloud 2>/dev/null
`
	out.Write([]byte(zsh_tail))
	return nil
}
