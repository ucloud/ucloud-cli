package ui

import (
	"fmt"
	"io"
	"strings"
)

// Confirm prompts for a yes/no answer and reports one of three outcomes:
//
//   - (true, nil): confirmed (yes==true short-circuits, or user answered y/yes)
//   - (false, nil): declined (user answered anything else)
//   - (false, err): could not prompt — not interactive and no --yes. Callers
//     must surface err (non-zero exit) instead of silently skipping, matching
//     gcloud/aliyun: a destructive op in a pipe/CI needs an explicit --yes.
//
// The prompt appends " (y/n):" if not already present.
func Confirm(in io.Reader, out io.Writer, yes, interactive bool, text string) (bool, error) {
	if yes {
		return true, nil
	}
	if !interactive {
		return false, fmt.Errorf("refusing to prompt for confirmation in non-interactive mode; pass --yes to proceed")
	}
	if !strings.HasSuffix(text, "(y/n):") {
		text += " (y/n):"
	}
	fmt.Fprint(out, text)
	var answer string
	if _, err := fmt.Fscanf(in, "%s\n", &answer); err != nil {
		return false, nil
	}
	answer = strings.ToLower(strings.Trim(answer, " "))
	return answer == "y" || answer == "yes", nil
}
