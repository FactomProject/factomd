// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// logger is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level specifies a level of verbosity. The available levels are the eight
// severities described in RFC 5424 and none.
type Level int8

// Levels
const (
	None         Level = iota - 1
	EmergencyLvl       // 1
	AlertLvl           // 2
	CriticalLvl        // 3
	ErrorLvl           // 4
	WarningLvl         // 5
	NoticeLvl          // 6
	InfoLvl            // 7
	DebugLvl           // 8
)

// A FLogger represents an active logging object that generates lines of output
// to an io.Writer.
type FLogger struct {
	out    io.Writer
	level  Level
	prefix string
}

// NewLogFromConfig outputs logs to a file given by logpath
func NewLogFromConfig(logPath, logLevel, prefix string) *FLogger {
	var logFile io.Writer
	if logPath == "stdout" {
		logFile = os.Stdout
	} else {
		logFile, _ = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	}
	return New(logFile, logLevel, prefix)
}

// New makes a new logger with a given level
func New(w io.Writer, level, prefix string) *FLogger {
	return &FLogger{
		out:    w,
		level:  levelFromString(level),
		prefix: prefix,
	}
}

// Level Get the current log level
func (logger *FLogger) Level() (level Level) {
	return logger.level
}

// Println is implemented so this logger shares the same functions as "log"
func (logger *FLogger) Println(args ...interface{}) {
	logger.write(InfoLvl, args...)
}

// Printf is implemented so this logger shares the same functions as "log"
func (logger *FLogger) Printf(format string, args ...interface{}) {
	logger.write(InfoLvl, fmt.Sprintf(format, args...))
}

// Emergency logs with an emergency level and exits the program.
func (logger *FLogger) Emergency(args ...interface{}) {
	logger.write(EmergencyLvl, args...)
}

// Emergencyf logs with an emergency level and exits the program.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Emergencyf(format string, args ...interface{}) {
	logger.write(EmergencyLvl, fmt.Sprintf(format, args...))
}

// Alert logs with an alert level and exits the program.
func (logger *FLogger) Alert(args ...interface{}) {
	logger.write(AlertLvl, args...)
}

// Alertf logs with an alert level and exits the program.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Alertf(format string, args ...interface{}) {
	logger.write(AlertLvl, fmt.Sprintf(format, args...))
}

// Critical logs with a critical level and exits the program.
func (logger *FLogger) Critical(args ...interface{}) {
	logger.write(CriticalLvl, args...)
}

// Criticalf logs with a critical level and exits the program.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Criticalf(format string, args ...interface{}) {
	logger.write(CriticalLvl, fmt.Sprintf(format, args...))
}

// Error logs with an error level.
func (logger *FLogger) Error(args ...interface{}) {
	logger.write(ErrorLvl, args...)
}

// Errorf logs with an error level.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Errorf(format string, args ...interface{}) {
	// Do not do overhead of formatting a string if not going to log
	if ErrorLvl > logger.level {
		return
	}
	logger.write(ErrorLvl, fmt.Sprintf(format, args...))
}

// Warning logs with a warning level.
func (logger *FLogger) Warning(args ...interface{}) {
	logger.write(WarningLvl, args...)
}

// Warningf logs with a warning level.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Warningf(format string, args ...interface{}) {
	// Do not do overhead of formatting a string if not going to log
	if WarningLvl > logger.level {
		return
	}
	logger.write(WarningLvl, fmt.Sprintf(format, args...))
}

// Notice logs with a notice level.
func (logger *FLogger) Notice(args ...interface{}) {
	logger.write(NoticeLvl, args...)
}

// Noticef logs with a notice level.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Noticef(format string, args ...interface{}) {
	// Do not do overhead of formatting a string if not going to log
	if NoticeLvl > logger.level {
		return
	}
	logger.write(NoticeLvl, fmt.Sprintf(format, args...))
}

// Info logs with an info level.
func (logger *FLogger) Info(args ...interface{}) {
	logger.write(InfoLvl, args...)
}

// Infof logs with an info level.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Infof(format string, args ...interface{}) {
	// Do not do overhead of formatting a string if not going to log
	if InfoLvl > logger.level {
		return
	}
	logger.write(InfoLvl, fmt.Sprintf(format, args...))
}

// Debug logs with a debug level.
func (logger *FLogger) Debug(args ...interface{}) {
	logger.write(DebugLvl, args...)
}

// Debugf logs with a debug level.
// Arguments are handled in the manner of fmt.Printf.
func (logger *FLogger) Debugf(format string, args ...interface{}) {
	// Do not do overhead of formatting a string if not going to log
	if DebugLvl > logger.level {
		return
	}
	logger.write(DebugLvl, fmt.Sprintf(format, args...))
}

// write outputs to the FLogger.out based on the FLogger.level and calls os.Exit
// if the level is <= Error
func (logger *FLogger) write(level Level, args ...interface{}) {
	if level > logger.level {
		return
	}

	l := fmt.Sprint(args...) // get string for formatting
	if level == DebugLvl {
		fmt.Fprintf(logger.out, "%s [%s] %s: %s\n", debugPrefix(), levelPrefix[level], logger.Prefix, l)
	} else {
		fmt.Fprintf(logger.out, "%s [%s] %s: %s\n", time.Now().Format(time.RFC3339), levelPrefix[level], logger.Prefix, l)
	}

	if level <= CriticalLvl {
		os.Exit(1)
	}
}

var levelPrefix = map[Level]string{
	EmergencyLvl: "EMERGENCY",
	AlertLvl:     "ALERT",
	CriticalLvl:  "CRITICAL",
	ErrorLvl:     "ERROR",
	WarningLvl:   "WARNING",
	NoticeLvl:    "NOTICE",
	InfoLvl:      "INFO",
	DebugLvl:     "DEBUG",
}

func levelFromString(levelName string) (level Level) {
	switch levelName {
	case "debug":
		level = DebugLvl
	case "info":
		level = InfoLvl
	case "notice":
		level = NoticeLvl
	case "warning":
		level = WarningLvl
	case "error":
		level = ErrorLvl
	case "critical":
		level = CriticalLvl
	case "alert":
		level = AlertLvl
	case "emergency":
		level = EmergencyLvl
	case "none":
		level = None
	default:
		fmt.Fprintf(os.Stderr, "Invalid level value %q, allowed values are: debug, info, notice, warning, error, critical, alert, emergency and none\n", levelName)
		fmt.Fprintln(os.Stderr, "Using log level of warning")
		level = WarningLvl
	}
	return
}
