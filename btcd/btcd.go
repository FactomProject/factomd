// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"github.com/FactomProject/factomd/common/interfaces"
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
//func btcdMain(serverChan chan<- *server) error {
func btcdMain(serverChan chan<- *Server, state interfaces.IState) error {

	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg
	// tweak some config options
	cfg.DisableCheckpoints = true
	defer backendLog.Flush()

	// Show version at startup.
	btcdLog.Infof("Version %s", version())

	// Ensure the database is sync'd and closed on Ctrl+C.
	AddInterruptHandler(func() {
		btcdLog.Infof("Gracefully shutting down the database...")
		state.GetDB().(interfaces.IDatabase).Close()		//db.RollbackClose()
	})

	// Create server and start it.
	server, err := newServer(cfg.Listeners, activeNetParams.Params, state)
	if err != nil {
		// TODO(oga) this logging could do with some beautifying.
		btcdLog.Errorf("Unable to start server on %v: %v",
			cfg.Listeners, err)
		return err
	}
	AddInterruptHandler(func() {
		btcdLog.Infof("Gracefully shutting down the server...")
		server.Stop()
		server.WaitForShutdown()
	})
	server.Start()
	if serverChan != nil {
		serverChan <- server
	}

	// Factom Additions BEGIN
	//factomForkInit(server)
	// Factom Additions END

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
