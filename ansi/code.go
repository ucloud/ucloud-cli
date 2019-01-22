//Package ansi reference https://github.com/sindresorhus/ansi-escapes
package ansi

import (
	"fmt"
)

const csi = "\x1b["

// const OSC = "\x1b]"
// const BEL = "\x07"
const sep = ";"

//CursorLeft move cursor to the left side
var CursorLeft = fmt.Sprintf("%sG", csi)

//EraseDown Erase the screen from the current line down to the bottom of the
var EraseDown = fmt.Sprintf("%sJ", csi)

func CursorUp(count int) string {
	return fmt.Sprintf("%s%dA", csi, count)
}

//CursorTo
func CursorTo(x, y int) string {
	return fmt.Sprintf("%s%d;%dH", csi, y+1, x+1)
}
