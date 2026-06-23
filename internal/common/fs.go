package common

import (
	"io/ioutil" //nolint:staticcheck // verbatim copy from base; keep ioutil for zero-behavior-change
	"os"
	"runtime"
	"strings"
)

// GetHomePath returns the user's home directory.
// Verbatim from base.GetHomePath.
func GetHomePath() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// GetFileList completes file names by suffix for shell completion, reading the
// last token of COMP_LINE as the directory prefix (with ~ expansion).
// Verbatim from base.GetFileList.
func GetFileList(suffix string) []string {
	cmdLine := strings.TrimSpace(os.Getenv("COMP_LINE"))
	words := strings.Split(cmdLine, " ")
	last := words[len(words)-1]
	pathPrefix := "."

	if !strings.HasPrefix(last, "-") {
		pathPrefix = last
	}
	hasTilde := false
	//https://tiswww.case.edu/php/chet/bash/bashref.html#Tilde-Expansion
	if strings.HasPrefix(pathPrefix, "~") {
		pathPrefix = strings.Replace(pathPrefix, "~", GetHomePath(), 1)
		hasTilde = true
	}
	files, err := ioutil.ReadDir(pathPrefix)
	if err != nil {
		return nil
	}
	names := []string{}
	for _, f := range files {
		name := f.Name()
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		if hasTilde {
			pathPrefix = strings.Replace(pathPrefix, GetHomePath(), "~", 1)
		}
		if strings.HasSuffix(pathPrefix, "/") {
			names = append(names, pathPrefix+name)
		} else {
			names = append(names, pathPrefix+"/"+name)
		}
	}
	return names
}
