package anchor_test

import (
	"encoding/hex"
	"testing"

	. "github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
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

func TestMarshalUnmarshalAnchorRecordV2(t *testing.T) {
	record := `{"AnchorRecordVer":1,"DBHeight":5,"KeyMR":"980ab6d50d9fad574ad4df6dba06a8c02b1c67288ee5beab3fbfde2723f73ef6","RecordHeight":6,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"e2ac71c9c0fd8edc0be8c0ba7098b77fb7d90dcca755d5b9348116f3f9d9f951","BlockHeight":372576,"BlockHash":"000000000000000003059382ed4dd82b2086e99ec78d1b6e811ebb9d53d8656d","Offset":1144}}`
	ar, err := UnmarshalAnchorRecord([]byte(record))
	if err != nil {
		t.Error(err)
	}

	priv := "ec9f1cefa00406b80d46135a53504f1f4182d4c0f3fed6cca9281bc020eff973"
	pk1, err := primitives.NewPrivateKeyFromHex(priv)
	if err != nil {
		t.Error(err)
	}

	data, sig, err := ar.MarshalAndSignV2(pk1)
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual([]byte(record), data) == false {
		t.Errorf("Anchors are not equal\n%s\nvs\n%s", record, string(data))
	}

	ar, valid, err := UnmarshalAndValidateAnchorRecordV2(data, [][]byte{sig}, []interfaces.Verifier{pk1.Pub})
	if err != nil {
		t.Error(err)
	}
	if valid != true {
		t.Errorf("Anchor record is not valid")
	}
	if ar == nil {
		t.Errorf("No anchor record returned!")
	}
}

func TestValidateAnchorRecord(t *testing.T) {
	pub := new(primitives.PublicKey)
	err := pub.UnmarshalText([]byte("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"))
	if err != nil {
		t.Error(err)
	}
	signedRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0807`

	ar, valid, err := UnmarshalAndValidateAnchorRecord([]byte(signedRecord), []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != true {
		t.Errorf("Anchor record is not valid")
	}
	if ar == nil {
		t.Errorf("No anchor record returned!")
	}

	invalidRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0806`

	ar, valid, err = UnmarshalAndValidateAnchorRecord([]byte(invalidRecord), []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != false {
		t.Errorf("Anchor record is valid when it shouldn't be")
	}
	if ar != nil {
		t.Errorf("Anchor record returned when it shouldn't be!")
	}
}

func TestValidateAnchorRecordV2(t *testing.T) {
	pub := new(primitives.PublicKey)
	err := pub.UnmarshalText([]byte("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a"))
	if err != nil {
		t.Error(err)
	}
	signedRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}`
	sig, err := hex.DecodeString("e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0807")
	if err != nil {
		t.Error(err)
	}

	ar, valid, err := UnmarshalAndValidateAnchorRecordV2([]byte(signedRecord), [][]byte{sig}, []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != true {
		t.Errorf("Anchor record is not valid")
	}
	if ar == nil {
		t.Errorf("No anchor record returned!")
	}

	invalidRecord := `{"AnchorRecordVer":1,"DBHeight":46226,"KeyMR":"831287d96f9955ff438836b26219b8c937e01e668295e8334a5d140cdaacd2fd","RecordHeight":46226,"Bitcoin":{"Address":"1K2SXgApmo9uZoyahvsbSanpVWbzZWVVMF","TXID":"9a47f131677e135645cebe74085cd980044eda2abbb05ac09ee79ba8640f368d","BlockHeight":421417,"BlockHash":"000000000000000003781bf1b36dbaa1a71f94e23b4b9dc13024379aaf244d89","Offset":599}}`
	invalidSig, err := hex.DecodeString("e5d40b8673fbeba9e4e5a4322a514ec28989acef55b6fcdf49eae400b6d302852c2cf6659dc1f3298ffae60f6605666efbcef3ac4947333515084e1269ec0806")

	ar, valid, err = UnmarshalAndValidateAnchorRecordV2([]byte(invalidRecord), [][]byte{invalidSig}, []interfaces.Verifier{pub})
	if err != nil {
		t.Error(err)
	}
	if valid != false {
		t.Errorf("Anchor record is valid when it shouldn't be")
	}
	if ar != nil {
		t.Errorf("Anchor record returned when it shouldn't be!")
	}
}
