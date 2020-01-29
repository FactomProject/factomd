package p2p

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func TestNewLimitedListener(t *testing.T) {
	type args struct {
		address string
		limit   time.Duration
	}
	tests := []struct {
		name    string
		args    args
		pattern string
		wantErr bool
	}{
		{"localhost1", args{":0", time.Second}, "^\\[::\\]:\\d+", false},
		{"localhost2", args{"127.0.0.1:0", time.Second}, "^127\\.0\\.0\\.1:\\d+", false},
		{"negative time", args{":0", -time.Second}, "", true},
		{"wrong ip", args{"8.8.8.8:0", -time.Second}, "", true}, // google's
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLimitedListener(tt.args.address, tt.args.limit)
			if got != nil {
				defer got.Close()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("LimitedListen() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil { // done at this point
				return
			}

			if match, err := regexp.MatchString(tt.pattern, got.Addr().String()); !match && err == nil {
				t.Errorf("LimitedListen() address = %v, want %v (compile error %v)", got.Addr().String(), tt.pattern, err)
			}
		})
	}
}

func createLLHistory() *LimitedListener {
	past := limitedConnect{"past", time.Now().Add(time.Hour * -2)}
	presence := limitedConnect{"presence", time.Now()}
	future := limitedConnect{"future", time.Now().Add(time.Hour * 2)}
	return &LimitedListener{
		listener:       nil,
		limit:          time.Hour,
		lastConnection: time.Now(),
		history:        []limitedConnect{past, presence, future}, // order matters, old < new
	}
}

func TestLimitedListener_clearHistory(t *testing.T) {
	ll := createLLHistory()

	ll.clearHistory() // removes past
	if len(ll.history) != 2 {
		t.Errorf("Did not remove past connection properly")
		fmt.Println(ll.history)
	}

	ll.lastConnection = time.Now().Add(time.Hour * -5)
	ll.clearHistory()
	if len(ll.history) != 0 {
		t.Errorf("Did not truncate history properly")
	}

}

func TestLimitedListener_isInHistory(t *testing.T) {
	ll := createLLHistory()
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		ll   *LimitedListener
		args args
		want bool
	}{
		{"past", ll, args{"past"}, false},
		{"presence", ll, args{"presence"}, true},
		{"future", ll, args{"future"}, true},
		{"unknown", ll, args{"unknown"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ll.isInHistory(tt.args.addr); got != tt.want {
				t.Errorf("LimitedListener.isInHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLimitedListener_addToHistory(t *testing.T) {
	ll := createLLHistory()
	ll.history = nil // reset history for this one
	ll.lastConnection = time.Time{}

	type args struct {
		addr string
	}
	tests := []struct {
		name string
		ll   *LimitedListener
		args args
	}{
		{"adding anything", ll, args{"anything"}},
		{"adding anything again", ll, args{"anything"}},
		{"adding something else", ll, args{"something else"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ll.addToHistory(tt.args.addr)
			if !tt.ll.isInHistory(tt.args.addr) {
				t.Errorf("%s did not get added", tt.args.addr)
			}
			if !time.Now().Add(-time.Second).Before(tt.ll.lastConnection) { // 1 second margin of error
				t.Errorf("lastConnection did not update: now=%v, lastConnection=%v", time.Now(), tt.ll.lastConnection)
			}
		})
	}
}

func TestLimitedListener_Accept(t *testing.T) {
	tests := []struct {
		name    string
		ll      *LimitedListener
		want    net.Conn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.ll.Accept()
			if (err != nil) != tt.wantErr {
				t.Errorf("LimitedListener.Accept() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LimitedListener.Accept() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewLimitedListener(t *testing.T) {
	// testing on actual tcp connections
	// servers runs on 127.0.0.1:0
	// connections will be made via 127.0.0.x:0

	ll, err := NewLimitedListener("127.0.0.1:0", time.Millisecond*10)
	if err != nil {
		t.Fatalf("Error starting listener: %v", err)
	}
	go func() {
		for {
			con, err := ll.Accept()
			if err != nil { // test over
				break
			}
			con.Close()
		}
	}()
	defer ll.Close()

	raddr, err := net.ResolveTCPAddr("tcp", ll.Addr().String()) // server address with dynamic socket
	if err != nil {
		t.Fatalf("Unable to resolve server address: %v", err)
	}

	time.Sleep(time.Millisecond) // wait for start to listen

	for i := 1; i < 6; i++ {
		a := fmt.Sprintf("127.0.0.%d", i)
		laddr, err := net.ResolveTCPAddr("tcp", a+":0")
		if err != nil {
			t.Errorf("unable to resolve local address. wanted: %s, got: %v", a, err)
			continue
		}
		con, err := net.DialTCP("tcp", laddr, raddr)
		if err != nil {
			t.Errorf("unable to connect to %s with addr %s", raddr, laddr)
			continue
		}
		defer con.Close()
	}

	bad, err := net.Dial("tcp", ll.Addr().String()) // connection from .1

	time.Sleep(time.Millisecond) // give a little time

	if err != nil {
		// listener accepts the connection fine but closes it right away
		t.Errorf("Bad connection was unable to connect: %v", err)
	} else {
		one := make([]byte, 8)
		bad.SetReadDeadline(time.Now().Add(time.Millisecond))
		n, err := bad.Read(one) // tcp connection should be closed already
		if err != io.EOF {
			t.Errorf("Bad connection was not closed right away, got: %d bytes read, %v", n, err)
		}
		bad.Close()
	}

	for i := 1; i < 6; i++ {
		a := fmt.Sprintf("127.0.0.%d", i)
		if !ll.isInHistory(a) {
			t.Errorf("Address %s is not in the history after 1ms", a)
		}
	}

	time.Sleep(time.Millisecond * 10) // pass the limit

	for i := 1; i < 6; i++ {
		a := fmt.Sprintf("127.0.0.%d", i)
		if ll.isInHistory(a) {
			t.Errorf("Address %s is still in the history after 11ms", a)
		}
	}
}
