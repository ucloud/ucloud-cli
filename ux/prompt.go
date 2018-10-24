package ux

import (
	"fmt"
	"strings"

	"github.com/ucloud/ucloud-cli/base"
)

// Prompt confirm
func Prompt(text string) (bool, error) {
	base.Cxt.Printf(text)
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
