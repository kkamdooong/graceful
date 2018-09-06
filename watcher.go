package graceful

import (
	"context"
	"os"
	"os/signal"
)

type Watcher interface {
	Watch() chan bool
}

type SignalWatcher struct {
	signals []os.Signal
}

func (s *SignalWatcher) Watch() chan bool {
	signal.Reset(s.signals...)

	ch := make(chan os.Signal)
	signal.Notify(ch, s.signals...)

	watchChan := make(chan bool)
	go func() {
		select {
		case <-ch:
			watchChan <- true
		}
	}()

	return watchChan
}

func NewSignalWatcher(signals ...os.Signal) *SignalWatcher {
	return &SignalWatcher{
		signals: signals,
	}
}

type ChannelWatcher struct {
	ch chan interface{}
}

func (c *ChannelWatcher) Watch() chan bool {
	watchChan := make(chan bool)
	go func() {
		select {
		case <-c.ch:
			watchChan <- true
		}
	}()

	return watchChan
}

func NewChannelWatcher(ch chan interface{}) *ChannelWatcher {
	return &ChannelWatcher{
		ch: ch,
	}
}

type ContextWatcher struct {
	ctx context.Context
}

func (c *ContextWatcher) Watch() chan bool {
	watchChan := make(chan bool)
	go func() {
		select {
		case <-c.ctx.Done():
			watchChan <- true
		}
	}()

	return watchChan
}

func NewContextWatcher(ctx context.Context) *ContextWatcher {
	return &ContextWatcher{
		ctx: ctx,
	}
}
