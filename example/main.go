package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/kkamdooong/graceful"
)

func main() {
	manager := graceful.NewManager(3 * time.Second)
	manager.WatchSignal(syscall.SIGINT, syscall.SIGTERM)
	notifier, _ := manager.RegisterNotifier(make(chan chan bool))
	manager.Start()

	// Foo wants to be notified when program is finished for cleaning some resources
	go Foo(notifier)

	select {}

	// If you send a shutdown signal after running program,
	// program waits to shutdown until all cleaning process complete.
}

func Foo(notifier chan chan bool) {
	select {
	case finishedCh := <-notifier:
		// If program receive SIGINT or SIGTERM, notifier sends finished channel

		// some resource clean process here
		// ...
		fmt.Println("cleaning process is finished")

		// finally, notify to manager
		finishedCh <- true
	}
}
