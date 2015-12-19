package main

import (
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestTest(t *testing.T) {
	t.Error("Test")

	dbo := testHelper.CreateAndPopulateTestDatabaseOverlay()

	exportDChain(dbo)
	exportECChain(dbo)
	exportAChain(dbo)
	exportFctChain(dbo)
	exportDirBlockInfo(dbo)

	exportEChain(testHelper.GetChainID().String(), dbo)
	exportEChain(testHelper.GetAnchorChainID().String(), dbo)
}
