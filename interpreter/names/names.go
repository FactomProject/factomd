package names

import . "github.com/FactomProject/electiontesting/interpreter/common"

type Name uint32

const bitFlags = 3
const bitMask = (-1 << bitFlags)

func (n Name) GetRawName() Name {
	return Name(int(n) & bitMask)
}

func (n Name) GetFlags() FlagsStruct {
	var f FlagsStruct
	f.Executable = (int(n) & (1 << 0)) != 0
	f.Immediate = (int(n) & (1 << 1)) != 0
	f.Traced = (int(n) & (1 << 2)) != 0
	return f
}

func (n *Name) SetFlags(f FlagsStruct) {

	if f.Executable {
		*n |= 1 << 0
	}
	if f.Immediate {
		*n |= 1 << 1
	}
	if f.Traced {
		*n |= 1 << 2
	}
	return
}

func (n Name) IsExecutable() bool { return (int(n) & (1 << 0)) != 0 }

func (n Name) MakeExecutable() Name { return Name(int(n) | 1) }

type NameManager struct {
	n2s map[Name]string
	s2n map[string]Name
}

func NewNameManager() NameManager {
	var nm NameManager
	nm.n2s = make(map[Name]string)
	nm.s2n = make(map[string]Name)
	return nm
}

func (nm *NameManager) GetRawName(n Name) Name {
	return Name(int(n) & bitMask)
}

func (nm *NameManager) GetString(n Name) string {
	s, ok := nm.n2s[n.GetRawName()] // Mask the executable and trace bits
	if !ok {
		panic("GetString(): undefined name " + s)
	}
	if int(n)&1 != 0 {
		return s
	}

	return "/" + s
}

func (nm *NameManager) GetName(s string) Name {
	var executable int = 1
	if s[0] == '/' {
		s = s[1:]      // remove leading /
		executable = 0 // not an executable name
	}
	n, ok := nm.s2n[s]
	// Add missing names
	if !ok {
		n = Name((len(nm.n2s) + 1) << bitFlags) // +1 so we don't use ID 0 shifted over to leave room for flags
		nm.n2s[n] = s
		nm.s2n[s] = n
	}
	return Name(int(n) | executable)
}
