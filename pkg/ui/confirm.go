package ui

import (
	"fmt"
	"io"
	"strings"
)

// Confirm prompts the user with text and reads a yes/no answer.
// If yes is true it returns true immediately without reading from in.
// The prompt appends " (y/n):" if it is not already present.
// Returns true only for answers "y" or "yes" (case-insensitive).
func Confirm(in io.Reader, out io.Writer, yes bool, text string) bool {
	if yes {
		return true
	}
	if !strings.HasSuffix(text, "(y/n):") {
		text += " (y/n):"
	}
	fmt.Fprint(out, text)
	var answer string
	if _, err := fmt.Fscanf(in, "%s\n", &answer); err != nil {
		return false
	}
	answer = strings.ToLower(strings.Trim(answer, " "))
	return answer == "y" || answer == "yes"
}
