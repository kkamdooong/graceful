package graceful

import (
	"testing"
	"time"
)

func TestManager_Shutdown(t *testing.T) {
	timeout := 3 * time.Second
	m := NewManager(timeout)

	notifier, err := m.RegisterNotifier(make(chan chan bool))
	if err != nil {
		t.Fatal(err)
	}

	exitFuncInvoked := make(chan bool, 1)
	m.RegisterExitFunc(func() {
		exitFuncInvoked <- true
	})

	if err := m.Start(); err != nil {
		t.Fatal(err)
	}

	m.Shutdown()

	select {
	case finishedCh := <-notifier:
		finishedCh <- true
	case <-time.After(timeout):
		t.Fatal("shutdown was invoked but notifier dosn't received channel until timeout")
	}

	select {
	case <-exitFuncInvoked:
	case <-time.After(timeout):
		t.Fatal("canceler was invoked but exit function dosn't invoked until timeout")
	}
}
