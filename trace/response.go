package trace

import (
	"context"

	"github.com/chromedp/cdproto/fetch"
)

// message that is send to trace collector
type capturedResponseBodyEvent struct {
	body      []byte
	requestID string
}

// wraper over cromedp fetch.GetResponseBodyParams
type fetchResponseAction struct {
	action    *fetch.GetResponseBodyParams
	requestID string
	out       chan<- any
}

// Do provides cromedp Action interface
func (a *fetchResponseAction) Do(ctx context.Context) error {
	body, err := a.action.Do(ctx)
	if err != nil {
		return err
	}
	a.out <- &capturedResponseBodyEvent{body, string(a.requestID)}
	return nil
}
