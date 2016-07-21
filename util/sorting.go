package util

import (
	//"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	. "github.com/FactomProject/factomd/common/interfaces"

	"bytes"
)

//------------------------------------------------
// DBlock array sorting implementation - accending
type ByDBlockIDAccending []IDirectoryBlock

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
type ByECBlockIDAccending []IEntryCreditBlock

func (f ByECBlockIDAccending) Len() int {
	return len(f)
}
func (f ByECBlockIDAccending) Less(i, j int) bool {
	return f[i].GetHeader().GetDBHeight() < f[j].GetHeader().GetDBHeight()
}
func (f ByECBlockIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// ABlock array sorting implementation - accending
type ByABlockIDAccending []IAdminBlock

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
type ByFBlockIDAccending []IFBlock

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
type ByEBlockIDAccending []IEntryBlock

func (f ByEBlockIDAccending) Len() int {
	return len(f)
}
func (f ByEBlockIDAccending) Less(i, j int) bool {
	return f[i].GetHeader().GetEBSequence() < f[j].GetHeader().GetEBSequence()
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

//------------------------------------------------
// DirBlock Info array sorting implementation - accending
type ByDirBlockInfoIDAccending []IDirBlockInfo

func (f ByDirBlockInfoIDAccending) Len() int {
	return len(f)
}
func (f ByDirBlockInfoIDAccending) Less(i, j int) bool {
	return f[i].GetDBHeight() < f[j].GetDBHeight()
}
func (f ByDirBlockInfoIDAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

// ByDirBlockInfoTimestamp defines the methods needed to satisify sort.Interface to
// sort a slice of DirBlockInfo by their Timestamp.
type ByDirBlockInfoTimestamp []IDirBlockInfo

func (u ByDirBlockInfoTimestamp) Len() int { return len(u) }
func (u ByDirBlockInfoTimestamp) Less(i, j int) bool {
	if u[i].GetTimestamp() == u[j].GetTimestamp() {
		return u[i].GetDBHeight() < u[j].GetDBHeight()
	}
	return u[i].GetTimestamp().GetTimeMilli() < u[j].GetTimestamp().GetTimeMilli()
}
func (u ByDirBlockInfoTimestamp) Swap(i, j int) { u[i], u[j] = u[j], u[i] }
