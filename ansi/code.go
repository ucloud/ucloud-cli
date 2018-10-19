// Reference https://github.com/sindresorhus/ansi-escapes
package ansi

import (
	"fmt"
)

const ESC = "\x1b["
const OSC = "\x1b]"
const BEL = "\x07"
const SEP = ";"

var CursorLeft = fmt.Sprintf("%sG", ESC)
var EraseDown = fmt.Sprintf("%sJ", ESC)

func CursorUp(count int) string {
	return fmt.Sprintf("%s%dA", ESC, count)
}
