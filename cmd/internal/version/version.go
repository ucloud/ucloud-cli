package version

import "fmt"

var Version = "dev"

func UserAgent() string {
	return fmt.Sprintf("UCloud-CLI/%s", Version)
}
