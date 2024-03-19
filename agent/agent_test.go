package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExec(t *testing.T) {
	a := NewExec(context.Background())
	assert.NotNil(t, a.allocator)
	assert.NotNil(t, a.allocatorCancel)
	a.Cancel()
	assert.Error(t, a.allocator.Err())
}

func TestNewRemote(t *testing.T) {
	a := NewRemote(context.Background(), "wss://localhost:9222/")
	assert.NotNil(t, a.allocator)
	assert.NotNil(t, a.allocatorCancel)
	a.Cancel()
	assert.Error(t, a.allocator.Err())
}

func TestNewTab(t *testing.T) {
	a := NewExec(context.Background())
	defer a.Cancel()
	assert.NotNil(t, a.allocator)
	assert.NotNil(t, a.allocatorCancel)
	tab := a.NewTab()
	assert.Nil(t, tab.Err())
	a.CloseTab(tab)
	assert.Error(t, tab.Err())
}

func TestCancelAllTabs(t *testing.T) {
	a := NewExec(context.Background())
	tab1 := a.NewTab()
	assert.Nil(t, tab1.Err())
	tab2 := a.NewTab()
	assert.Nil(t, tab2.Err())
	a.Cancel()
	assert.Error(t, tab1.Err())
	assert.Error(t, tab2.Err())
}
