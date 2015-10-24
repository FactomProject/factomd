package log

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var LogLevel int

const (
	StandardLog = 0
	DebugLog    = 1
)

func init() {
	LogLevel = StandardLog
}

func SetLevel(level string) {
	if strings.ToLower(level) == "debug" {
		LogLevel = DebugLog
	} else {
		LogLevel = StandardLog
	}
}

func debugPrefix() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}
	return file + ":" + strconv.Itoa(line) + " - "
}

func Fatal(str string, args ...interface{}) {
	printf(LogLevel == DebugLog, str+"\n", args...)
	os.Exit(1)
}

func Print(str string) {
	printf(LogLevel == DebugLog, str+"\n")
}

func Println(str string) {
	printf(LogLevel == DebugLog, str+"\n")
}

func Printf(format string, args ...interface{}) {
	printf(LogLevel == DebugLog, format, args...)
}

func Printfln(format string, args ...interface{}) {
	printf(LogLevel == DebugLog, format+"\n", args...)
}

func Debug(format string, args ...interface{}) {
	printf(true, format+"\n", args...)
}

func printf(debug bool, format string, args ...interface{}) {
	if debug {
		format = debugPrefix() + format
	}
	if len(args) > 0 {
		fmt.Printf(format, args...)
	} else {
		fmt.Print(format)
	}
}
