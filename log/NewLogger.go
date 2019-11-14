package main

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/FactomProject/factomd/util/atomic"
)

type delay_format func() string

func (d delay_format) String() string {
	return d()
}

func delay_formater(formatstring string, values ...interface{}) delay_format {
	fmt.Println("creating format")
	return func() string { fmt.Println("executing format"); return fmt.Sprintf(formatstring, values...) }
}

type format_func func(interface{}) string
type logdata map[string]interface{}

func formatter(format string) func(interface{}) string {
	return func(v interface{}) string { return fmt.Sprintf(format, v) }
}

func main() {
	foo := NewLayerLogger(NewSequenceLogger(NewFileLogger(".", []string{"thread", "filename"})), map[string]string{"thread": "fnode0"})
	foo.AddPrintField("foo", formatter("%6v"), "FOO").AddPrintField("bar", formatter("|%6s|"), "BAR")

	foo.Log(logdata{"filename": "test.log", "foo": 1, "bar": "bar", "baz": 5.0})

}

type Ilogger interface {
	Log(logdata) bool // return true if enabled
	GetPrintFieldOrder() []string
	AddPrintField(name string, format format_func, defaultValue string) Ilogger
}

type FileLogger struct {
	traceMutex    sync.Mutex
	files         map[string]*os.File    // cache of file handles
	enabled       map[string]bool        // cache of enable/disable judgments for each filename
	testRegex     *regexp.Regexp         // regex that defined log enable matched against the filename
	nameFields    []string               // list of fields concatenated to form log file name
	printFields   []string               // list of keys printed for each line
	formats       map[string]format_func // set of formatting function for each field
	defaultValues map[string]string      // insure default value strings are equal length to formatted values
}

func NewFileLogger(regexString string, filekeys []string) *FileLogger {
	testRegex, err := regexp.Compile(regexString)
	if err != nil {
		panic(err)
	}
	return &FileLogger{
		files:         make(map[string]*os.File),
		enabled:       make(map[string]bool),
		testRegex:     testRegex,
		nameFields:    filekeys,
		printFields:   nil,
		formats:       make(map[string]format_func),
		defaultValues: make(map[string]string),
	}
}

// Get the list of key in the order they are printed.
func (f *FileLogger) GetPrintFieldOrder() []string {
	return append([]string(nil), f.printFields...)
}

// append a field, define it's format and it's default value to the list of fields printed
// the default value is defined as a string and can be the empty string or a string equal to the formatted value length
func (f *FileLogger) AddPrintField(name string, format format_func, defaultValue string) Ilogger {
	f.printFields = append(f.printFields, name)
	f.formats[name] = format
	f.defaultValues[name] = defaultValue
	return f
}

// Given a name and value and ok (value is valid) return the string for the v or the default value if not valid
// or the empty string in the no value and no default value case.
// assumes lock is already locked
func (f *FileLogger) getValue(n string, v interface{}, ok bool) string {
	if !ok {
		defaultValue, ok := f.defaultValues[n]
		if !ok {
			return "" // no value and no default value just return empty string
		}
		return defaultValue
	}
	formatter, ok := f.formats[n]
	if !ok || formatter == nil {
		return fmt.Sprintf("%v", v) // format value with sprintf
	}
	return formatter(v) // format value or default value with formatter
}

// Get the already open file or open it if it is enabled
// return nil for disabled files
// assumes lock is already locked
func (f *FileLogger) getFile(filename string) *os.File {
	enabled, ok := f.enabled[filename]
	if !ok {
		enabled = f.testRegex.MatchString(filename)
		f.enabled[filename] = enabled
	}
	if !enabled {
		return nil
	}

	if f.files == nil {
		f.files = make(map[string]*os.File)
	}
	file := f.files[filename]
	if file != nil {
		_, err := os.Stat(filename)
		if os.IsNotExist(err) {
			// The file was deleted out from under us
			file.Close() // close the old log
			file = nil   // make the code reopen the file
		}
	}
	if file == nil {
		fmt.Println("Creating " + filename)
		var err error
		file, err = os.Create(filename)
		if err != nil {
			panic(err)
		}
		f.files[filename] = file
		file.WriteString(filename + " " + time.Now().String() + "\n")
	}
	return file
}

// Log the fields in map m to a file built from fields
// return true if the file was enabled
func (f *FileLogger) Log(m logdata) bool {
	f.traceMutex.Lock()
	defer f.traceMutex.Unlock()

	filename := ""
	// loop thru the nameFields and concatenate them all together
	for _, n := range f.nameFields {
		v, ok := m[n]
		value := f.getValue(n, v, ok)
		if filename == "" && value != "" {
			filename = value
		} else if value != "" {
			filename = filename + "_" + value
		}
		delete(m, n)
	}

	file := f.getFile(filename)
	if file == nil {
		return false
	}

	s := ""
	// loop thru the printFields and concatenate them all together
	for _, n := range f.printFields {
		v, ok := m[n]
		value := f.getValue(n, v, ok)
		s = s + value
		delete(m, n)
	}
	// loop thru any unhandled fields and print them
	for n, v := range m {
		s = s + n + ":" + fmt.Sprintf("%v", v)
	}
	file.WriteString(s + "\n")
	return true
}

// add a sequence number and an timestamp to every logged line
type SequenceLogger struct {
	Ilogger
	sequence atomic.AtomicUint32
}

func NewSequenceLogger(logger Ilogger) *SequenceLogger {
	sequenceLogger := SequenceLogger{logger, 0}
	sequenceLogger.AddPrintField("sequence",
		func(v interface{}) string {
			i := v.(uint32)
			return fmt.Sprintf("%7d", i)
		},
		"unset_seq")
	sequenceLogger.AddPrintField("timestamp",
		func(v interface{}) string {
			now := v.(time.Time)
			return fmt.Sprintf("%02d:%02d:%02d.%03d", now.Hour()%24, now.Minute()%60, now.Second()%60, (now.Nanosecond()/1e6)%1000)
		},
		"unset_seq")

	return &sequenceLogger
}

func (f *SequenceLogger) Log(m logdata) bool {
	m["sequence"] = f.sequence.Load()
	m["timestamp"] = time.Now().Local()
	return f.Ilogger.Log(m)
}

// Add some fields with values and call an underlying logger
type LayerLogger struct {
	Ilogger
	fields map[string]string
}

func NewLayerLogger(logger Ilogger, fields map[string]string) *LayerLogger {
	return &LayerLogger{logger, fields}
}

func (f *LayerLogger) Log(m logdata) bool {
	// Copy the layers fields into the instance
	for k, v := range f.fields {
		m[k] = v
	}
	return f.Ilogger.Log(m)
}
