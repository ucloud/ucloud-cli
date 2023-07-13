//Inspaired by https://github.com/oclif/cli-ux

package ux

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/ucloud/ucloud-cli/ansi"
)

const windows = "windows"

// Spinner type
type Spinner struct {
	out             io.Writer
	frames          []rune
	framesPerSecond int
	DoingText       string
	DoneText        string
	TimeoutText     string
	ticker          *time.Ticker
	output          string
}

// Start start render
func (s *Spinner) Start(doingText string) {
	if doingText != "" {
		s.DoingText = doingText
	}
	s.ticker = time.NewTicker(time.Second / time.Duration(s.framesPerSecond))
	s.render()
}

// Stop stop render
func (s *Spinner) Stop() {
	s.ticker.Stop()
	s.reset()
	output := fmt.Sprintf("%s...%s\n", s.DoingText, s.DoneText)
	fmt.Fprintf(s.out, output)
}

// Timeout stop render
func (s *Spinner) Timeout() {
	s.ticker.Stop()
	s.reset()
	output := fmt.Sprintf("%s...%s\n", s.DoingText, s.TimeoutText)
	fmt.Fprintf(s.out, output)
}

// Fail stop render
func (s *Spinner) Fail(err error) {
	s.ticker.Stop()
	s.reset()
	output := fmt.Sprintf("%s...fail: %v\n", s.DoingText, err)
	fmt.Fprintf(s.out, output)
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
		send := false
		for range s.ticker.C {
			if runtime.GOOS == windows {
				if !send {
					fmt.Printf("%s...\n", s.DoingText)
					send = true
				}
				continue
			}
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
var DotSpinner = NewDotSpinner(os.Stdout)

// NewDotSpinner get new DotSpinner instance
func NewDotSpinner(out io.Writer) *Spinner {
	return &Spinner{
		out:             out,
		frames:          spinnerFrames,
		framesPerSecond: 12,
		DoingText:       "running",
		DoneText:        "done",
		TimeoutText:     "timeout",
	}
}

// Refresh 刷新显示文本
type Refresh struct {
	out   io.Writer
	reset bool
}

// Do 刷新显示
func (r *Refresh) Do(text string) {
	if r.reset {
		fmt.Fprintf(r.out, ansi.CursorLeft+ansi.CursorUp(1)+ansi.EraseDown)
	} else {
		r.reset = true
	}
	fmt.Fprintln(r.out, text)
}

// NewRefresh create a new Refresh instance
func NewRefresh() *Refresh {
	return &Refresh{
		out: os.Stdout,
	}
}
