package anchor_test

import (
	. "github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/primitives"
	"testing"
)

func TestMarshalUnmarshalAnchorRecord(t *testing.T) {
	record := `{"AnchorRecordVer":1,"DBHeight":5,"KeyMR":"980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6","RecordHeight":6,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951","BlockHeight":372576,"BlockHash":"000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d","Offset":1144}}`
	ar, err := UnmarshalAnchorRecord([]byte(record))
	if err != nil {
		t.Error(err)
	}
	data, err := ar.Marshal()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual([]byte(record), data) == false {
		t.Errorf("Anchors are not equal\n%s\nvs\n%s", record, string(data))
	}
}
