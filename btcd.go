// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	//"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	//	"runtime"
	"runtime/pprof"

	"github.com/FactomProject/FactomCode/util"
)

var (
	cfg             *config
	shutdownChannel = make(chan struct{})
)

// winServiceMain is only invoked on Windows.  It detects when btcd is running
// as a service and reacts accordingly.
var winServiceMain func() (bool, error)

// btcdMain is the real main function for btcd.  It is necessary to work around
// the fact that deferred functions do not run when os.Exit() is called.  The
// optional serverChan parameter is mainly used by the service code to be
// notified with the server once it is setup so it can gracefully stop it when
// requested from the service control manager.
func btcdMain(serverChan chan<- *server) error {
	util.Trace("******************************")

	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg
	defer backendLog.Flush()

	// Show version at startup.
	btcdLog.Infof("Version %s", version())

	// Enable http profiling server if requested.
	if cfg.Profile != "" {
		go func() {
			listenAddr := net.JoinHostPort("", cfg.Profile)
			btcdLog.Infof("Profile server listening on %s", listenAddr)
			profileRedirect := http.RedirectHandler("/debug/pprof",
				http.StatusSeeOther)
			http.Handle("/", profileRedirect)
			btcdLog.Errorf("%v", http.ListenAndServe(listenAddr, nil))
		}()
	}

	// Write cpu profile if requested.
	if cfg.CPUProfile != "" {
		f, err := os.Create(cfg.CPUProfile)
		if err != nil {
			btcdLog.Errorf("Unable to create cpu profile: %v", err)
			return err
		}
		pprof.StartCPUProfile(f)
		defer f.Close()
		defer pprof.StopCPUProfile()
	}

	/*
		// Perform upgrades to btcd as new versions require it.
		if err := doUpgrades(); err != nil {
			btcdLog.Errorf("%v", err)
			return err
		}
	*/

	// Load the block database.
	db, err := loadBlockDB()
	if err != nil {
		btcdLog.Errorf("%v", err)
		return err
	}
	defer db.Close()

	/*
		if cfg.DropAddrIndex {
			btcdLog.Info("Deleting entire addrindex.")
			err := db.DeleteAddrIndex()
			if err != nil {
				btcdLog.Errorf("Unable to delete the addrindex: %v", err)
				return err
			}
			btcdLog.Info("Successfully deleted addrindex, exiting")
			return nil
		}
	*/

	// Ensure the database is sync'd and closed on Ctrl+C.
	addInterruptHandler(func() {
		btcdLog.Infof("Gracefully shutting down the database...")
		db.RollbackClose()
	})

	// Create server and start it.
	server, err := newServer(cfg.Listeners, db, activeNetParams.Params)
	if err != nil {
		// TODO(oga) this logging could do with some beautifying.
		btcdLog.Errorf("Unable to start server on %v: %v",
			cfg.Listeners, err)
		return err
	}
	addInterruptHandler(func() {
		btcdLog.Infof("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()
	if serverChan != nil {
		serverChan <- server
	}

	// Factom Additions BEGIN
	factomForkInit(server)

	// Factom Additions END

	// Monitor for graceful server shutdown and signal the main goroutine
	// when done.  This is done in a separate goroutine rather than waiting
	// directly so the main goroutine can be signaled for shutdown by either
	// a graceful shutdown or from the main interrupt handler.  This is
	// necessary since the main goroutine must be kept running long enough
	// for the interrupt handler goroutine to finish.
	go func() {
		server.WaitForShutdown()
		srvrLog.Infof("Server shutdown complete")
		shutdownChannel <- struct{}{}
	}()

	// Wait for shutdown signal from either a graceful server stop or from
	// the interrupt handler.
	<-shutdownChannel
	btcdLog.Info("Shutdown complete")
	return nil
}
