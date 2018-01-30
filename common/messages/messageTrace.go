package messages

import (
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"os"
	"sync"
	"github.com/FactomProject/factomd/common/constants"
)

var (
	traceMutex sync.Mutex
	files      map[string]*os.File
	sequence   int
)

// assumes traceMutex is locked already
func getTraceFile(name string) (f *os.File) {
	//traceMutex.Lock()	defer traceMutex.Unlock()
	if files == nil {
		files = make(map[string]*os.File)
	}
	f, _ = files[name]
	if f == nil {
		fmt.Println("Creating " + name)
		var err error
		f, err = os.Create(name)
		if err != nil {
			panic(err)
		}
		files[name] = f
	}
	return f
}

func LogMessage(name string, note string, msg interfaces.IMsg) {

	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	sequence++
	seq := sequence
	t := msg.Type()
	embeddedHash := ""

	switch msg.Type() {
	case constants.ACK_MSG:
		m := msg.(*Ack)
		embeddedHash = fmt.Sprintf(" EmbeddedMsg %26v[%2v]:%v", MessageName(m.Type()), m.Type(), m.MessageHash.String()[:8])
	case constants.MISSING_MSG_RESPONSE:
		m := msg.(*MissingMsgResponse).MsgResponse
		embeddedHash = fmt.Sprintf(" EmbeddedMsg %26v[%2v]:%v", MessageName(m.Type()), m.Type(), m.GetHash().String()[:8])
	}


	myfile.WriteString(fmt.Sprintf("%5v %20s %v %26s[%2v]:%v%v {%v}\n", seq, note, msg.GetMsgHash().String()[:8], MessageName(byte(t)), t, msg.GetHash().String()[:8], embeddedHash, msg.String()))

}

// stringify it in the caller to avoid having to deal with the import loop
func LogParcel(name string, note string, msg string) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	sequence++
	seq := sequence

	myfile.WriteString(fmt.Sprintf("%5v %26s %s\n", seq, note, msg))
}

// unused -- of.File is written by direct calls to write and not buffered and the os closes the files on exit.
func Cleanup() {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	for name, f := range files {
		delete(files, name)
		f.Close()
	}
}
