// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/modules/registry"
	"github.com/FactomProject/factomd/modules/worker"
	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

// packageLogger is the general logger for all engine related logs. You can add additional fields,
// or create more context loggers off of this
var packageLogger = log.WithFields(log.Fields{"package": "engine"})

// start the process
func Run(params *FactomParams) {
	p := registry.New()
	p.Register(func(w *worker.Thread) {
		Factomd(w, params, params.Sim_Stdin)
		w.OnRun(func() { fmt.Print("Running") })
		w.OnComplete(func() {
			fmt.Println("Waiting to Shut Down")
			time.Sleep(time.Second * 5)
		})
	})
	p.Run()
}

func Factomd(w *worker.Thread, params *FactomParams, listenToStdin bool) {
	fmt.Printf("SpawnRun compiler version: %s\n", runtime.Version())
	fmt.Printf("Using build: %s\n", Build)
	fmt.Printf("Version: %s\n", FactomdVersion)
	StartTime = time.Now()
	fmt.Printf("Start time: %s\n", StartTime.String())
	go StartProfiler(params.MemProfileRate, params.ExposeProfiling)
	NetStart(w, params, listenToStdin)
}

func HandleLogfiles(stdoutlog string, stderrlog string) {
	var outfile *os.File
	var err error
	var wait sync.WaitGroup

	if stdoutlog != "" {
		// start a go routine to tee stdout to out.txt
		outfile, err = os.Create(stdoutlog)
		if err != nil {
			panic(err)
		}

		wait.Add(1)
		go func(outfile *os.File) {
			defer outfile.Close()
			defer os.Stdout.Close()                  // since I'm taking this away from  OS I need to close it when the time comes
			defer time.Sleep(100 * time.Millisecond) // Let the output all complete
			r, w, _ := os.Pipe()                     // Can't use the writer directly as os.Stdout so make a pipe
			oldStdout := os.Stdout
			os.Stdout = w
			wait.Done()
			// tee stdout to out.txt
			if _, err := io.Copy(io.MultiWriter(outfile, oldStdout), r); err != nil { // copy till EOF
				panic(err)
			}
		}(outfile) // stdout redirect func
	}

	if stderrlog != "" {
		if stderrlog != stdoutlog {
			outfile, err = os.Create(stderrlog)
			if err != nil {
				panic(err)
			}
		}

		wait.Add(1)
		go func(outfile *os.File) {
			if stderrlog != stdoutlog {
				defer outfile.Close()
			}
			defer os.Stderr.Close()                  // since I'm taking this away from  OS I need to close it when the time comes
			defer time.Sleep(100 * time.Millisecond) // Let the output all complete

			r, w, _ := os.Pipe() // Can't use the writer directly as os.Stdout so make a pipe
			oldStderr := os.Stderr
			os.Stderr = w
			wait.Done()
			if _, err := io.Copy(io.MultiWriter(outfile, oldStderr), r); err != nil { // copy till EOF
				panic(err)
			}
		}(outfile) // stderr redirect func
	}

	wait.Wait()                           // wait for the redirects to be active
	os.Stdout.WriteString("STDOUT Log\n") // Write any file header you want here e.g. node name and date and ...
	os.Stderr.WriteString("STDERR Log\n") // Write any file header you want here e.g. node name and date and ...
}

func LaunchDebugServer(service string) {

	// start a go routine to tee stderr to the debug console
	debugConsole_r, debugConsole_w, _ := os.Pipe() // Can't use the writer directly as os.Stdout so make a pipe
	var wait sync.WaitGroup
	wait.Add(1)
	go func() {

		r, w, _ := os.Pipe() // Can't use the writer directly as os.Stderr so make a pipe
		oldStderr := os.Stderr
		os.Stderr = w
		defer oldStderr.Close()                  // since I'm taking this away from  OS I need to close it when the time comes
		defer time.Sleep(100 * time.Millisecond) // let the output all complete
		wait.Done()
		if _, err := io.Copy(io.MultiWriter(oldStderr, debugConsole_w), r); err != nil { // copy till EOF
			panic(err)
		}
	}() // stderr redirect func

	//wait.Add(1)
	//go func() {
	//
	//	r, w, _ := os.Pipe() // Can't use the writer directly as os.Stderr so make a pipe
	//	oldStdout := os.Stdout
	//	os.Stdout = w
	//	defer oldStdout.Close()                  // since I'm taking this away from  OS I need to close it when the time comes
	//	defer time.Sleep(100 * time.Millisecond) // let the output all complete
	//	wait.Done()
	//	if _, err := io.Copy(io.MultiWriter(oldStdout, debugConsole_w), r); err != nil { // copy till EOF
	//		panic(err)
	//	}
	//}() // stdout redirect func

	wait.Wait() // Let the redirection become active ...

	host, port := "localhost", "8093" // defaults
	if service != "" {
		parts := strings.Split(service, ":")
		if len(parts) == 1 { // No port
			parts = append(parts, port) // use default
		}
		if parts[0] == "" { //no
			parts[0] = host // use default
		}
		host, port = parts[0], parts[1]

		_, badPort := strconv.Atoi(port)
		if (host != "localhost" && host != "remotehost") || badPort != nil {
			panic("Malformed -debugconsole option. Should be localhost:[port] or remotehost:[port] where [port] is a port number")
		}
	}

	// Start a listener port to connect to the debug server
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Debug Server listening for %v on port %v\n", host, port)

	newStdInR, newStdInW, _ := os.Pipe() // Can't use the reader directly as os.Stdin so make a pipe

	// Accept connections (one at a time)
	go func() {
		for {
			fmt.Printf("Debug server waiting for connection.\n") // Does not accept a reconnect not sure why ... revist
			connection, err := ln.Accept()
			if err != nil {
				panic(err)
			}
			fmt.Printf("Debug server accepted a connection.\n")

			writer := bufio.NewWriter(connection) // if we want to send something back to the telnet
			reader := bufio.NewReader(connection)

			writer.WriteString("Hello from Factom Debug Console\n")
			writer.Flush()
			// copy stderr to debug console
			go func() {
				if _, err := io.Copy(writer, debugConsole_r); err != nil { // copy till EOF
					fmt.Printf("Error copying stderr to debug consol: %v\n", err)
				}
			}()

			// copy input from debug console to stdin
			if false { // not sure why this doesn't work -- revist down the road
				if _, err = io.Copy(newStdInW, reader); err != nil {
					panic(err)
				}
			} else {
				for { // copy input from debug console to stdin until eof
					writer.WriteString(">") // print a prompt
					writer.Flush()
					if buf, err := reader.ReadString('\n'); err != nil {
						if err == io.EOF {
							break
						} // This connection is closed
						if err != nil {
							panic(err)
						} // This listen has an error
					} else {
						newStdInW.WriteString(string(buf))
					}
				}
			}
			fmt.Printf("Client disconnected.\n")
		}
	}() // the accept routine

	if host == "localhost" {
		cmd := exec.Command("/usr/bin/gnome-terminal", "-x", "telnet", "localhost", port)
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Debug terminal pid %v\n", cmd.Process.Pid)
	}
	os.Stdin = newStdInR               // start using the pipe as input
	time.Sleep(100 * time.Millisecond) // Let the redirection become active ...

}
