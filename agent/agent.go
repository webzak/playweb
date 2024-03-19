package agent

import (
	"context"

	"github.com/chromedp/chromedp"
)

// Agent is a wrapper around chromedp allocator and its derivatives (tabs)
type Agent struct {
	allocator       context.Context
	allocatorCancel context.CancelFunc
	tabs            map[context.Context]context.CancelFunc
}

// Cancel cancels allocator and all the tabs opened
func (a *Agent) Cancel() {
	for _, f := range a.tabs {
		f()
	}
	a.allocatorCancel()
}

// NewTab returns a new tab created with internal allocator
func (a *Agent) NewTab(opts ...chromedp.ContextOption) context.Context {
	ctx, cancel := chromedp.NewContext(a.allocator, opts...)
	a.tabs[ctx] = cancel
	return ctx
}

// CloseTab closes the tab associated with the provided context
func (a *Agent) CloseTab(ctx context.Context) {
	cancel := a.tabs[ctx]
	delete(a.tabs, ctx)
	cancel()
}

// NewExec returns a new Agent using the provided context and exec allocator options
func NewExec(ctx context.Context, opts ...chromedp.ExecAllocatorOption) *Agent {
	a := &Agent{
		tabs: make(map[context.Context]context.CancelFunc),
	}
	iopts := chromedp.DefaultExecAllocatorOptions[:]
	iopts = append(iopts, opts...)

	a.allocator, a.allocatorCancel = chromedp.NewExecAllocator(ctx, iopts...)
	return a
}

// NewRemote returns a new Agent using the provided context and remote allocator options
func NewRemote(ctx context.Context, url string, opts ...chromedp.RemoteAllocatorOption) *Agent {
	a := &Agent{
		tabs: make(map[context.Context]context.CancelFunc),
	}
	a.allocator, a.allocatorCancel = chromedp.NewRemoteAllocator(ctx, url, opts...)
	return a
}
