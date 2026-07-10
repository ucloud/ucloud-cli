package ui

import (
	"fmt"
	"strings"
)

// Prompt asks for y/n confirmation on the process stdin/stdout.
func Prompt(text string) (bool, error) {
	if !strings.HasSuffix(text, "(y/n):") {
		text += " (y/n):"
	}
	fmt.Print(text)
	var agreeClose string
	_, err := fmt.Scanf("%s\n", &agreeClose)
	if err != nil {
		return false, err
	}
	agreeClose = strings.Trim(agreeClose, " ")
	agreeClose = strings.ToLower(agreeClose)

	return agreeClose == "y" || agreeClose == "yes", nil
}
