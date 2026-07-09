package firewall

import (
	"bufio"
	"os"
)

// parseRulesFromFile reads a rules file, one rule per line. Verbatim from
// cmd/firewall.go.
func parseRulesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
