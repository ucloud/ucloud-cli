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
	"strings"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-cli/util"
)

// NewCmdCompletion ucloud completion
func NewCmdCompletion() *cobra.Command {
	var desc = `Description:
  On macOS, using bash

  On macOS, you will need to install bash-completion support via Homebrew first:
  If running Bash 3.2 included with macOS
  > brew install bash-completion
  or, if running Bash 4.1+
  > brew install bash-completion@2
  Follow the “caveats” section of brew’s output to add the appropriate bash completion path to your local .bash_profile.
  and then generate bash completion scripts for ucloud 
  > ucloud completion
  
  On Linux, using bash
  
  On CentOS Linux, you may need to install the bash-completion package which is not installed by default.
  > yum install bash-completion -y
  and then genreate bash completion scripts for ucloud 
  > ucloud completion
  
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

func bashCompletion(cmd *cobra.Command) {
	home := util.GetHomePath()
	shellPath := home + "/" + util.ConfigPath + "/ucloud.sh"
	cmd.GenBashCompletionFile(shellPath)
	for _, rc := range [...]string{".bashrc", ".bash_profile", ".bash_login", ".profile"} {
		rcPath := home + "/" + rc
		if _, err := os.Stat(rcPath); err == nil {
			cmd := "source " + shellPath
			if util.LineInFile(rcPath, cmd) == false {
				util.AppendToFile(rcPath, cmd)
				fmt.Println("Auto completion is on. Please install bash-completion on your platform using brew,yum or apt-get. ucloud completion --help for more information")
			} else {
				fmt.Println("Auto completion update. Restart session")
			}
		}
	}
}

func zshCompletion(cmd *cobra.Command) {
	home := util.GetHomePath()
	shellPath := home + "/" + util.ConfigPath + "/_ucloud"
	file, err := os.Create(shellPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	runCompletionZsh(file, cmd)

	rcPath := home + "/.zshrc"
	if _, err = os.Stat(rcPath); err == nil {
		cmd := fmt.Sprintf("fpath=(%s/%s $fpath);", home, util.ConfigPath)
		cmd += "autoload -U +X compinit && compinit"
		if util.LineInFile(rcPath, cmd) == false {
			util.AppendToFile(rcPath, cmd)
			fmt.Println("Auto completion is on")
		} else {
			fmt.Println("Auto completion update. Restart session")
		}
	}
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
