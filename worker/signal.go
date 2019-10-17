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
	"syscall"
)

// interruptChannel is used to receive SIGINT (Ctrl+C) signals.
var interruptChannel chan os.Signal

// addHandlerChannel is used to add an interrupt handler to the list of handlers
// to be invoked on SIGINT (Ctrl+C) signals.
var addHandlerChannel = make(chan func())

// mainInterruptHandler listens for SIGINT (Ctrl+C) signals on the
// interruptChannel and invokes the registered interruptCallbacks accordingly.
// It also listens for callback registration.  It must be run as a goroutine.
func mainInterruptHandler() {
	// interruptCallbacks is a list of callbacks to invoke when a
	// SIGINT (Ctrl+C) is received.
	var interruptCallbacks []func()

	// isShutdown is a flag which is used to indicate whether or not
	// the shutdown signal has already been received and hence any future
	// attempts to add a new interrupt handler should invoke them
	// immediately.
	var isShutdown bool

	for {
		select {
		case <-interruptChannel:
			// Ignore more than one shutdown signal.
			if isShutdown {
				fmt.Println("Ctrl+C Already being processed!")
				continue
			}
			isShutdown = true
			fmt.Println("Received SIGINT (Ctrl+C).  Shutting down...")

			// Run handlers in LIFO order.
			for i := range interruptCallbacks {
				idx := len(interruptCallbacks) - 1 - i
				callback := interruptCallbacks[idx]
				callback()
			}

		case handler := <-addHandlerChannel:
			// The shutdown signal has already been received, so
			// just invoke and new handlers immediately.
			if isShutdown {
				handler()
			}

			interruptCallbacks = append(interruptCallbacks, handler)
		}
	}
}

// AddInterruptHandler adds a handler to call when a SIGINT (Ctrl+C) is
// received.
func AddInterruptHandler(handler func()) {
	// Create the channel and start the main interrupt handler which invokes
	// all other callbacks and exits if not already done.
	if interruptChannel == nil {
		interruptChannel = make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)
		go mainInterruptHandler()
	}

	addHandlerChannel <- handler
}

func SendSigInt() {
	interruptChannel <- syscall.SIGINT
}
