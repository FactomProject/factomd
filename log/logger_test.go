package log_test

import (
	"testing"

	"github.com/PaulSnow/factom2d/log"
)

func TestLogPrintf(t *testing.T) {
	log.LogPrintf("testing", "unittest %v", "FOO")
}
