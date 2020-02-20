// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FactomProject/btcutil/certs"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var Servers map[string]*Server
var ServersMutex sync.Mutex

func Start(state interfaces.IState) {
	RegisterPrometheus()

	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	if Servers == nil {
		Servers = make(map[string]*Server)
	}

	port := strconv.Itoa(state.GetPort())
	if Servers[port] == nil {
		server := InitServer(state)
		server.AddRootEndpoints()
		server.AddV1Endpoints()
		server.AddV2Endpoints()

		Servers[port] = server

		rpcUser := state.GetRpcUser()
		rpcPass := state.GetRpcPass()
		h := sha256.New()
		h.Write(httpBasicAuth(rpcUser, rpcPass))
		// TODO verify if there already runs a Server on the port, this code isn't executed, would change behavior.
		state.SetRpcAuthHash(h.Sum(nil)) //set this in the beginning to prevent timing attacks

		server.Start()
	}
}

func Stop(state interfaces.IState) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()

	port := strconv.Itoa(state.GetPort())
	Servers[port].Stop()
}

func SetState(state interfaces.IState) {
	wait := func() {
		port := strconv.Itoa(state.GetPort())
		ServersMutex.Lock()
		defer ServersMutex.Unlock()
		//todo: Should wait() instead of sleep but that requires plumbing a wait group....
		for Servers == nil && Servers[port] != nil && Servers[port].State != nil {
			ServersMutex.Unlock()
			time.Sleep(10 * time.Millisecond)
			ServersMutex.Lock()
		}
		if Servers == nil {
			wsLog.Error("Got here early need synchronization")
		}
		if Servers[port] == nil {
			wsLog.Error("Got here early need synchronization")
		}
		if Servers[port].State == nil {
			wsLog.Error("Got here early need synchronization")
		}

		Servers[port].State = state
	}
	go wait()
}

func GetState(r *http.Request) (state interfaces.IState, err error) {
	ServersMutex.Lock()
	defer ServersMutex.Unlock()
	port := r.Header.Get("factomd-port")
	if server, ok := Servers[port]; ok {
		return server.State, nil
	} else {
		return nil, errors.New(fmt.Sprintf("failed to get state, server initialization on port: %s failed", port))
	}
}

/*********************************************************
 * Support Functions
 *********************************************************/
func returnMsg(writer http.ResponseWriter, msg interface{}, success bool) {
	r := msg

	if p, err := json.Marshal(r); err != nil {
		wsLog.Error(err)
		return
	} else {
		_, err := writer.Write(p)
		if err != nil {
			wsLog.Errorf("failed to write response: %v", err)
		}
	}
}

func returnV1Msg(writer http.ResponseWriter, msg string, success bool) {
	/* V1 requires call specific case changes that can't be handled with
	interfaces for example.  Block Height needs to return  height as the json item name
	in golang, lower case names are private so won't be returned.
	Deal with the responses in the call specific v1 handlers until they are depricated.
	*/
	bMsg := []byte(msg)
	_, err := writer.Write(bMsg)
	if err != nil {
		wsLog.Errorf("failed to write v1 repsonse: %v", err)
	}
}

func handleV1Error(writer http.ResponseWriter, err *primitives.JSONError) {
	writer.WriteHeader(http.StatusBadRequest)
	return
}

func returnV1(writer http.ResponseWriter, jsonResp *primitives.JSON2Response, jsonError *primitives.JSONError) {
	if jsonError != nil {
		handleV1Error(writer, jsonError)
		return
	}
	returnMsg(writer, jsonResp.Result, true)
}

// httpBasicAuth returns the UTF-8 bytes of the HTTP Basic authentication
// string:
//
//   "Basic " + base64(username + ":" + password)
func httpBasicAuth(username, password string) []byte {
	const header = "Basic "
	base64 := base64.StdEncoding

	b64InputLen := len(username) + len(":") + len(password)
	b64Input := make([]byte, 0, b64InputLen)
	b64Input = append(b64Input, username...)
	b64Input = append(b64Input, ':')
	b64Input = append(b64Input, password...)

	output := make([]byte, len(header)+base64.EncodedLen(b64InputLen))
	copy(output, header)
	base64.Encode(output[len(header):], b64Input)
	return output
}

func checkAuthHeader(state interfaces.IState, request *http.Request) error {
	if "" == state.GetRpcUser() {
		//no username was specified in the config file or command line, meaning factomd API is open access
		return nil
	}

	authhdr := request.Header["Authorization"]
	if len(authhdr) == 0 {
		return errors.New("no auth")
	}

	correctAuth := state.GetRpcAuthHash()

	h := sha256.New()
	h.Write([]byte(authhdr[0]))
	presentedPassHash := h.Sum(nil)

	cmp := subtle.ConstantTimeCompare(presentedPassHash, correctAuth) //compare hashes because ConstantTimeCompare takes a constant time based on the slice size.  hashing gives a constant slice size.
	if cmp != 1 {
		return errors.New("bad auth")
	}
	return nil
}

func handleUnauthorized(request *http.Request, writer http.ResponseWriter) {
	remoteIP := ""
	remoteIP += strings.Split(request.RemoteAddr, ":")[0]
	wsLog.Debugf("Unauthorized V2 API client connection attempt from %s\n", remoteIP)
	writer.Header().Add("WWW-Authenticate", `Basic realm="factomd RPC"`)
	http.Error(writer, "401 Unauthorized.", http.StatusUnauthorized)
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func genCertPair(certFile string, keyFile string, extraAddress string) error {
	wsLog.Info("Generating TLS certificates...")

	org := "factom autogenerated cert"
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)

	var externalAddresses []string
	if extraAddress != "" {
		externalAddresses = strings.Split(extraAddress, ",")
		for _, i := range externalAddresses {
			wsLog.Infof("adding %s to certificate\n", i)
		}
	}

	cert, key, err := certs.NewTLSCertPair(org, validUntil, externalAddresses)
	if err != nil {
		return err
	}

	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}

	wsLog.Info("Done generating TLS certificates")
	return nil
}
