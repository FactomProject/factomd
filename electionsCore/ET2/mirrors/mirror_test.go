package mirrors_test

import (
	"testing"
	. "github.com/FactomProject/factomd/electionsCore/ET2/mirrors"
	"time"
	"io/ioutil"
	"os"
	"fmt"
)

func makeH(s string) [32]byte {
	var h [32]byte
	copy(h[:], s[:])
	return h
}

func AddAndCheckMirror(m Mirrors, l []string, t *testing.T) {
	for _, x := range l {
		if m.IsMirror(makeH(x)) {
			t.Errorf(`Found "%s" in empty mirrorlist`, x)
		}
		if !m.IsMirror(makeH(x)) {
			t.Errorf(`Missing "%s" in mirrorlist that has "%s"`, x, x)
		}
	}
}

func AddToMirror(m1 Mirrors, l2 []string, t *testing.T) {
	for _, x := range l2 {
		m1.Add(makeH(x))
	}
	_ = t // For consistency
}

func CheckMirror(m Mirrors, l []string, t *testing.T) {
	for _, x := range l {
		if !m.IsMirror(makeH(x)) {
			t.Errorf(`Missing "%s" in mirrorlist that has "%s"`, x, x)
		}
	}
}

func TestMirrors1(t *testing.T) {
	var m Mirrors

	m.Init("m1")

	var l []string = []string{"fred", "barney", "wilma", "betty", "bambam", "pebbles"}
	AddAndCheckMirror(m, l, t)
	m.Done()
}

func TestMirrors2(t *testing.T) {
	var m Mirrors

	m.Init("save")

	// Load some mirrors
	var l []string = []string{"fred", "barney", "wilma", "betty", "bambam", "pebbles"}
	AddToMirror(m, l, t)

	file, err := ioutil.TempFile(os.TempDir(), "MirrorTesting")
	if err != nil {
		t.Errorf(`Create TempFile failed:%v`, err)
	}
	file.Close()
	defer os.Remove(file.Name())

	m.Save(file.Name())

	m.Init("load")

	x := "fred"
	if m.IsMirror(makeH(x)) {
		t.Errorf(`Found "%s" in empty mirrorlist`, x)
	}
	x = "dino"
	if m.IsMirror(makeH(x)) {
		t.Errorf(`Found "%s" in empty mirrorlist`, x)
	}

	m.Load("test.sav")
	CheckMirror(m, l, t)

	x = "fred"
	if !m.IsMirror(makeH(x)) {
		t.Errorf(`Did not find "%s" in mirrorlist`, x)
	}
	x = "dino"
	if !m.IsMirror(makeH(x)) {
		t.Errorf(`Did not find "%s" in mirrorlist`, x)
	}
	m.Done()
}

//Check network share functionality
func TestMirrors3(t *testing.T) {
	var m1, m2 Mirrors
	m1.Init("m1")
	m2.Init("m2")

	m2.Listen("8080")
	m1.Connect("localhost:8080")

	for m1.MaxConn == 0 || m2.MaxConn == 0 {
		time.Sleep(1 * time.Second) // Insure the connection have been made
	}

	var l2 []string = []string{"00", "01", "10", "11"}
	AddToMirror(m1, l2, t)
	time.Sleep(1 * time.Second) // Gotta let the other threads run

	//Start Check of m2
	CheckMirror(m2, l2, t)

	// Test connecting a with a preloaded mirror
	var m3 Mirrors
	m3.Init("m3")
	m3.Connect("localhost:8080")
	time.Sleep(1 * time.Second)
	for _, x := range l2 {
		if !m3.IsMirror(makeH(x)) {
			t.Errorf(`%s:Missing "%s" in mirrorlist that should have "%s"`, m3.Name, x, x)
		}
	}
	m1.Done()
	m2.Done()
}

// Test connecting a with a preloaded mirror
func TestMirrors4(t *testing.T) {
	var m1, m2 Mirrors
	m1.Init("src")
	m2.Init("dst")

	// Load some mirrors
	var l []string = []string{"fred", "barney", "wilma", "betty", "bambam", "pebbles"}
	AddAndCheckMirror(m1, l, t)

	m2.Listen("8080")
	m1.Connect("localhost:8080")

	time.Sleep(1 * time.Second) // Gotta let the other threads run

	CheckMirror(m2, l, t)
	m1.Done()
	m2.Done()
}

// Test connecting a with both mirrors preloaded
func TestMirrors5(t *testing.T) {
	var m1, m2 Mirrors
	m1.Init("m1")
	m2.Init("m2")

	// Load some mirrors
	var l1 []string = []string{"fred", "barney", "wilma", "betty", "bambam", "pebbles"}
	var l2 []string = []string{"00", "01", "10", "11"}
	AddAndCheckMirror(m1, l1, t)
	AddAndCheckMirror(m2, l2, t)

	m2.Listen("8080")
	m1.Connect("localhost:8080")

	time.Sleep(1 * time.Second) // Gotta let the other threads run

	for _, x := range append(l1, l2...) {
		if !m1.IsMirror(makeH(x)) {
			t.Errorf(`%s:Missing "%s" in mirrorlist that should have "%s"`, m1.Name, x, x)
		}
		if !m2.IsMirror(makeH(x)) {
			t.Errorf(`%s:Missing "%s" in mirrorlist that should have "%s"`, m2.Name, x, x)
		}
	}

	m1.Done()
	m2.Done()
}

func TestPrimeWalk(t *testing.T) {
	var visited [119]bool
	const prime = 7907
	i:=0
	for j:=0;j<len(visited);j++{
		i = (i+prime)%len(visited)
		fmt.Printf("%d ", i)
		if visited[i] {
			t.Errorf(`Found issue %d %d`,i,j)
		}
		visited[i]=true
	}


}