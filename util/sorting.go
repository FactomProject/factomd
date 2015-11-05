package util

import (
	. "github.com/FactomProject/factomd/common/directoryBlock"
	. "github.com/FactomProject/factomd/common/entryBlock"
	. "github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"

	"bytes"
)

//------------------------------------------------
// DBlock array sorting implementation - accending
type ByDBlockIDAccending []*DirectoryBlock

func (f ByDBlockIDAccending) Len() int {
	return len(f)
}
func (f ByDBlockIDAccending) Less(i, j int) bool {
	return f[i].GetHeader().GetDBHeight() < f[j].GetHeader().GetDBHeight()
}
func (f ByDBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// CBlock array sorting implementation - accending
type ByECBlockIDAccending []*ECBlock

func (f ByECBlockIDAccending) Len() int {
	return len(f)
}
func (f ByECBlockIDAccending) Less(i, j int) bool {
	return f[i].Header.DBHeight < f[j].Header.DBHeight
}
func (f ByECBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// ABlock array sorting implementation - accending
type ByABlockIDAccending []interfaces.IAdminBlock

func (f ByABlockIDAccending) Len() int {
	return len(f)
}
func (f ByABlockIDAccending) Less(i, j int) bool {
	return f[i].GetHeader().GetDBHeight() < f[j].GetHeader().GetDBHeight()
}
func (f ByABlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// ABlock array sorting implementation - accending
type ByFBlockIDAccending []interfaces.IFBlock

func (f ByFBlockIDAccending) Len() int {
	return len(f)
}
func (f ByFBlockIDAccending) Less(i, j int) bool {
	return f[i].GetDBHeight() < f[j].GetDBHeight()
}
func (f ByFBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// EBlock array sorting implementation - accending
type ByEBlockIDAccending []*EBlock

func (f ByEBlockIDAccending) Len() int {
	return len(f)
}
func (f ByEBlockIDAccending) Less(i, j int) bool {
	return f[i].Header.EBSequence < f[j].Header.EBSequence
}
func (f ByEBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// Byte array sorting - ascending
type ByByteArray [][]byte

func (f ByByteArray) Len() int {
	return len(f)
}
func (f ByByteArray) Less(i, j int) bool {
	return bytes.Compare(f[i], f[j]) < 0
}
func (f ByByteArray) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
