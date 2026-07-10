package cli

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/ui"
)

// Block is the platform alias for ui.Block, so product packages get the
// concurrent-progress block type without importing platform internals.
type Block = ui.Block

// Progress is a per-invocation concurrent-progress session bound to the ctx
// progress writer (stdout in table mode, stderr in json/yaml).
type Progress struct {
	out io.Writer
	doc *ui.Document
}

// NewProgress builds a Progress bound to the ctx progress writer. Non-TTY
// writers suppress animation (handled inside ui.NewDocument).
func (c *Context) NewProgress() *Progress {
	w := c.ProgressWriter()
	return &Progress{out: w, doc: ui.NewDocument(w)}
}

// Disable switches off per-block animation (the count>5 aggregate path).
func (p *Progress) Disable() { p.doc.Disable() }

// Animated reports whether the progress document renders live frames. It is
// false when the bound writer is not a TTY (pipe/file/json mode) or Disable()
// was called for the aggregate count>5 path. When false, block content is never
// shown, so callers must surface errors to stderr (ctx.Err()) themselves.
func (p *Progress) Animated() bool { return !p.doc.Disabled() }

// NewBlock appends a fresh block to the document and returns it.
func (p *Progress) NewBlock() *Block {
	b := ui.NewBlock()
	p.doc.Append(b)
	return b
}

// Refresh prints an aggregate counter line to the progress writer.
func (p *Progress) Refresh(text string) { ui.NewRefresh(p.out).Do(text) }

// Sspoll runs the concurrent poller into block, bound to the progress writer.
func (p *Progress) Sspoll(describe func(string, *request.CommonBase) (interface{}, error),
	resourceID, text string, targetStates []string, block *Block, common *request.CommonBase) {
	NewPoller(describe, p.out).Sspoll(resourceID, text, targetStates, block, common)
}

// ConcurrentAction runs actionFunc over reqs with bounded concurrency (limit),
// aggregating a refresh counter when count>5. It is a verbatim port of
// cmd/util.go concurrentAction, rebound to the ctx progress writer. Products
// call this instead of touching platform internals.
func (c *Context) ConcurrentAction(reqs []request.Common, limit int, actionFunc func(request.Common) (bool, []string)) {
	if limit <= 0 {
		limit = 10
	}
	w := c.ProgressWriter()
	refresh := ui.NewRefresh(w)
	count := len(reqs)
	var wg sync.WaitGroup
	result := make(chan bool)
	tokens := make(chan bool, limit) // 控制并发量，最多 limit 个并发
	success, fail := 0, 0

	// 同时执行任务数量大于 5 时，不再单独显示每个任务，而是聚合显示。
	if count > 5 {
		refresh.Do(fmt.Sprintf("total:%d, doing:%d, success:%d, fail:%d", count, len(tokens), success, fail))
	}
	go func() {
		for {
			select {
			case ret := <-result:
				if ret {
					success++
				} else {
					fail++
				}
			case <-time.Tick(time.Second / 30):
				if count == (success+fail) && fail > 0 {
					fmt.Fprintf(w, "Check logs in %s\n", c.LogFilePath())
					return
				}
				if count > 5 {
					refresh.Do(fmt.Sprintf("total:%d, doing:%d, success:%d, fail:%d", count, len(tokens), success, fail))
				}
			}
		}
	}()

	for _, req := range reqs {
		wg.Add(1)
		go func(req request.Common) {
			tokens <- true
			ok, logs := actionFunc(req)
			result <- ok
			logs = append([]string{"========================================"}, logs...)
			c.LogInfo(logs...)
			<-tokens
			time.Sleep(time.Second / 5)
			wg.Done()
		}(req)
	}
	wg.Wait()
}
