package ux

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ucloud/ucloud-cli/ansi"
)

var width, rows, _ = terminalSize()

//Document 当前进程在打印的内容
type document struct {
	blocks          []*Block
	mux             sync.RWMutex
	framesPerSecond int
	once            sync.Once
	out             io.Writer
	ticker          *time.Ticker
	disable         bool
}

func (d *document) reset() {
	size := 0
	d.mux.RLock()
	for _, block := range d.blocks {
		size += block.printLineNum
	}
	d.mux.RUnlock()
	if size != 0 {
		fmt.Printf(ansi.CursorLeft + ansi.CursorPrevLine(size) + ansi.EraseDown)
	}
}

func (d *document) Disable() {
	d.disable = true
}

func (d *document) SetWriter(out io.Writer) {
	d.out = out
}

func (d *document) Content() []string {
	var lines []string
	for _, block := range d.blocks {
		for _, line := range <-block.getLines {
			lines = append(lines, line)
		}
	}
	return lines
}

func (d *document) Render() {
	if d.disable {
		return
	}
	d.once.Do(func() {
		go func() {
			for range d.ticker.C {
				d.reset()
				d.mux.RLock()
				for _, block := range d.blocks {
					block.printLineNum = 0
					for _, line := range <-block.getLines {
						fmt.Fprintln(d.out, line)
						if width != 0 {
							lineNum := len(line)/width + 1
							block.printLineNum += lineNum
						} else {
							block.printLineNum++
						}
					}
					fmt.Fprintf(d.out, "\n")
					block.printLineNum++
				}
				d.mux.RUnlock()
			}
		}()
	})
}

func (d *document) Append(b *Block) {
	d.Render()
	d.mux.Lock()
	defer d.mux.Unlock()
	d.blocks = append(d.blocks, b)
}

func newDocument(out io.Writer) *document {
	doc := &document{
		out:             out,
		framesPerSecond: 20,
	}
	doc.ticker = time.NewTicker(time.Second / time.Duration(doc.framesPerSecond))
	return doc
}

//Doc global document
var Doc = newDocument(os.Stdout)

//Block in document, including a spinner and some text
type Block struct {
	spinner      *Spin
	spinnerIndex int
	printLineNum int //已打印到屏幕上的行数
	lines        []string
	updateLine   chan updateBlockLine
	getLines     chan []string
}

//Update lines in Block
func (b *Block) Update(text string, index int) {
	b.updateLine <- updateBlockLine{text, index}
}

//Append text to Block
func (b *Block) Append(text string) {
	b.updateLine <- updateBlockLine{text, -1}
}

//SetSpin set spin for block
func (b *Block) SetSpin(s *Spin) error {
	if b.spinner != nil {
		return fmt.Errorf("block has spinner already")
	}
	b.spinner = s
	b.spinnerIndex = len(<-b.getLines)
	strsCh := b.spinner.renderToString()
	go func() {
		for text := range strsCh {
			if len(<-b.getLines) == 0 {
				b.Append(text)
			} else {
				b.Update(text, b.spinnerIndex)
			}
		}
	}()
	return nil
}

type updateBlockLine struct {
	line  string
	index int
}

//NewSpinBlock create a new Block with spinner
func NewSpinBlock(s *Spin) *Block {
	block := NewBlock()
	if s != nil {
		block.SetSpin(s)
	}
	return block
}

//NewBlock  create a new Block without spinner. block.Stable closed
func NewBlock() *Block {
	block := &Block{
		lines:      []string{},
		updateLine: make(chan updateBlockLine, 0),
		getLines:   make(chan []string, 0),
	}

	go func() {
		for {
			select {
			case updateLine := <-block.updateLine:
				index, line := updateLine.index, updateLine.line
				if index < 0 {
					block.lines = append(block.lines, line)
				} else {
					block.lines[index] = line
				}

			case block.getLines <- block.lines:
			}
		}
	}()

	return block
}
