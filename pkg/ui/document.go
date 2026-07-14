package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Document is a writer-bound live rendering surface for progress blocks.
type Document struct {
	blocks          []*Block
	mux             sync.RWMutex
	framesPerSecond int
	once            sync.Once
	out             io.Writer
	ticker          *time.Ticker
	disable         bool
}

func (d *Document) reset() {
	size := 0
	d.mux.RLock()
	for _, block := range d.blocks {
		size += block.printLineNum
	}
	d.mux.RUnlock()
	if size != 0 {
		fmt.Fprint(d.out, ansiCursorLeft+ansiCursorPrevLine(size)+ansiEraseDown)
	}
}

func (d *Document) Disable() {
	d.disable = true
}

func (d *Document) Disabled() bool {
	return d.disable
}

func (d *Document) SetWriter(out io.Writer) {
	d.out = out
}

func (d *Document) Content() []string {
	var lines []string
	for _, block := range d.blocks {
		for _, line := range <-block.getLines {
			lines = append(lines, line)
		}
	}
	return lines
}

func (d *Document) Render() {
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
						block.printLineNum++
					}
					fmt.Fprintf(d.out, "\n")
					block.printLineNum++
				}
				d.mux.RUnlock()
			}
		}()
	})
}

func (d *Document) Append(b *Block) {
	d.Render()
	d.mux.Lock()
	defer d.mux.Unlock()
	d.blocks = append(d.blocks, b)
}

func (d *Document) GetLastBlock() *Block {
	d.mux.Lock()
	defer d.mux.Unlock()
	if len(d.blocks) == 0 {
		return nil
	}
	return d.blocks[len(d.blocks)-1]
}

func (d *Document) GetBlockCount() int {
	d.mux.Lock()
	defer d.mux.Unlock()
	return len(d.blocks)
}

func NewDocument(out io.Writer) *Document {
	doc := &Document{
		out:             out,
		framesPerSecond: 20,
		disable:         !IsTTY(out),
	}
	doc.ticker = time.NewTicker(time.Second / time.Duration(doc.framesPerSecond))
	return doc
}

// Block is one progress document block, including an optional spinner and text.
type Block struct {
	spinner      *Spin
	spinnerIndex int
	printLineNum int
	lines        []string
	updateLine   chan updateBlockLine
	getLines     chan []string
}

func (b *Block) Update(text string, index int) {
	b.updateLine <- updateBlockLine{text, index}
}

func (b *Block) Append(text string) {
	b.updateLine <- updateBlockLine{text, -1}
}

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

func NewSpinBlock(s *Spin) *Block {
	block := NewBlock()
	if s != nil {
		block.SetSpin(s)
	}
	return block
}

func NewBlock() *Block {
	block := &Block{
		lines:      []string{},
		updateLine: make(chan updateBlockLine),
		getLines:   make(chan []string),
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
