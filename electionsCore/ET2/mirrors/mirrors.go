package mirrors

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"sync"
)

var debug bool = false

// Check an error code and print a message if there was an error
func check(err error, format string, message ...interface{}) bool {
	if err != nil {
		s := err.Error()
		_ = s
		log.Printf("%v:%s\n", fmt.Sprintf(format, message...), err)
	}
	return err != nil
}

// Check an error code and print a message if there was an error and exit
func checkOrDie(err error, format string, message ...interface{}) {
	if err != nil {
		s := err.Error()
		_ = s
		log.Fatalf("%v:%s\n", fmt.Sprintf(format, message...), err)
	}
}

type Mirrors struct {
	Name        string
	mapMu       sync.RWMutex
	mirrorMap   map[[32]byte]int
	connMu      sync.RWMutex
	connections map[int]net.Conn
	MaxConn     int
	l           net.Listener
}

func (m *Mirrors) Len() int {
	m.mapMu.Lock()
	defer m.mapMu.Unlock()
	return len(m.mirrorMap)

}

func (m *Mirrors) Done() {
	if m.l != nil {
		m.l.Close()
	}
	m.connMu.Lock()
	for n, conn := range m.connections {
		delete(m.connections, n)
		conn.Close()
	}
	m.connMu.Unlock()
}

func (m *Mirrors) Init(name string) {
	m.Name = name
	m.mapMu.Lock()
	m.mirrorMap = make(map[[32]byte]int)
	m.mapMu.Unlock()
	m.connMu.Lock()
	m.connections = make(map[int]net.Conn)
	m.connMu.Unlock()

}

// Send a new mirror to  a connection
func (m *Mirrors) SendC(conn net.Conn, h [32]byte) (err error) {
	if debug {
		fmt.Printf("%s:sendc(%s,%x)\n", m.Name, conn.RemoteAddr(), h)
	}
	n, err2 := conn.Write(h[:])
	if err2 == nil && n != 32 {
		err2 = errors.New("Short write to " + conn.RemoteAddr().String())
	}
	if err == nil && err2 != nil {
		err = err2
	}
	return err
}

// Send a new mirror to all our connections
func (m *Mirrors) Send(h [32]byte) (err error) {
	if debug {
		fmt.Printf("%s:send(%x)\n", m.Name, h)
	}
	m.connMu.Lock()
	defer m.connMu.Unlock()
	for _, conn := range m.connections {
		n, err2 := conn.Write(h[:])
		if err2 == nil && n != 32 {
			err2 = errors.New("Short write to " + conn.RemoteAddr().String())
		}
		if err == nil && err2 != nil {
			err = err2
		}
	}
	return err
}

// Add a mirror to our list of mirrors
func (m *Mirrors) Add(h [32]byte) {
	if debug {
		fmt.Printf("%s:add(%x)\n", m.Name, h)
	}
	m.mapMu.Lock()
	m.mirrorMap[h] = 1
	m.mapMu.Unlock()
	m.Send(h)
}

// Check if we have seen a mirror and add it if it is new
func (m *Mirrors) IsMirror(h [32]byte) (rval bool) {
	if debug {
		//		fmt.Printf("%s:ismirror(%x)\n",m.Name, h)
	}

	m.mapMu.RLock()
	x, ok := m.mirrorMap[h]
	m.mapMu.RUnlock()
	rval = x > 0
	if debug {
		fmt.Printf("%s:ismirror(%x)=%v\n", m.Name, h, *&rval)
	}
	if !ok {
		m.Add(h)
	}
	return rval
}

func (m *Mirrors) Save(filename string) error {
	if debug {
		fmt.Printf("%s:save(%s)\n", m.Name, filename)
	}
	f, err := os.Create(filename)
	check(err, "Mirrors.Save(%s) error:", filename)
	if err == nil {
		m.mapMu.RLock() // lock the map for read
		checkOrDie(gob.NewEncoder(f).Encode(m.mirrorMap), "Mirrors.Save(%s) encode error:", filename)
		m.mapMu.RUnlock()
		checkOrDie(f.Close(), "Mirrors.Save(%s) close error:", filename)
	}
	return err
}

func (m *Mirrors) Load(filename string) error {
	if debug {
		fmt.Printf("%s:load(%s)\n", m.Name, filename)
	}
	if m.mirrorMap == nil {
		m.mirrorMap = make(map[[32]byte]int)
	}
	f, err := os.Open(filename)
	check(err, "Mirrors.Load(%s) error:", filename)

	if err == nil {
		m.mapMu.Lock() // Lock the map for write
		checkOrDie(gob.NewDecoder(f).Decode(&m.mirrorMap), "Mirrors.Load(%s) decode error:", filename)
		m.mapMu.Unlock()
		checkOrDie(f.Close(), "Mirrors.Load(%s) close error:", filename)
	}
	return err
}

func (m *Mirrors) Listen(port string) {
	if debug {
		fmt.Printf("%s:listen(%s)\n", m.Name, port)
	}
	// Listen for incoming connections.
	var err error
	m.l, err = net.Listen("tcp", "localhost"+":"+port)
	checkOrDie(err, "Mirrors.Listen error:")

	// spawn a thread to accept connections
	go func() {
		// Close the listener when the application closes.
		defer m.l.Close()
		for {
			// Listen for an incoming connection.
			conn, err := m.l.Accept()
			// Awkward way to detect we are shutting down the listener socket.
			if err != nil {
				if ok, _ := regexp.MatchString("accept tcp .*:.*: use of closed network connection", err.Error()); ok {
					return
				}
			}
			checkOrDie(err, "Mirrors.Listen accept error:")

			// simple handshake to insure the connection is alive
			{
				// When I accept an incoming connection I say hello
				_, err := conn.Write([]byte("Hello"))
				checkOrDie(err, "%s:Send Hello Failed on %s", m.Name, conn.RemoteAddr())
				var hello [5]byte
				_, err = conn.Read(hello[:])
				checkOrDie(err, "%s:Recv Hello Failed on %s", m.Name, conn.RemoteAddr())
				if string(hello[:]) != "Hello" {
					log.Fatalf("%s:Recv Hello Failed on %s got <%s>", m.Name, conn.RemoteAddr(), string(hello[:]))
				}
			}

			// send our map when we connect.
			m.mapMu.Lock()
			if len(m.mirrorMap) > 0 {
				for h, _ := range m.mirrorMap {
					m.SendC(conn, h)
				}
			}
			m.mapMu.Unlock()

			// Handle incoming mirrors in a new goroutine.
			go m.handleRequest(conn)

		}
	}()
}

// Handles connection. Read and add mirrors to the map until the connection closes
func (m *Mirrors) handleRequest(conn net.Conn) {
	if debug {
		fmt.Printf("%s:handleRequest(%v)\n", m.Name, conn.RemoteAddr())
	}
	m.connMu.Lock()
	var myConn = m.MaxConn
	m.MaxConn++
	m.connections[myConn] = conn
	// Add the new connection to my list
	m.connMu.Unlock()
	conn.(*net.TCPConn).SetNoDelay(true) // Should be the default.

	var err error
	for {
		// Make a buffer to hold incoming data.
		var buf [32]byte
		var reqLen int
		// Read the incoming connection into the buffer.
		reqLen, err = conn.Read(buf[:])
		// Normal way to detect the connection has closed
		if err == io.EOF {
			return
		}
		// Awkward way to detect we are shutting down the connection socket.
		if err != nil {
			if ok, _ := regexp.MatchString("read tcp .*: use of closed network connection", err.Error()); ok {
				break
			}
		}
		if check(err, "Error on connection to %v read %d bytes, error:", conn.RemoteAddr(), reqLen) || reqLen != 32 {
			break
		}
		m.IsMirror(buf) // Add the mirror to my map if needed
	}
	if debug {
		fmt.Printf("%s:handleRequest(%v) exiting %v\n", m.Name, conn.RemoteAddr(), err)
	}
	// remove the connection from my list
	m.connMu.Lock()
	delete(m.connections, myConn)
	m.connMu.Unlock()
	// Close the connection when you're done with it.
	conn.Close()
}

func (m *Mirrors) Connect(addressPort string) error {
	if debug {
		fmt.Printf("%s:connect(%v)\n", m.Name, addressPort)
	}
	conn, err := net.Dial("tcp", addressPort)
	checkOrDie(err, "Mirrors.Connect(%s) error:", addressPort)

	// simple handshake to insure the connection is alive
	{
		// When I accept an incoming connection I say hello
		_, err := conn.Write([]byte("Hello"))
		checkOrDie(err, "%s:Send Hello Failed on %s", m.Name, conn.RemoteAddr())
		var hello [5]byte
		_, err = conn.Read(hello[:])
		checkOrDie(err, "%s:Recv Hello Failed on %s", m.Name, conn.RemoteAddr())
		if string(hello[:]) != "Hello" {
			log.Fatalf("%s:Recv Hello Failed on %s got <%s>", m.Name, conn.RemoteAddr(), string(hello[:]))
		}
	}
	// send our map when we connect.
	m.mapMu.Lock()
	if len(m.mirrorMap) > 0 {
		for h, _ := range m.mirrorMap {
			m.SendC(conn, h)
		}
	}
	m.mapMu.Unlock()

	go m.handleRequest(conn)

	return err
}
