package ux

import (
	"fmt"
	"strings"
)

// Prompt confirm
func Prompt(text string) (bool, error) {
	if !strings.HasSuffix(text, "(y/n):") {
		text += " (y/n):"
	}
	fmt.Printf(text)
	var agreeClose string
	_, err := fmt.Scanf("%s\n", &agreeClose)
	if err != nil {
		return false, err
	}
	agreeClose = strings.Trim(agreeClose, " ")
	agreeClose = strings.ToLower(agreeClose)

	if agreeClose == "y" || agreeClose == "yes" {
		return true, nil
	}
	return false, nil
}
