package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/FactomProject/factomd/util/atomic"
)

type delay_format func() string

func (d delay_format) String() string {
	return d()
}

// Create a function to format data only if it is printed or logged
func Delay_formater(formatstring string, values ...interface{}) delay_format {
	return func() string { return fmt.Sprintf(formatstring, values...) }
}

type FormatFunc func(interface{}) string // a function that will format a bit of data
type LogData map[string]interface{}      // a map if key value pairs to log

func Formatter(format string) func(interface{}) string {
	return func(v interface{}) string { return fmt.Sprintf(format, v) }
}

func main() {
	foo := NewModuleLoggerLogger(NewLayerLogger(NewSequenceLogger(NewFileLogger(".")), map[string]string{"thread": "fnode0"}), "test.txt")
	foo.AddNameField("logname", Formatter("%s"), "unknown_log")
	foo.AddPrintField("foo", Formatter("%6v"), "FOO").AddPrintField("bar", Formatter("|%6s|"), "BAR")

	foo.Log(LogData{"foo": 1, "bar": "bar", "baz": 5.0})

}

type Ilogger interface {
	Log(LogData) bool // return true if logging of specified log data is enabled
	GetPrintFieldOrder() []string
	AddPrintField(name string, format FormatFunc, defaultValue string) Ilogger
	AddNameField(name string, format FormatFunc, defaultValue string) Ilogger
}

//todo: separate logging enabling to a separate layer from filelogger

type FileLogger struct {
	traceMutex    sync.Mutex
	files         map[string]*os.File   // cache of file handles
	enabled       map[string]bool       // cache of enable/disable judgments for each filename
	testRegex     *regexp.Regexp        // regex that defined log enable matched against the filename
	nameFields    []string              // ordered list of fields to concatenated to form log file name
	printFields   []string              // ordered list of keys printed for each line
	formats       map[string]FormatFunc // set of formatting function for each field
	defaultValues map[string]string     // insure default value strings are equal length to formatted values
	rootDir       string                // path to files
}

// Create a new FileLogger
// regexString is a combination of both the path to the log directory and the regex that decides if the data is logged
func NewFileLogger(regexString string) *FileLogger {
	var logDir string
	var testRegex *regexp.Regexp
	var err error

	if regexString != "" {
		//strip quotes if they are included in the string
		lastChar := len(regexString) - 1
		if (regexString[0] == '"' && regexString[lastChar] == '"') ||
			(regexString[0] == '\'' && regexString[lastChar] == '\'') {
			regexString = regexString[1:lastChar] // Trim the quote characters
		}
		logDir, regexString = filepath.Split(regexString)

		//strip quotes if they are included in the regex portion of the string
		if (regexString[0] == '"' && regexString[lastChar] == '"') ||
			(regexString[0] == '\'' && regexString[lastChar] == '\'') {
			regexString = regexString[1:lastChar] // Trim the quote characters
		}

		testRegex, err = regexp.Compile(regexString)
		if err != nil {
			panic(err)
		}
	}
	return &FileLogger{
		files:         make(map[string]*os.File),
		enabled:       make(map[string]bool),
		testRegex:     testRegex,
		nameFields:    nil,
		printFields:   nil,
		formats:       make(map[string]FormatFunc),
		defaultValues: make(map[string]string),
		rootDir:       logDir,
	}
}

// Get the list of key in the order they are printed.
func (f *FileLogger) GetPrintFieldOrder() []string {
	return append([]string(nil), f.printFields...)
}

// Get the list of key in the order they are used to build the name.
func (f *FileLogger) GetNameFieldOrder() []string {
	return append([]string(nil), f.nameFields...)
}

// prepend a field, define it's format and it's default value to the list of fields used to build the log name
// the default value is defined as a string and can be the empty string or a string equal to the formatted value length
func (f *FileLogger) AddNameField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.nameFields = append([]string{name}, f.nameFields...) // prepend name fields
	f.formats[name] = format
	f.defaultValues[name] = defaultValue
	return f
}

// append a field, define it's format and it's default value to the list of fields printed
// the default value is defined as a string and can be the empty string or a string equal to the formatted value length
func (f *FileLogger) AddPrintField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.printFields = append(f.printFields, name) // append print fields
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
	return formatter(v) // format value or default value with Formatter
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
		fullpath := f.rootDir + filename
		fmt.Println("Creating " + fullpath)
		dir := filepath.Dir(fullpath)
		if dir != "" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					panic(err)
				}
			}
		}
		var err error
		file, err = os.Create(fullpath)
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
func (f *FileLogger) Log(m LogData) bool {
	f.traceMutex.Lock()
	defer f.traceMutex.Unlock()

	// make a local copy of the map
	var x LogData = make(LogData)
	for k, v := range m {
		x[k] = v
	}

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
	filename = strings.ToLower(filename)
	file := f.getFile(filename)
	if file == nil {
		return false
	}

	if false {
		s := ""
		for n, v := range m {
			s = s + "--" + n + ":" + f.getValue(n, v, true) + "-- "
		}
		file.WriteString(s + "\n")

		var v string
		t := reflect.TypeOf(m["comment"]).String()
		if t == "string" {
			v = m["comment"].(string)
		}
		if t == "logging.delay_format" {
			v = m["comment"].(delay_format).String()
			if strings.Contains(v, "SignDB") {
				fmt.Println("?")
			}
		}
		fmt.Println(v)
	}
	s := ""
	// loop thru the printFields and concatenate them all together
	for _, n := range f.printFields {
		v, ok := m[n]
		value := f.getValue(n, v, ok)
		if s == "" {
			s = value
		} else if value != "" {
			s = s + " " + value
		}
		delete(m, n)
	}

	// loop thru any unhandled fields and print them
	for n, v := range m {
		s = s + "--" + n + ":" + fmt.Sprintf("%v", v) + "-- "
	}
	file.WriteString(s + "\n")
	return true
}

// unused -- os.File is written by direct calls to write and not buffered and the os closes the files on exit.
func (f *FileLogger) Cleanup() {
	f.traceMutex.Lock()
	defer f.traceMutex.Unlock()
	for filename, file := range f.files {
		delete(f.files, filename)
		file.WriteString("Close " + filename + " " + time.Now().String() + "\n")
		file.Close()
	}
}

// SequenceLogger add a sequence number and an timestamp to every logged line
type SequenceLogger struct {
	Ilogger
	sequence atomic.AtomicUint32
}

func (f *SequenceLogger) Log(m LogData) bool {
	m["sequence"] = f.sequence.Add(1)
	m["timestamp"] = time.Now().Local()
	return f.Ilogger.Log(m)
}

// just letting the call fall info the underlying logger changed the return type to be the underlying logger type
func (f *SequenceLogger) AddPrintField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.Ilogger.AddPrintField(name, format, defaultValue)
	return f
}

// just letting the call fall info the underlying logger changed the return type to be the underlying logger type
func (f *SequenceLogger) AddNameField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.Ilogger.AddNameField(name, format, defaultValue)
	return f
}

func NewSequenceLogger(logger Ilogger) *SequenceLogger {
	sequenceLogger := SequenceLogger{logger, 0}
	sequenceLogger.AddPrintField("sequence",
		func(v interface{}) string {
			i := v.(uint32)
			return fmt.Sprintf("%9d", i)
		},
		"unset_seq")
	sequenceLogger.AddPrintField("timestamp",
		func(v interface{}) string {
			now := v.(time.Time)
			return fmt.Sprintf("%02d:%02d:%02d.%03d", now.Hour()%24, now.Minute()%60, now.Second()%60, (now.Nanosecond()/1e6)%1000)
		},
		"unset_ts")

	return &sequenceLogger
}

// Add some fields with values and call an underlying logger
type LayerLogger struct {
	Ilogger
	fields map[string]string
}

func NewLayerLogger(logger Ilogger, fields map[string]string) *LayerLogger {
	return &LayerLogger{logger, fields}
}

// just letting the call fall info the underlying logger changed the return type to be the underlying logger type
func (f *LayerLogger) AddPrintField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.Ilogger.AddPrintField(name, format, defaultValue)
	return f
}

// just letting the call fall info the underlying logger changed the return type to be the underlying logger type
func (f *LayerLogger) AddNameField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.Ilogger.AddNameField(name, format, defaultValue)
	return f
}

func (f *LayerLogger) Log(m LogData) bool {
	// Copy the layers fields into the instance
	for k, v := range f.fields {
		m[k] = v
	}
	return f.Ilogger.Log(m)
}

// Add filename fields and skip calling if the file is disabled.
// used for logging inside a routine. these should have limited life so changes in logging mode are quickly reflected
type ModuleLogger struct {
	Ilogger
	filename string
	enabled  bool // cache log enabled status for a filename
}

func NewModuleLoggerLogger(logger Ilogger, filename string) *ModuleLogger {
	return &ModuleLogger{logger, filename, true}
}

// just letting the call fall info the underlying logger changed the return type to be the underlying logger type
func (f *ModuleLogger) AddPrintField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.Ilogger.AddPrintField(name, format, defaultValue)
	return f
}

// just letting the call fall info the underlying logger changed the return type to be the underlying logger type
func (f *ModuleLogger) AddNameField(name string, format FormatFunc, defaultValue string) Ilogger {
	f.Ilogger.AddNameField(name, format, defaultValue)
	return f
}

func (f *ModuleLogger) Log(m LogData) bool {
	if f.enabled {
		m["logname"] = f.filename // add the filename this module is logging to
		f.enabled = f.Ilogger.Log(m)
	}
	return f.enabled
}
