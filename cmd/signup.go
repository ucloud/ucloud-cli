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
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// NewCmdSignup ucloud signup
func NewCmdSignup() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "signup",
		Short:   "Launch UCloud sign up page in browser",
		Long:    `Launch UCloud sign up page in browser`,
		Args:    cobra.NoArgs,
		Example: "ucloud signup",
		Run: func(cmd *cobra.Command, args []string) {
			openbrowser("https://passport.ucloud.cn/#register")
		},
	}
	return cmd
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("Open url: %s in your browser", url)
	}
	if err != nil {
		fmt.Printf("Open url: %s in your browser\n", url)
	}
}
