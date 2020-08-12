// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package worker

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// interruptChannel is used to receive SIGINT (Ctrl+C) signals.
var interruptChannel chan os.Signal
var once sync.Once

// addHandlerChannel is used to add an interrupt handler to the list of handlers
// to be invoked on SIGINT (Ctrl+C) signals.
var addHandlerChannel = make(chan func())

// mainInterruptHandler listens for SIGINT (Ctrl+C) signals on the
// interruptChannel and invokes the registered interruptCallbacks accordingly.
// It also listens for callback registration.  It must be run as a goroutine.
func mainInterruptHandler() {
	// In Go the defer statements are executed in LIFO order, guaranteeing that
	// os.Exit will be called last, after all the `defer handler()` have been executed
	defer os.Exit(0)

	for {
		select {
		case <-interruptChannel:
			fmt.Println("Received SIGINT (Ctrl+C). Shutting down...")
			return
		case handler := <-addHandlerChannel:
			defer handler()
		}
	}
}

// Create the channel and start the main interrupt handler which invokes
// all other callbacks and exits if not already done.
func startHandler() {
	once.Do(func() {
		interruptChannel = make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)
		go mainInterruptHandler()
	})
}

// AddInterruptHandler adds a handler to call when a SIGINT (Ctrl+C) is
// received.
func AddInterruptHandler(handler func()) {
	startHandler()
	addHandlerChannel <- handler
}

// SendSigInt adds a SIGINT to the intterupt handler
func SendSigInt() {
	startHandler()
	interruptChannel <- syscall.SIGINT
}
