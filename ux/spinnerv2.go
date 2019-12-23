//Inspaired by https://github.com/oclif/cli-ux

package ux

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"

	"github.com/ucloud/ucloud-cli/ansi"
)

// Spin type
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

// Stop stop render
func (s *Spin) Stop() {
	s.ticker.Stop()
	s.reset()
	output := fmt.Sprintf("%s...%s", s.DoingText, s.DoneText)
	s.textChan <- output
	//等待最后一帧渲染
	<-time.After(time.Millisecond * 100)
	close(s.textChan)
}

// Timeout  stop render
func (s *Spin) Timeout() {
	s.ticker.Stop()
	s.reset()
	output := fmt.Sprintf("%s...%s", s.DoingText, s.TimeoutText)
	s.textChan <- output
	//等待最后一帧渲染
	<-time.After(time.Millisecond * 100)
	close(s.textChan)
}

func (s *Spin) reset() {
	if s.output == "" {
		return
	}
	fmt.Printf(ansi.CursorLeft + ansi.CursorUp(1) + ansi.EraseDown)
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
			frame := nextFrame()
			s.textChan <- fmt.Sprintf("%s...%c", s.DoingText, frame)
		}
	}()
	return s.textChan
}

func (s *Spin) renderToScreen() {
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

func (s *Spin) newFrameFactory() func() rune {
	index := 0
	size := len(s.frames)
	return func() rune {
		char := s.frames[index%size]
		index++
		return char
	}
}

var spinFrames = []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'}

//NewDotSpin get new DotSpinner instance
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
