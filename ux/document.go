package ux

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ucloud/ucloud-cli/ansi"
)

//Document 当前进程在打印的内容
type document struct {
	blocks          []*Block
	framesPerSecond int
	once            sync.Once
	out             io.Writer
	ticker          *time.Ticker
	mux             sync.Mutex
	ctx             context.Context
	cancel          context.CancelFunc
	allBlockFull    chan bool
	Done            chan bool
	disable         bool
}

var width, rows, _ = terminalSize()

func (d *document) reset() {
	size := 0
	for _, block := range d.blocks {
		size += block.printLineNum
	}
	if size != 0 {
		fmt.Printf(ansi.CursorLeft + ansi.CursorPrevLine(size) + ansi.EraseDown)
	}
}

func (d *document) Disable() {
	d.disable = true
}

func (d *document) Render() {
	if d.disable {
		return
	}
	d.once.Do(func() {
		go func() {
			for range d.ticker.C {
				d.reset()
				for _, block := range d.blocks {
					block.printLineNum = 0
					block.mux.Lock()
					for _, line := range block.lines {
						fmt.Fprintln(d.out, line)
						if width != 0 {
							lineNum := len(line)/width + 1
							block.printLineNum += lineNum
						} else {
							block.printLineNum++
						}
					}
					block.mux.Unlock()
					fmt.Fprintf(d.out, "\n")
					block.printLineNum++
				}
			}
		}()
		go d.checkBlockDone()
	})
}

func (d *document) Append(b *Block) {
	d.Render()
	d.mux.Lock()
	defer d.mux.Unlock()
	if d.cancel != nil {
		d.cancel()
	}
	d.ctx, d.cancel = context.WithCancel(context.Background())
	go d.checkBlockFull(d.ctx)
	d.blocks = append(d.blocks, b)
}

func (d *document) checkBlockFull(ctx context.Context) {
	allFull := make(chan struct{})
	go func() {
		for _, b := range d.blocks {
			<-b.full
		}
		close(allFull)
	}()

	select {
	case <-ctx.Done():
		return
	case <-allFull:
		d.allBlockFull <- true
		return
	}
}

func (d *document) checkBlockDone() {
	<-d.allBlockFull
	allStable := make(chan struct{})
	go func() {
		for _, b := range d.blocks {
			<-b.stable
		}
		close(allStable)
	}()
	<-allStable
	//等待最后一帧渲染
	<-time.After(time.Millisecond * 200)
	close(d.Done)
}

func newDocument(out io.Writer) *document {
	doc := &document{
		out:             out,
		framesPerSecond: 20,
		Done:            make(chan bool),
		allBlockFull:    make(chan bool),
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
	mux          sync.Mutex
	lines        []string
	stable       chan struct{} //标识此块已稳定，不再轮询
	full         chan struct{} //标识此块不再添加新的内容
}

//Update lines in Block
func (b *Block) Update(text string, index int) {
	b.mux.Lock()
	b.lines[index] = text
	b.mux.Unlock()
}

//Append text to Block
func (b *Block) Append(text string) {
	b.lines = append(b.lines, text)
}

//AppendDone 表示不再往Block内部添加内容
func (b *Block) AppendDone() {
	close(b.full)
}

//SetSpin set spin for block
func (b *Block) SetSpin(s *Spin) error {
	if b.spinner != nil {
		return fmt.Errorf("block has spinner already")
	}
	b.stable = make(chan struct{})
	b.spinner = s
	b.spinnerIndex = len(b.lines)
	b.lines = append(b.lines, "loading")
	strsCh := b.spinner.renderToString()
	go func() {
		for text := range strsCh {
			if len(b.lines) == 0 {
				b.Append(text)
			} else {
				b.Update(text, b.spinnerIndex)
			}
		}
		close(b.stable)
	}()
	return nil
}

//NewSpinBlock create a new Block with spinner
func NewSpinBlock(s *Spin) *Block {
	block := &Block{
		lines:  []string{},
		stable: make(chan struct{}),
		full:   make(chan struct{}),
	}
	if s != nil {
		block.SetSpin(s)
	} else {
		close(block.stable)
	}
	return block
}

//NewBlock  create a new Block without spinner. block.Stable closed
func NewBlock() *Block {
	block := &Block{
		lines:  []string{},
		stable: make(chan struct{}),
		full:   make(chan struct{}),
	}
	close(block.stable)
	return block
}
