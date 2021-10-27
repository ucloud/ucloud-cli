package stepImpl

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/getgauge-contrib/gauge-go/gauge"
	"github.com/getgauge-contrib/gauge-go/testsuit"
)

var _ = gauge.Step(`Execute command: <command>`, func(command string) {
	_ = execCmd(command)
})

var _ = gauge.Step(`Extract <variable> by regexp(<pattern>): <command>`, func(variable, pattern, command string) {
	out := execCmd(command)
	matched := regexp.MustCompile(pattern).FindStringSubmatch(string(out))
	if len(matched) < 2 {
		testsuit.T.Fail(fmt.Errorf("no matched for %s: %s", pattern, string(out)))
	}
	gauge.GetScenarioStore()[variable] = matched[1]
})

var _ = gauge.Step(`Execute command with <variable>: <command>`, func(variable, command string) {
	_ = execCmd(strings.ReplaceAll(command, "$"+variable, fmt.Sprint(gauge.GetScenarioStore()[variable])))
})

func execCmd(command string) []byte {
	cmd := newCmd(command)
	out, err := cmd.CombinedOutput()
	gauge.WriteMessage(string(out))
	if err != nil {
		testsuit.T.Fail(fmt.Errorf("cmd.Run() failed with %s\n", err))
	}
	return out
}

func newCmd(command string) *exec.Cmd {
	tokens := strings.Split(command, " ")
	binary, err := exec.LookPath(tokens[0])
	if err != nil {
		testsuit.T.Fail(fmt.Errorf("can not found binary: %s", binary))
	}
	return exec.Command(binary, tokens[1:]...)
}
