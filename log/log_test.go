// +build all

package log_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/log"
)

var _ = fmt.Println

func TestLog(t *testing.T) {
}

func TestBadNew(t *testing.T) {
	l := New(nil, "warninggit c", "testing")
	if l.Level() != WarningLvl {
		t.Error("Should be set to warning")
	}
}

func TestNew(t *testing.T) {
	buf := new(bytes.Buffer)

	floggers := make([]*FLogger, 0)
	lDebug := New(buf, "debug", "testing")
	linfo := New(buf, "info", "testing")
	lnotice := New(buf, "notice", "testing")
	lwarning := New(buf, "warning", "testing")
	lerror := New(buf, "error", "testing")
	lcritical := New(buf, "critical", "testing")
	lalert := New(buf, "alert", "testing")
	lemergency := New(buf, "emergency", "testing")
	lnone := New(buf, "none", "testing")

	floggers = append(floggers, lDebug)
	floggers = append(floggers, linfo)
	floggers = append(floggers, lnotice)
	floggers = append(floggers, lwarning)
	floggers = append(floggers, lerror)
	floggers = append(floggers, lcritical)
	floggers = append(floggers, lalert)
	floggers = append(floggers, lemergency)
	floggers = append(floggers, lnone)

	for _, f := range floggers {
		if !check(f, buf, DebugLvl) {
			t.Error("Debug level not working")
		}
		if !check(f, buf, InfoLvl) {
			t.Error("Info level not working")
		}
		if !check(f, buf, NoticeLvl) {
			t.Error("Notice level not working")
		}
		if !check(f, buf, WarningLvl) {
			t.Error("Warning level not working")
		}
		if !check(f, buf, ErrorLvl) {
			t.Error("Error level not working")
		}
	}
}

func check(f *FLogger, out *bytes.Buffer, lvl Level) bool {
	if !checkLevel(f, out, lvl) {
		return false
	}
	if !checkLevelf(f, out, lvl) {
		return false
	}
	return true
}

func checkLevel(f *FLogger, out *bytes.Buffer, lvl Level) bool {
	pre := out.Len()

	switch lvl {
	case DebugLvl:
		f.Debug("Test")
	case InfoLvl:
		f.Info("Test")
	case NoticeLvl:
		f.Notice("Test")
	case WarningLvl:
		f.Warning("Test")
	case ErrorLvl:
		f.Error("Test")
	case CriticalLvl:
	case AlertLvl:
	case EmergencyLvl:
	}

	post := out.Len()
	if f.Level() == None {
		if pre != post {
			return false
		}
	}

	if f.Level() >= lvl {
		if post <= pre {
			return false
		}
	} else {
		if post > pre {
			return false
		}
	}
	return true
}

func checkLevelf(f *FLogger, out *bytes.Buffer, lvl Level) bool {
	pre := out.Len()

	switch lvl {
	case DebugLvl:
		f.Debugf("Test")
	case InfoLvl:
		f.Infof("Test")
	case NoticeLvl:
		f.Noticef("Test")
	case WarningLvl:
		f.Warningf("Test")
	case ErrorLvl:
		f.Errorf("Test")
	case CriticalLvl:
	case AlertLvl:
	case EmergencyLvl:
	}

	post := out.Len()
	if f.Level() == None {
		if pre != post {
			return false
		}
	}

	if f.Level() >= lvl {
		if post <= pre {
			return false
		}
	} else {
		if post > pre {
			return false
		}
	}
	return true
}

func BenchmarkZeroFormat(b *testing.B) {
	buf := new(bytes.Buffer)
	l := New(buf, "debug", "testing")
	doBenchmark(b, l, "zero")
}

func BenchmarkSingleFormat(b *testing.B) {
	buf := new(bytes.Buffer)
	l := New(buf, "debug", "testing")
	doBenchmark(b, l, "%s", "single")
}

func BenchmarkDoubleFormat(b *testing.B) {
	buf := new(bytes.Buffer)
	l := New(buf, "debug", "testing")
	doBenchmark(b, l, "%s, %s", "single", "double")
}

func doBenchmark(b *testing.B, logger *FLogger, format string, args ...interface{}) {
	for i := 0; i < b.N; i++ {
		logger.Debugf(format, args...)
	}
}
func TestPrints(t *testing.T) {
	SetLevel("standard")
	Print("Standard log test Print")
	Println("Standard log test Println")
	Printf("Standard log test %s\n", "Printf")
	Printfln("Standard log test %s", "Printfln")
	Debug("Debug log call %d", 1)
	PrintStack()
	Print("note: the above printout is not an error, this is just printing out a stack trace.")
	SetLevel("debug")
	Print("Debug log test Print")
	UnsetTestLogger()
}
