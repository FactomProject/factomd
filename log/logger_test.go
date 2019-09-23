package log_test

import (
	"github.com/FactomProject/factomd/log"
	"testing"
)

func TestLogPrintf(t *testing.T) {
	log.LogPrintf("testing", "unittest %v", "FOO")
}