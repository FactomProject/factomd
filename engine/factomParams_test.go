package engine_test

import "testing"

func TestParamHandling(t *testing.T) {
	args := append([]string{},
		"-db=Map",
		"-network=LOCAL",
		"-enablenet=true",
		"-blktime=15",
		"-count=10",
		"-logPort=37000",
		"-port=37001",
		"-ControlPanelPort=37002",
		"-networkPort=37003")

	p := ParseCmdLine(args)
	if p.db

}
