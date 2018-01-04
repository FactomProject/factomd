// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"

	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"bufio"
	"os/exec"
)

var _ = fmt.Print

// winServiceMain is only invoked on Windows.  It detects when btcd is running
// as a service and reacts accordingly.
//var winServiceMain func() (bool, error)

// packageLogger is the general logger for all engine related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "engine"})

// Build sets the factomd build id using git's SHA
// Version sets the semantic version number of the build
// $ go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.=`cat VERSION`"
// It also seems to need to have the previous binary deleted if recompiling to have this message show up if no code has changed.
// Since we are tracking code changes, then there is no need to delete the binary to use the latest message
var Build string
var FactomdVersion string = "BuiltWithoutVersion"

func Factomd(params *FactomParams, listenToStdin bool) interfaces.IState {
	log.Print("//////////////////////// Copyright 2017 Factom Foundation")
	log.Print("//////////////////////// Use of this source code is governed by the MIT")
	log.Print("//////////////////////// license that can be found in the LICENSE file.")
	log.Printf("Go compiler version: %s\n", runtime.Version())
	log.Printf("Using build: %s\n", Build)
	log.Printf("Version: %s\n", FactomdVersion)

	if !isCompilerVersionOK() {
		for i := 0; i < 30; i++ {
			log.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		}
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	// launch debug console if requested
	if (params.DebugConsole) {
		launchDebugServer()
	}

	//  Go Optimizations...
	runtime.GOMAXPROCS(runtime.NumCPU())

	state0 := new(state.State)
	state0.IsRunning = true
	state0.SetLeaderTimestamp(primitives.NewTimestampFromMilliseconds(0))
	fmt.Println("len(Args)", len(os.Args))

	go NetStart(state0, params, listenToStdin)
	return state0
}

func isCompilerVersionOK() bool {
	goodenough := false

	if strings.Contains(runtime.Version(), "1.5") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.6") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.7") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.8") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.9") {
		goodenough = true
	}
	return goodenough
}

func launchDebugServer() {
	// start a go routine to tee stdout to out.txt
	go func() {
		outfile, err := os.Create("./out.txt")
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
		defer os.Stdout.Close()                  // since I'm taking this away from  OS I need to close it when the time comes
		defer time.Sleep(100 * time.Millisecond) // Let the output all complete
		outfile.WriteString("test\n")
		r, w, _ := os.Pipe() // Can't use the writer directly as os.Stdout so make a pipe
		oldStdout := os.Stdout
		os.Stdout = w
		for { // for ever ....
			_, err := io.Copy(io.MultiWriter(outfile, oldStdout), r) // tee stdout to out.txt
			if err != nil {
				panic(err)
			}
		}
	}()

	// start a go routine to tee stderr to err.txt and the debug console
	stdErrPipe_r, stdErrPipe_w, _ := os.Pipe() // Can't use the writer directly as os.Stdout so make a pipe

	go func() {
		outfile, err := os.Create("./err.txt")
		if err != nil {
			panic(err)
		}
		defer outfile.Close()
		defer os.Stderr.Close()                  // since I'm taking this away from  OS I need to close it when the time comes
		defer time.Sleep(100 * time.Millisecond) // Let the output all complete
		outfile.WriteString("test error\n")

		r, w, _ := os.Pipe() // Can't use the writer directly as os.Stdout so make a pipe
		oldStderr := os.Stderr
		os.Stderr = w
		for { // for ever ....
			_, err := io.Copy(io.MultiWriter(outfile, oldStderr, stdErrPipe_w), r)
			if err != nil {
				panic(err)
			}

		}
	}()

	time.Sleep(100 * time.Millisecond) // Let the redirection become active ...

	// test tee
	os.Stdout.WriteString("This is stdout!\n")
	os.Stderr.WriteString("This is stderr!\n")

	// Start a listener port to connect to the debug server
	ln, err := net.Listen("tcp", ":8091")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Debug Server is ready. ")

	newStdIn_r, newStdIn_w, _ := os.Pipe() // Can't use the reader directly as os.Stdin so make a pipe

	// Accept
	go func() {
		connection, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Debug server accepted a connection.")

		writer := bufio.NewWriter(connection) // if we want to send something back to the telnet
		reader := bufio.NewReader(connection)

		writer.WriteString("Hello User\n")
		writer.Flush()
		for {
			// copy stderr to debug console
			go func() {
				_, err := io.Copy(writer, stdErrPipe_r)
				if err != nil {
					fmt.Printf("Error copying stderr to debug consol: %v\n", err)
				}
			}()

			// copy input from debug console to stdin
			if false {
				_, err = io.Copy(newStdIn_w, reader) // not sure why this doesn't work
				if  err != nil {
					break
				}
			} else {
				buf, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				newStdIn_w.WriteString(string(buf))
			}
		}
		fmt.Printf("Client disconnected., %v\n", err)
	}()

	cmd := exec.Command("/usr/bin/gnome-terminal", "-x", "telnet", "localhost", "8091")
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Debug terminal pid %v\n", cmd.Process.Pid)

	os.Stdin = newStdIn_r              // start using the pipe as input
	time.Sleep(100 * time.Millisecond) // Let the redirection become active ...

}
