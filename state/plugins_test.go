// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/testHelper"
)

func TestSetAndGetUseTorrent(t *testing.T) {
	state := testHelper.CreateAndPopulateTestStateAndStartValidator()
	if state.UsingTorrent() {
		t.Error("State unexpectedly using torrents without having been set to true")
	}

	state.SetUseTorrent(true)
	if !state.UsingTorrent() {
		t.Error("State not using torrents despite having been set to true")
	}
}
func TestSetAndGetTorrentUploader(t *testing.T) {
	state := testHelper.CreateAndPopulateTestStateAndStartValidator()
	if state.TorrentUploader() {
		t.Error("State unexpectedly using TorrentUploader without having been set to true")
	}

	state.SetTorrentUploader(true)
	if !state.TorrentUploader() {
		t.Error("State not using TorrentUploader despite having been set to true")
	}
}
