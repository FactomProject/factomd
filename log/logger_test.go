package log_test

import (
	"testing"

	"github.com/FactomProject/factomd/log"
)

func TestLogPrintf(t *testing.T) {
	log.LogPrintf("testing", "unittest %v", "FOO")
}
