package ui

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/ucloud/ucloud-cli/ansi"
)

const windows = "windows"

var spinnerFrames = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

// Spinner renders an animated single-line spinner to a writer.
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

func (s *Spinner) Start(doingText string) {
	if doingText != "" {
		s.DoingText = doingText
	}
	s.ticker = time.NewTicker(time.Second / time.Duration(s.framesPerSecond))
	s.render()
}

func (s *Spinner) Stop() {
	s.ticker.Stop()
	s.reset()
	fmt.Fprintf(s.out, "%s...%s\n", s.DoingText, s.DoneText)
}

func (s *Spinner) Timeout() {
	s.ticker.Stop()
	s.reset()
	fmt.Fprintf(s.out, "%s...%s\n", s.DoingText, s.TimeoutText)
}

func (s *Spinner) Fail(err error) {
	s.ticker.Stop()
	s.reset()
	fmt.Fprintf(s.out, "%s...fail: %v\n", s.DoingText, err)
}

func (s *Spinner) reset() {
	if s.output == "" {
		return
	}
	fmt.Fprint(s.out, ansi.CursorLeft+ansi.CursorUp(1)+ansi.EraseDown)
	s.output = ""
}

func (s *Spinner) render() {
	nextFrame := s.newFrameFactory()
	go func() {
		send := false
		for range s.ticker.C {
			if runtime.GOOS == windows {
				if !send {
					fmt.Fprintf(s.out, "%s...\n", s.DoingText)
					send = true
				}
				continue
			}
			frame := nextFrame()
			s.reset()
			s.output = fmt.Sprintf("%s...%c\n", s.DoingText, frame)
			fmt.Fprint(s.out, s.output)
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

// Spin renders spinner frames as strings for a Document block.
type Spin struct {
	out             io.Writer
	frames          []rune
	framesPerSecond int
	DoingText       string
	DoneText        string
	TimeoutText     string
	ticker          *time.Ticker
	output          string
	textChan        chan string
	wg              sync.WaitGroup
}

func (s *Spin) Stop() {
	s.ticker.Stop()
	s.reset()
	s.textChan <- fmt.Sprintf("%s...%s", s.DoingText, s.DoneText)
	<-time.After(time.Millisecond * 100)
	close(s.textChan)
}

func (s *Spin) Timeout() {
	s.ticker.Stop()
	s.reset()
	s.textChan <- fmt.Sprintf("%s...%s", s.DoingText, s.TimeoutText)
	<-time.After(time.Millisecond * 100)
	close(s.textChan)
}

func (s *Spin) reset() {
	if s.output == "" {
		return
	}
	fmt.Fprint(s.out, ansi.CursorLeft+ansi.CursorUp(1)+ansi.EraseDown)
	s.output = ""
}

func (s *Spin) renderToString() chan string {
	nextFrame := s.newFrameFactory()
	go func() {
		send := false
		for range s.ticker.C {
			if runtime.GOOS == windows {
				if !send {
					s.textChan <- fmt.Sprintf("%s...", s.DoingText)
					send = true
				}
				continue
			}
			s.textChan <- fmt.Sprintf("%s...%c", s.DoingText, nextFrame())
		}
	}()
	return s.textChan
}

func (s *Spin) newFrameFactory() func() rune {
	index := 0
	size := len(s.frames)
	return func() rune {
		char := s.frames[index%size]
		index++
		return char
	}
}

func NewDotSpin(out io.Writer, doingText string) *Spin {
	s := &Spin{
		out:             out,
		frames:          spinnerFrames,
		framesPerSecond: 12,
		DoingText:       doingText,
		DoneText:        "done",
		TimeoutText:     "timeout",
		textChan:        make(chan string),
	}
	s.ticker = time.NewTicker(time.Second / time.Duration(s.framesPerSecond))
	return s
}
