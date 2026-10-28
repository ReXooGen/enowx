package canvasview

import (
	"strings"
	"testing"
	"time"
)

type delayedConnectorPage struct {
	livePage
	started chan string
	release chan struct{}
}

func (p *delayedConnectorPage) Eval(code string) (any, error) {
	p.started <- code
	<-p.release
	return true, nil
}

func TestConnectorUpdatesCoalesceWhileBrowserIsBusy(t *testing.T) {
	p := &delayedConnectorPage{started: make(chan string, 2), release: make(chan struct{}, 2)}
	updates := make(chan func(), 2)
	w := &Web{page: p, Update: func(fn func()) { updates <- fn }}
	w.updateConnectors("first")
	select {
	case <-p.started:
	case <-time.After(time.Second):
		t.Fatal("first update never started")
	}
	w.updateConnectors("second")
	w.updateConnectors("latest")
	select {
	case code := <-p.started:
		t.Fatalf("concurrent browser update: %s", code)
	default:
	}
	p.release <- struct{}{}
	select {
	case fn := <-updates:
		fn()
	case <-time.After(time.Second):
		t.Fatal("first update never completed")
	}
	select {
	case code := <-p.started:
		if !strings.Contains(code, "latest") || strings.Contains(code, "second") {
			t.Fatalf("queued stale position: %s", code)
		}
	case <-time.After(time.Second):
		t.Fatal("latest update was lost")
	}
	p.release <- struct{}{}
	select {
	case fn := <-updates:
		fn()
	case <-time.After(time.Second):
		t.Fatal("latest update never completed")
	}
	if w.connectorBusy {
		t.Fatal("browser remains busy after its reply")
	}
}
