package blockExtractor_test

import (
	. "github.com/FactomProject/factomd/database/blockExtractor"
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestTest(t *testing.T) {
	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	err := ExportDChain(dbo)
	if err != nil {
		t.Error(err)
	}
	err = ExportECChain(dbo)
	if err != nil {
		t.Error(err)
	}
	err = ExportAChain(dbo)
	if err != nil {
		t.Error(err)
	}
	err = ExportFctChain(dbo)
	if err != nil {
		t.Error(err)
	}
	err = ExportDirBlockInfo(dbo)

	if err != nil {
		t.Error(err)
	}
	err = ExportEChain(testHelper.GetChainID().String(), dbo)
	if err != nil {
		t.Error(err)
	}
	err = ExportEChain(testHelper.GetAnchorChainID().String(), dbo)
	if err != nil {
		t.Error(err)
	}
}
