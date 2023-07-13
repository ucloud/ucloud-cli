// Package ansi reference https://github.com/sindresorhus/ansi-escapes
package ansi

import (
	"fmt"
)

const csi = "\x1b["

const sep = ";"

// CursorLeft move cursor to the left side
var CursorLeft = fmt.Sprintf("%sG", csi)

// EraseDown Erase the screen from the current line down to the bottom of the
var EraseDown = fmt.Sprintf("%sJ", csi)

// EraseUp Erase the screen from the current line up to the top of the screen
var EraseUp = fmt.Sprintf("%s1J", csi)

// CursorUp Move cursor up a specific amount of rows.
func CursorUp(count int) string {
	return fmt.Sprintf("%s%dA", csi, count)
}

// CursorPrevLine Move cursor up a specific amount of rows.
func CursorPrevLine(count int) string {
	return fmt.Sprintf("%s%dF", csi, count)
}

// CursorTo Set the absolute position of the cursor. `x` `y` is the top left of the screen.
func CursorTo(x, y int) string {
	return fmt.Sprintf("%s%d;%dH", csi, y+1, x+1)
}
