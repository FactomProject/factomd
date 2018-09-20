package engine

import (
	"bufio"
	"flag"
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
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/elections"
	"github.com/FactomProject/factomd/p2p"
)

func init() {
	p := &Params // Global copy of decoded Params global.Params

	flag.StringVar(&p.DebugConsole, "debugconsole", "", "Enable DebugConsole on port. localhost:8093 open 8093 and spawns a telnet console, remotehost:8093 open 8093")
	flag.StringVar(&p.StdoutLog, "stdoutlog", "", "Log stdout to a file")
	flag.StringVar(&p.StderrLog, "stderrlog", "", "Log stderr to a file, optionally the same file as stdout")
	flag.StringVar(&p.DebugLogRegEx, "debuglog", "", "regex to pick which logs to save")
	flag.IntVar(&p.FaultTimeout, "faulttimeout", 120, "Seconds before considering Federated servers at-fault. Default is 120.")
	flag.IntVar(&p.RoundTimeout, "roundtimeout", 30, "Seconds before audit servers will increment rounds and volunteer.")
	flag.IntVar(&p2p.NumberPeersToBroadcast, "broadcastnum", 16, "Number of peers to broadcast to in the peer to peer networking")
	flag.StringVar(&p.ConfigPath, "config", "", "Override the config file location (factomd.conf)")
	flag.BoolVar(&p.CheckChainHeads, "checkheads", true, "Enables checking chain heads on boot")
	flag.BoolVar(&p.FixChainHeads, "fixheads", true, "If --checkheads is enabled, then this will also correct any errors reported")
	flag.BoolVar(&p.AckbalanceHash, "balancehash", true, "If false, then don't pass around balance hashes")
	flag.BoolVar(&p.EnableNet, "enablenet", true, "Enable or disable networking")
	flag.BoolVar(&p.WaitEntries, "waitentries", false, "Wait for Entries to be validated prior to execution of messages")
	flag.IntVar(&p.ListenTo, "node", 0, "Node Number the simulator will set as the focus")
	flag.IntVar(&p.Cnt, "count", 1, "The number of nodes to generate")
	flag.StringVar(&p.Net, "net", "tree", "The default algorithm to build the network connections")
	flag.StringVar(&p.Fnet, "fnet", "", "Read the given file to build the network connections")
	flag.IntVar(&p.DropRate, "drop", 0, "Number of messages to drop out of every thousand")
	flag.StringVar(&p.Journal, "journal", "", "Rerun a Journal of messages")
	flag.BoolVar(&p.Journaling, "journaling", false, "Write a journal of all messages received. Default is off.")
	flag.BoolVar(&p.Follower, "follower", false, "If true, force node to be a follower.  Only used when replaying a journal.")
	flag.BoolVar(&p.Leader, "leader", true, "If true, force node to be a leader.  Only used when replaying a journal.")
	flag.StringVar(&p.Db, "db", "", "Override the Database in the Config file and use this Database implementation. Options Map, LDB, or Bolt")
	flag.StringVar(&p.CloneDB, "clonedb", "", "Override the main node and use this database for the clones in a Network.")
	flag.StringVar(&p.NetworkName, "network", "", "Network to join: MAIN, TEST or LOCAL")
	flag.StringVar(&p.Peers, "peers", "", "Array of peer addresses. ")
	flag.IntVar(&p.BlkTime, "blktime", 0, "Seconds per block.  Production is 600.")
	flag.BoolVar(&p.RuntimeLog, "runtimeLog", false, "If true, maintain runtime logs of messages passed.")
	flag.BoolVar(&p.Exclusive, "exclusive", false, "If true, we only dial out to special/trusted peers.")
	flag.BoolVar(&p.ExclusiveIn, "exclusive_in", false, "If true, we only dial out to special/trusted peers and no incoming connections are accepted.")
	flag.StringVar(&p.Prefix, "prefix", "", "Prefix the Factom Node Names with this value; used to create leaderless networks.")
	flag.BoolVar(&p.Rotate, "rotate", false, "If true, responsibility is owned by one leader, and Rotated over the leaders.")
	flag.IntVar(&p.TimeOffset, "timedelta", 0, "Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.")
	flag.BoolVar(&p.KeepMismatch, "keepmismatch", false, "If true, do not discard DBStates even when a majority of DBSignatures have a different hash")
	flag.Int64Var(&p.StartDelay, "startdelay", 10, "Delay to start processing messages, in seconds")
	flag.IntVar(&p.Deadline, "deadline", 1000, "Timeout Delay in milliseconds used on Reads and Writes to the network comm")
	//flag.StringVar(&p.CustomNetName,"customnet", "", "This string specifies a custom blockchain network ID.")
	//p.CustomNet = primitives.Sha([]byte(*CustomNetPtr)).Bytes()[:4]
	flag.StringVar(&p.RpcUser, "rpcuser", "", "Username to protect factomd local API with simple HTTP authentication")
	flag.StringVar(&p.RpcPassword, "rpcpass", "", "Password to protect factomd local API. Ignored if rpcuser is blank")
	flag.BoolVar(&p.FactomdTLS, "tls", false, "Set to true to require encrypted connections to factomd API and Control Panel") //to get tls, run as "factomd -tls=true"
	flag.StringVar(&p.FactomdLocations, "selfaddr", "", "comma separated IPAddresses and DNS names of this factomd to use when creating a cert file")
	flag.IntVar(&p.MemProfileRate, "mpr", 512*1024, "Set the Memory Profile Rate to update profiling per X bytes allocated. Default 512K, set to 1 to profile everything, 0 to disable.")
	flag.BoolVar(&p.ExposeProfiling, "exposeprofiler", false, "Setting this exposes the profiling port to outside localhost.")
	//flag.StringVar(&,"factomhome", "", "Set the Factom home directory. The .factom folder will be placed here if set, otherwise it will default to $HOME")
	flag.StringVar(&p.LogPort, "logPort", "6060", "Port for pprof logging")
	flag.IntVar(&p.PortOverride, "port", 0, "Port where we serve WSAPI;  default 8088")
	flag.IntVar(&p.ControlPanelPortOverride, "controlpanelport", 0, "Port for control panel webserver;  Default 8090")
	flag.IntVar(&p.NetworkPortOverride, "networkport", 0, "Port for p2p network; default 8110")
	flag.BoolVar(&p.Fast, "fast", true, "If true, Factomd will fast-boot from a file.")
	flag.StringVar(&p.FastLocation, "fastlocation", "", "Directory to put the Fast-boot file in.")
	flag.StringVar(&p.Loglvl, "loglvl", "none", "Set log level to either: none, debug, info, warning, error, fatal or panic")
	flag.BoolVar(&p.Logjson, "logjson", false, "Use to set logging to use a json formatting")
	flag.BoolVar(&p.Sim_Stdin, "sim_stdin", true, "If true, sim control reads from stdin.")
	// Plugins
	flag.StringVar(&p.PluginPath, "plugin", "", "Input the path to any plugin binaries")
	// 	Torrent Plugin
	flag.BoolVar(&p.TorManage, "tormanage", false, "Use torrent dbstate manager. Must have plugin binary installed and in $PATH")
	flag.BoolVar(&p.TorUpload, "torupload", false, "Be a torrent uploader")
	// Logstash connection (if used)
	flag.BoolVar(&p.UseLogstash, "logstash", false, "If true, use Logstash")
	flag.StringVar(&p.LogstashURL, "logurl", "localhost:8345", "Endpoint URL for Logstash")
	flag.IntVar(&p.Sync2, "sync2", -1, "Set the initial blockheight for the second Sync pass. Used to force a total sync, or skip unnecessary syncing of entries.")

	flag.StringVar(&p.CustomNetName, "customnet", "", "This string specifies a custom blockchain network ID.")
	flag.StringVar(&p.FactomHome, "factomhome", "", "Set the Factom home directory. The .factom folder will be placed here if set, otherwise it will default to $HOME")
	flag.StringVar(&p.ControlPanelSetting, "controlpanelsetting", "", "Can set to 'disabled', 'readonly', or 'readwrite' to overwrite config file")

}

func ParseCmdLine(args []string) *FactomParams {
	p := &Params // Global copy of decoded Params global.Params
	flag.CommandLine.Parse(args)

	// Handle the global (not Factom server specific parameters
	if p.StdoutLog != "" || p.StderrLog != "" {
		handleLogfiles(p.StdoutLog, p.StderrLog)
	}

	fmt.Print("//////////////////////// Copyright 2017 Factom Foundation\n")
	fmt.Print("//////////////////////// Use of this source code is governed by the MIT\n")
	fmt.Print("//////////////////////// license that can be found in the LICENSE file.\n")

	elections.FaultTimeout = p.FaultTimeout
	elections.RoundTimeout = p.RoundTimeout

	p.CustomNet = primitives.Sha([]byte(p.CustomNetName)).Bytes()[:4]

	s, set := os.LookupEnv("FACTOM_HOME")
	if p.FactomHome != "" {
		os.Setenv("FACTOM_HOME", p.FactomHome)
		if set {
			fmt.Fprint(os.Stderr, "Overriding environment variable %s to be \"%s\"\n", "FACTOM_HOME", p.FactomHome)
		}
	} else {
		if set {
			p.FactomHome = s
		}

	}

	if !isCompilerVersionOK() {
		fmt.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		time.Sleep(3 * time.Second)
		os.Exit(1)
	}

	// launch debug console if requested
	if p.DebugConsole != "" {
		launchDebugServer(p.DebugConsole)
	}

	return p
}

func isCompilerVersionOK() bool {
	goodenough := false

	if strings.Contains(runtime.Version(), "1.7") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.8") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.9") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.10") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.11") {
		goodenough = true
	}

	return goodenough
}

var handleLogfilesOnce sync.Once

func handleLogfiles(stdoutlog string, stderrlog string) {

	handleLogfilesOnce.Do(func() {

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
	})
}

var launchDebugServerOnce sync.Once

func launchDebugServer(service string) {

	launchDebugServerOnce.Do(func() {

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

		//		wait.Add(1)
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
	})
}
