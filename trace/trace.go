package trace

import (
	"context"
	"log"
	"sync"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	cr "github.com/chromedp/chromedp"
)

type Capture struct {
	RequestID    string
	DocumentURL  string
	Request      *network.Request
	Response     *network.Response
	ResponseBody []byte
}

// Trace provides browser network tracing via cdp
type Trace struct {
	cdp context.Context
	in  chan any
	out chan *Capture
}

// Stop stops tracing
func (t *Trace) Stop() {
	cr.Run(t.cdp, fetch.Disable())
	close(t.in)
	close(t.out)
}

// Out returns the output channel
func (t *Trace) Out() <-chan *Capture {
	return t.out
}

// New returns a new Trace
func Start(ctx context.Context, wg *sync.WaitGroup, catchResponse bool, urlPattern ...string) *Trace {
	t := Trace{
		cdp: ctx,
		in:  make(chan any, 100),
		out: make(chan *Capture, 100),
	}
	if catchResponse {
		p := []*fetch.RequestPattern{}
		for _, up := range urlPattern {
			p = append(p, &fetch.RequestPattern{RequestStage: fetch.RequestStageResponse, URLPattern: up})
		}
		if len(p) == 0 {
			p = append(p, &fetch.RequestPattern{RequestStage: fetch.RequestStageResponse})
		}
		cr.Run(ctx, fetch.Enable().WithPatterns(p))
	}

	// listen target for particular messages
	cr.ListenTarget(ctx, func(ev any) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			go func(ev *network.EventRequestWillBeSent) {
				t.in <- ev
			}(ev)
		case *network.EventResponseReceived:
			go func(ev *network.EventResponseReceived) {
				t.in <- ev
			}(ev)
		case *fetch.EventRequestPaused:
			go func(ev *fetch.EventRequestPaused) {
				err := cr.Run(
					ctx,
					&fetchResponseAction{
						action:    fetch.GetResponseBody(ev.RequestID),
						requestID: string(ev.NetworkID),
						out:       t.in,
					},
					fetch.ContinueRequest(ev.RequestID),
				)
				if err != nil {
					log.Println(err)
				}
			}(ev)
		}
	})

	// collect traces together by request ID
	wg.Add(1)
	go func() {
		defer wg.Done()

		m := make(map[string]*Capture)

		for ev := range t.in {
			switch ev := ev.(type) {
			case *network.EventRequestWillBeSent:
				rid := string(ev.RequestID)
				c := m[rid]
				if c == nil {
					c = &Capture{}
					m[rid] = c
				}
				c.DocumentURL = string(ev.Request.URL)
				c.RequestID = rid
				c.Request = ev.Request

			case *capturedResponseBodyEvent:
				rid := ev.requestID
				c := m[rid]
				if c == nil {
					c = &Capture{}
					m[rid] = c
				}
				c.ResponseBody = ev.body
			case *network.EventResponseReceived:
				rid := string(ev.RequestID)

				c := m[rid]
				if c == nil {
					c = &Capture{}
					m[rid] = c
				}
				c.Response = ev.Response
				t.out <- c
				delete(m, rid)
			}
		}
	}()
	return &t
}
