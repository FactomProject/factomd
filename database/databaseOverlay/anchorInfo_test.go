// +build all 

// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package databaseOverlay_test

import (
	"testing"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/testHelper"
)

func TestRebuildDirBlockInfo(t *testing.T) {
	dbo := testHelper.CreateEmptyTestDatabaseOverlay()
	anchors, err := CreateAnchors()
	if err != nil {
		t.Fatalf("Failed to create test anchors: %s", err.Error())
	}

	for _, anchor := range anchors {
		dbi, err := dbo.CreateUpdatedDirBlockInfoFromAnchorRecord(anchor)
		if err != nil {
			t.Error(err)
		}

		err = dbo.ProcessDirBlockInfoBatch(dbi)
		if err != nil {
			t.Error(err)
		}
	}
}

func CreateAnchors() ([]*anchor.AnchorRecord, error) {
	anchors := []*anchor.AnchorRecord{}

	records := []string{
		`{"AnchorRecordVer": 1, "DBHeight": 8, "KeyMR": "637b6010cb6121f76c65b200a6cf94cb6655881fb4cac48979f8950e7a349da1","RecordHeight": 9,
	"Bitcoin": {
		"Address": "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF",
		"TXID": "b73b38b8af43f4dbaeb061f158d4bf5004b40216b30acd3beca43fae1ba6d1b7",
		"BlockHeight": 372579,
		"BlockHash": "00000000000000000589540fdaacf4f6ba37513aedc1033e68a649ffde0573ad",
		"Offset": 1185
	}}`,
		`{"AnchorRecordVer": 1, "DBHeight": 12, "KeyMR": "6b4ef43604d2d5fb14267411fa0d1fa6ea7cb5fce631dfbe619334f082bc504f", "RecordHeight": 12,
	"Bitcoin": {
		"Address": "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF",
		"TXID": "e0c67e64cfdbf025b41eb235cc07b2e63fb3c62c8b877f319cac8ec3b5483223",
		"BlockHeight": 372584,
		"BlockHash": "00000000000000000fdc6526a60522d44731c4d30b36421c10bb21fbe97eb468",
		"Offset": 536
	}}`,
		`{"AnchorRecordVer": 1, "DBHeight": 14, "KeyMR": "26d5be575d93b4fc2e15294266ab423fff8cb442a3c8111be99e6a5e7c3682d2", "RecordHeight": 16,
	"Bitcoin": {
		"Address": "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF",
		"TXID": "b8a9104f58ae697df3c6052f01243809222d4f17c1eb3983adf77aea60b7b17b",
		"BlockHeight": 372587,
		"BlockHash": "0000000000000000056b2c07b093727ae1d49186ef93857e5bc8ba97cac52790",
		"Offset": 1368
	}}`,
		`{"AnchorRecordVer": 1, "DBHeight": 21, "KeyMR": "b2e3b5dd50a0bbd6837e2a4cd6fa6b83a8ac2640f0caf2109200e53a7497d587", "RecordHeight": 21,
	"Bitcoin": {
		"Address": "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF",
		"TXID": "43faa9e0b4f8b3fd366bf1a3a4fe6d42276627f96bcc4376754130fc4c3faf63",
		"BlockHeight": 372594,
		"BlockHash": "000000000000000008101836aa63e20b5cd2b3e4bd4133cb990306c4fd2c4f60",
		"Offset": 415
	}}`,
		`{"AnchorRecordVer": 1, "DBHeight": 22,	"KeyMR": "5c9f7cf6667b6d46da730cddd13a7c0094ef39462a7db4e779f950e5a7c763cc", "RecordHeight": 24,
	"Bitcoin": {
		"Address": "1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF",
		"TXID": "5e77a98390e45f39bd38e908944858fa3e85ea47a06be5480af68fa256213572",
		"BlockHeight": 372595,
		"BlockHash": "0000000000000000007f3d7a17a7565d9b326ad7a692710a71ed070394399a33",
		"Offset": 1017
	}}`,
	}

	for _, record := range records {
		ar, err := anchor.UnmarshalAnchorRecord([]byte(record))
		if err != nil {
			return nil, err
		}
		anchors = append(anchors, ar)
	}

	return anchors, nil
}
