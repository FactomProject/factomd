package wsapi

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	State      interfaces.IState
	httpServer *http.Server
	router     *mux.Router
	tlsEnabled bool
	certFile   string
	keyFile    string
	Port       string
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

func InitServer(state interfaces.IState) *Server {
	tlsIsEnabled, keyFile, certFile := state.GetTlsInfo()
	address := fmt.Sprintf(":%d", state.GetPort())

	router := mux.NewRouter()
	port := strconv.Itoa(state.GetPort())
	server := Server{State: state, router: router, tlsEnabled: tlsIsEnabled, certFile: certFile, keyFile: keyFile, Port: port}

	if tlsIsEnabled {
		router.Schemes("HTTPS")
		wsLog.Info("Starting encrypted API server")
		if !fileExists(keyFile) && !fileExists(certFile) {
			err := genCertPair(certFile, keyFile, state.GetFactomdLocations())
			if err != nil {
				panic(fmt.Sprintf("could not start encrypted API server with error: %v", err))
			}
		}
		keypair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			panic(fmt.Sprintf("could not create TLS keypair with error: %v", err))
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keypair},
			MinVersion:   tls.VersionTLS12,
		}

		server.httpServer = &http.Server{Addr: address, Handler: router, TLSConfig: tlsConfig}
	} else {
		server.httpServer = &http.Server{Addr: address, Handler: router}
	}

	wsLog.Infof("Init API server at: %s\n", address)

	return &server
}

func (server *Server) Start() {
	wsLog.Info("Starting API server")
	go func() {
		// returns ErrServerClosed on graceful close
		if server.tlsEnabled {
			if err := server.httpServer.ListenAndServeTLS(server.certFile, server.keyFile); err != http.ErrServerClosed {
				wsLog.Errorf("ListenAndServeTLS %v", err)
			}
		} else {
			if err := server.httpServer.ListenAndServe(); err != http.ErrServerClosed {
				wsLog.Errorf("ListenAndServe %v", err)
			}
		}
	}()
}

func (server *Server) Stop() {
	// close the server gracefully ("shutdown")
	wsLog.Info("closing wsapi server")
	if err := server.httpServer.Shutdown(context.Background()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

// Logging logs all requests with its path and the time it took to process
func APILogger() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wsLog.Debugf("%s\t%s\t%s\n", r.Method, r.RequestURI, time.Since(start))

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// IDInjector injects the server's ID into every request as http header
func IDInjector(server *Server) Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("factomd-port", server.Port)
			f(w, r)
		}
	}
}

func PanicRecovery() Middleware {
	return func(f http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); r != nil {
					trace := debug.Stack()
					wsLog.Errorf("Recovered from a panic: %v: %s", rec, string(trace))
					HandleV2Error(w, nil, NewInternalError())
				}
			}()
			f(w, r)
		}
	}
}

// add route and Chain applies middlewares to a http.HandlerFunc
func (server *Server) addRoute(path string, f func(http.ResponseWriter, *http.Request), middlewares ...Middleware) *mux.Route {
	middlewares = append(middlewares, APILogger())
	middlewares = append(middlewares, IDInjector(server))
	middlewares = append(middlewares, PanicRecovery()) // keep this last
	for _, m := range middlewares {
		f = m(f)
	}
	return server.router.HandleFunc(path, f)
}

func (server *Server) AddRootEndpoints() {
	state := server.State
	if len(state.GetCorsDomains()) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins: state.GetCorsDomains(),
		})

		server.router.Use(c.Handler)
	}

	// for the v1 endpoints the default behavior of a not allowed method behaviour is different.
	// for v2 and debug endpoints this isn't applicable as all methods accept both gets, and posts
	server.router.MethodNotAllowedHandler = methodNotAllowedHandler()

	// start the debugging api if we are not on the main network
	if state.GetNetworkName() != "MAIN" {
		server.addRoute("/debug", HandleDebug).Methods("GET", "POST")
	}
}

// methodNotAllowed replies to the request with an HTTP status code 404 instead of default 405.
func methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

// methodNotAllowedHandler returns a simple request handler
// that replies to each request with a status code 404 instead of the default 405.
func methodNotAllowedHandler() http.Handler { return http.HandlerFunc(methodNotAllowed) }
