package cli

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/ux"
)

// Block is the platform alias for ux.Block, so product packages get the
// concurrent-progress block type without ever importing ux.
type Block = ux.Block

// Progress is a per-invocation concurrent-progress session bound to the ctx
// progress writer (stdout in table mode, stderr in json/yaml). It wraps
// ux.Document/Block/Refresh so products never import ux or base.
type Progress struct {
	out io.Writer
	doc *ux.Document
}

// NewProgress builds a Progress bound to the ctx progress writer. Non-TTY
// writers suppress animation (handled inside ux.NewDocument).
func (c *Context) NewProgress() *Progress {
	w := c.ProgressWriter()
	return &Progress{out: w, doc: ux.NewDocument(w)}
}

// Disable switches off per-block animation (the count>5 aggregate path).
func (p *Progress) Disable() { p.doc.Disable() }

// NewBlock appends a fresh block to the document and returns it.
func (p *Progress) NewBlock() *Block {
	b := ux.NewBlock()
	p.doc.Append(b)
	return b
}

// Refresh prints an aggregate counter line to the progress writer.
func (p *Progress) Refresh(text string) { ux.NewRefreshTo(p.out).Do(text) }

// Sspoll runs the concurrent poller into block, bound to the progress writer.
func (p *Progress) Sspoll(describe func(string, *request.CommonBase) (interface{}, error),
	resourceID, text string, targetStates []string, block *Block, common *request.CommonBase) {
	base.NewSpoller(describe, p.out).Sspoll(resourceID, text, targetStates, block, common)
}

// ConcurrentAction runs actionFunc over reqs with bounded concurrency (limit),
// aggregating a refresh counter when count>5. It is a verbatim port of
// cmd/util.go concurrentAction, rebound to the ctx progress writer: the global
// ux.Doc / ux.NewRefresh become the per-ctx writer; base logging (which is a
// platform concern) stays in base. Products call this instead of touching base.
func (c *Context) ConcurrentAction(reqs []request.Common, limit int, actionFunc func(request.Common) (bool, []string)) {
	if limit <= 0 {
		limit = 10
	}
	w := c.ProgressWriter()
	refresh := ux.NewRefreshTo(w)
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
					fmt.Fprintf(w, "Check logs in %s\n", base.GetLogFilePath())
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
			base.LogInfo(logs...)
			<-tokens
			time.Sleep(time.Second / 5)
			wg.Done()
		}(req)
	}
	wg.Wait()
}
