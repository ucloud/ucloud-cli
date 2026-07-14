package ui

import "fmt"

const ansiCSI = "\x1b["

var (
	ansiCursorLeft = fmt.Sprintf("%sG", ansiCSI)
	ansiEraseDown  = fmt.Sprintf("%sJ", ansiCSI)
)

func ansiCursorUp(count int) string {
	return fmt.Sprintf("%s%dA", ansiCSI, count)
}

func ansiCursorPrevLine(count int) string {
	return fmt.Sprintf("%s%dF", ansiCSI, count)
}
