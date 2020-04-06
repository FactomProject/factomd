package p2p

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func Test_controller_persist(t *testing.T) {
	f, err := ioutil.TempFile("", "peerfile*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	pc := new(PeerCache)
	pc.Bans = make(map[string]time.Time)

	pc.Bans["ban null"] = time.Time{}
	pc.Bans["ban outdated"] = time.Now().Add(-time.Second)
	pc.Bans["ban ok"] = time.Now().Add(time.Hour)

	for i := 0; i < 128; i++ {
		pc.Peers = append(pc.Peers, testRandomEndpoint())
	}

	if err := pc.WriteToFile(f.Name()); err != nil {
		t.Fatal(err)
	}

	npc, err := loadPeerCache(f.Name())
	if err != nil {
		t.Fatalf("unable to load cache: %v", err)
	}

	if !testEqualEndpointList(pc.Peers, npc.Peers) {
		t.Errorf("endpoint lists not identical")
	}

	if _, ok := npc.Bans["ban null"]; ok {
		t.Errorf("null time ban was loaded")
	}
	if _, ok := npc.Bans["ban outdated"]; ok {
		t.Errorf("outdated ban was loaded")
	}
	if _, ok := npc.Bans["ban ok"]; !ok || !pc.Bans["ban ok"].Equal(npc.Bans["ban ok"]) {
		t.Errorf("ok ban was not presented or mismatched. got = %v, want = %v", npc.Bans["ban ok"], pc.Bans["ban ok"])
	}
}
