package wsapi_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/testHelper"
	. "github.com/FactomProject/factomd/wsapi"
)

func TestHandleV2GetRaw(t *testing.T) {
	type RawData struct {
		Hash1 string
		Hash2 string
		Raw   string
	}

	toTest := []RawData{}
	var err error

	blockSet := testHelper.CreateTestBlockSet(nil)

	aBlock := blockSet.ABlock
	raw := RawData{}
	raw.Hash1 = aBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = aBlock.DatabaseSecondaryIndex().String()
	hex, err := aBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //1

	eBlock := blockSet.EBlock
	raw = RawData{}
	raw.Hash1 = eBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = eBlock.DatabaseSecondaryIndex().String()
	hex, err = eBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //2

	ecBlock := blockSet.ECBlock
	raw = RawData{}
	raw.Hash1 = ecBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	raw.Hash2 = ecBlock.(interfaces.DatabaseBatchable).DatabaseSecondaryIndex().String()
	hex, err = ecBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //3

	fBlock := blockSet.FBlock
	raw = RawData{}
	raw.Hash1 = fBlock.(interfaces.DatabaseBatchable).DatabasePrimaryIndex().String()
	raw.Hash2 = fBlock.(interfaces.DatabaseBatchable).DatabaseSecondaryIndex().String()
	hex, err = fBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //4

	dBlock := blockSet.DBlock
	raw = RawData{}
	raw.Hash1 = dBlock.DatabasePrimaryIndex().String()
	raw.Hash2 = dBlock.DatabaseSecondaryIndex().String()
	hex, err = dBlock.MarshalBinary()
	if err != nil {
		panic(err)
	}
	raw.Raw = primitives.EncodeBinary(hex)
	toTest = append(toTest, raw) //5

	//initializing server
	state := testHelper.CreateAndPopulateTestState()
	Start(state)

	for i, v := range toTest {
		data := new(HashRequest)
		data.Hash = v.Hash1
		req := primitives.NewJSON2Request("raw-data", 1, data)

		resp, err := v2Request(req)
		if err != nil {
			t.Errorf("%v", err)
		}

		if strings.Contains(resp.String(), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash1)
			t.Errorf("GetRaw %v/%v from Hash1 failed - %v", i, len(toTest), resp.String())
		}

		data.Hash = v.Hash2
		req = primitives.NewJSON2Request("raw-data", 1, data)
		resp, err = v2Request(req)
		if err != nil {
			t.Errorf("%v", err)
		}

		if strings.Contains(resp.String(), v.Raw) == false {
			t.Errorf("Looking for %v", v.Hash1)
			t.Errorf("GetRaw %v/%v from Hash2 failed - %v", i, len(toTest), resp.String())
		}
	}
}

func v2Request(req *primitives.JSON2Request) (*primitives.JSON2Response, error) {
	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		"http://localhost:8088/v2",
		"application/json",
		bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := primitives.NewJSON2Response()
	if err := json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}
