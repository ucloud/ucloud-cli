//Inspaired by https://github.com/oclif/cli-ux

package ux

import (
	"fmt"
	"time"

	"github.com/ucloud/ucloud-cli/ansi"
)

// Spinner type
type Spinner struct {
	frames          []rune
	framesPerSecond int
	DoingText       string
	DoneText        string
	ticker          *time.Ticker
	output          string
}

// Start start rendor
func (s *Spinner) Start(doingText string) {
	if doingText != "" {
		s.DoingText = doingText
	}
	s.ticker = time.NewTicker(time.Second / time.Duration(s.framesPerSecond))
	s.render()
}

// Stop stop rendor
func (s *Spinner) Stop() {
	s.ticker.Stop()
	s.reset()
	s.output = fmt.Sprintf("%s...%s\n", s.DoingText, s.DoneText)
	fmt.Printf(s.output)
}

func (s *Spinner) reset() {
	if s.output == "" {
		return
	}
	fmt.Printf(ansi.CursorLeft + ansi.CursorUp(1) + ansi.EraseDown)
	s.output = ""
}

func (s *Spinner) render() {
	nextFrame := s.newFrameFactory()
	go func() {
		for range s.ticker.C {
			frame := nextFrame()
			s.reset()
			s.output = fmt.Sprintf("%s...%c\n", s.DoingText, frame)
			fmt.Printf(s.output)
		}
	}()
}

func (s *Spinner) newFrameFactory() func() rune {
	index := 0
	size := len(s.frames)
	return func() rune {
		char := s.frames[index%size]
		index++
		return char
	}
}

var spinnerFrames = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

// DotSpinner dot spinner
var DotSpinner = &Spinner{frames: spinnerFrames, framesPerSecond: 12, DoingText: "running", DoneText: "done"}
