package graceful

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var errAlreadyStarted = errors.New("shutdown manager already started")

type Manager struct {
	mutex   sync.Mutex
	timeout time.Duration
	started atomic.Value

	watchCh   chan bool
	watchers  []Watcher
	notifiers []chan chan bool

	exitFunc func()
}

func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		timeout:  timeout,
		started:  atomic.Value{},
		exitFunc: func() { os.Exit(1) },
	}
}

func (s *Manager) RegisterWatcher(w Watcher) error {
	if s.started.Load() != nil {
		return errAlreadyStarted
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.watchers = append(s.watchers, w)
	return nil
}

func (s *Manager) WatchSignal(signals ...os.Signal) error {
	sw := NewSignalWatcher(signals...)

	return s.RegisterWatcher(sw)
}

func (s *Manager) WatchContext(ctx context.Context) error {
	cw := NewContextWatcher(ctx)

	return s.RegisterWatcher(cw)
}

func (s *Manager) WatchChannel(ch chan interface{}) error {
	cw := NewChannelWatcher(ch)

	return s.RegisterWatcher(cw)
}

func (s *Manager) RegisterNotifier(notifier chan chan bool) (chan chan bool, error) {
	if s.started.Load() != nil {
		return nil, errAlreadyStarted
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.notifiers = append(s.notifiers, notifier)

	return notifier, nil
}

func (s *Manager) RegisterExitFunc(f func()) error {
	if s.started.Load() != nil {
		return errAlreadyStarted
	}

	s.exitFunc = f

	return nil
}

func (s *Manager) Start() error {
	if s.started.Load() != nil {
		return errAlreadyStarted
	}

	s.started.Store(true)
	go s.watch()
	return nil
}

func (s *Manager) watch() {
	watchCh := make(chan bool)
	for _, w := range s.watchers {
		go func() {
			watched := <-w.Watch()
			watchCh <- watched
		}()
	}

	select {
	case <-watchCh:
		// wait until any watcher send watched
	}

	s.Shutdown()
}

func (s *Manager) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	done := make(chan bool)
	go s.notify(done)

	select {
	case <-time.After(s.timeout):
	case <-done:
		// wait untill all notifier send finished
	}

	s.exitFunc()
}

func (s *Manager) notify(done chan bool) {
	wg := new(sync.WaitGroup)
	wg.Add(len(s.notifiers))
	for _, n := range s.notifiers {
		go func(ch chan chan bool) {
			finished := make(chan bool)
			ch <- finished

			select {
			case <-finished:
				wg.Done()
			}
		}(n)
	}

	wg.Wait()
	done <- true
}
