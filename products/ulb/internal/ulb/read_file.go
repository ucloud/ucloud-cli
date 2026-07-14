package ulb

import "os"

func readFile(file string) (string, error) {
	byts, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(byts), nil
}
