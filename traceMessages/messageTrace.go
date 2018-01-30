package traceMessages

import (
	"os"
	"sync"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/messages"
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
		m := msg.(*messages.Ack)
		embeddedHash = fmt.Sprintf(" EmbeddedMsg %20v[%2v]:%v", messages.MessageName(m.Type()), m.Type(), m.MessageHash.String()[:8])
	case constants.MISSING_MSG_RESPONSE:
		m := msg.(*messages.MissingMsgResponse).MsgResponse
		embeddedHash = fmt.Sprintf(" EmbeddedMsg %20v[%2v]:%v", messages.MessageName(m.Type()), m.Type(), m.GetHash().String()[:8])
	}

	myfile.WriteString(fmt.Sprintf("%5v %20s %v %20s[%2v]:%v%v\n", seq, note, msg.GetMsgHash().String()[:8], messages.MessageName(byte(t)), t, msg.GetHash().String()[:8], embeddedHash))
}

// stringify it in the caller to avoid having to deal with the import loop
func LogParcel(name string, note string, msg string) {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	myfile := getTraceFile(name)
	sequence++
	seq := sequence

	myfile.WriteString(fmt.Sprintf("%5v %20s %s\n", seq, note, msg))
}

func Cleanup() {
	traceMutex.Lock()
	defer traceMutex.Unlock()
	for name, f := range files {
		delete(files, name)
		f.Close()
	}
}
//1049                      0xc420cbe100 Missing Msg Response[19]:a1db719314ced94a84b8f1e7bebc0fb563e0434b2115ad78da6a85395f58b5ce EmbeddedMsg Directory Block Signature[ 7]:cc36eb2f78ba0e9fd7501ce7401dd349325d949cf5733fe55a32fe1f8f30eb9d