package pdfjob

import (
	"context"
	"os"
	"sync"

	"github.com/chromedp/chromedp"
)

// BrowserPool holds a persistent Chrome exec allocator so every PDF reuses
// the same process instead of spawning a new one per request.
type BrowserPool struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
}

var (
	globalPool     *BrowserPool
	globalPoolOnce sync.Once
)

// GetBrowserPool returns the process-wide singleton. Safe for concurrent use.
func GetBrowserPool() *BrowserPool {
	globalPoolOnce.Do(func() {
		globalPool = newBrowserPool()
	})
	return globalPool
}

func newBrowserPool() *BrowserPool {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.DisableGPU,
	)
	if p := os.Getenv("CHROME_PATH"); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	return &BrowserPool{allocCtx: allocCtx, allocCancel: allocCancel}
}

// NewTab opens a new Chrome tab in the existing process.
// Caller must defer the returned cancel to close the tab.
func (p *BrowserPool) NewTab() (context.Context, context.CancelFunc) {
	return chromedp.NewContext(p.allocCtx)
}

// Close shuts down the Chrome process. Call on application exit.
func (p *BrowserPool) Close() {
	p.allocCancel()
}
