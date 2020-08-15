package util

import (
	//"github.com/PaulSnow/factom2d/common/directoryBlock/dbInfo"
	. "github.com/PaulSnow/factom2d/common/interfaces"

	"bytes"
)

//------------------------------------------------
// DBlock array sorting implementation - ascending
type ByDBlockIDAscending []IDirectoryBlock

func (f ByDBlockIDAscending) Len() int {
	return len(f)
}
func (f ByDBlockIDAscending) Less(i, j int) bool {
	return f[i].GetHeader().GetDBHeight() < f[j].GetHeader().GetDBHeight()
}
func (f ByDBlockIDAscending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// CBlock array sorting implementation - ascending
type ByECBlockIDAscending []IEntryCreditBlock

func (f ByECBlockIDAscending) Len() int {
	return len(f)
}
func (f ByECBlockIDAscending) Less(i, j int) bool {
	return f[i].GetHeader().GetDBHeight() < f[j].GetHeader().GetDBHeight()
}
func (f ByECBlockIDAscending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// ABlock array sorting implementation - ascending
type ByABlockIDAscending []IAdminBlock

func (f ByABlockIDAscending) Len() int {
	return len(f)
}
func (f ByABlockIDAscending) Less(i, j int) bool {
	return f[i].GetHeader().GetDBHeight() < f[j].GetHeader().GetDBHeight()
}
func (f ByABlockIDAscending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// ABlock array sorting implementation - ascending
type ByFBlockIDAscending []IFBlock

func (f ByFBlockIDAscending) Len() int {
	return len(f)
}
func (f ByFBlockIDAscending) Less(i, j int) bool {
	return f[i].GetDBHeight() < f[j].GetDBHeight()
}
func (f ByFBlockIDAscending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

//------------------------------------------------
// EBlock array sorting implementation - ascending
type ByEBlockIDAscending []IEntryBlock

func (f ByEBlockIDAscending) Len() int {
	return len(f)
}
func (f ByEBlockIDAscending) Less(i, j int) bool {
	return f[i].GetHeader().GetEBSequence() < f[j].GetHeader().GetEBSequence()
}
func (f ByEBlockIDAscending) Swap(i, j int) {
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
// DirBlock Info array sorting implementation - ascending
type ByDirBlockInfoIDAscending []IDirBlockInfo

func (f ByDirBlockInfoIDAscending) Len() int {
	return len(f)
}
func (f ByDirBlockInfoIDAscending) Less(i, j int) bool {
	return f[i].GetDBHeight() < f[j].GetDBHeight()
}
func (f ByDirBlockInfoIDAscending) Swap(i, j int) {
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
