package factoid_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/FactomProject/factomd/common/factoid"
)

func TestFBlockDump(t *testing.T) {
	var i uint32
	i = 1
	f := NewFBlock(1, i)
	str := f.String()
	line := findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
	i = 10
	f = NewFBlock(1, i)
	str = f.String()
	line = findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
	i = 255
	f = NewFBlock(1, i)
	str = f.String()
	line = findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
	i = 0xFFFF
	f = NewFBlock(1, i)
	str = f.String()
	line = findLine(str, "DBHeight")
	if strings.Contains(line, fmt.Sprintf("%v", i)) == false {
		t.Errorf("Did not find proper height for %v", i)
	}
}

func findLine(full, toFind string) string {
	strs := strings.Split(full, "\n")
	for _, v := range strs {
		if strings.Contains(v, toFind) {
			return v
		}
	}
	return ""
}
