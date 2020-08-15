package entryBlock_test

import (
	"encoding/hex"
	"testing"

	. "github.com/PaulSnow/factom2d/common/entryBlock"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// TestUnmarshalNilEBlock checks that unmarshalling nil and the empty interface produce errors
func TestUnmarshalNilEBlock(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Panic caught during the test - %v", r)
		}
	}()

	a := new(EBlock)
	err := a.UnmarshalBinary(nil)
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}

	err = a.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("Error is nil when it shouldn't be")
	}
}

// TestEBlockMarshal checks that an EBlock can be marshalled and unmarshalled properly
func TestEBlockMarshal(t *testing.T) {
	eb := newTestingEntryBlock()

	// Marshal
	p, err := eb.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	// Unmarshal
	eb2 := NewEBlock()
	err = eb2.UnmarshalBinary(p)
	if err != nil {
		t.Error(err)
	}

	// Marshal it again
	p2, err := eb2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	// Check the bytes are equal
	if primitives.AreBytesEqual(p, p2) == false {
		t.Logf("eb1 = %x\n", p)
		t.Logf("eb2 = %x\n", p2)
		t.Fail()
	}

	// Unmarshal it one more time
	eb3, err := UnmarshalEBlock(p)
	if err != nil {
		t.Error(err)
	}
	// And remarshal it
	p3, err := eb3.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(p, p3) == false {
		t.Logf("eb1 = %x\n", p)
		t.Logf("eb3 = %x\n", p3)
		t.Fail()
	}
}

// TestEBlockMassiveUnmarshal checks that a large (10k) entry can be marshalled and unmarshalled properly
func TestEBlockMassiveUnmarshal(t *testing.T) {
	eb := newTestingEntryBlock()

	e := primitives.NewZeroHash()
	total := 10000
	entries := make([]interfaces.IHash, total)
	for i := 0; i < total; i++ {
		entries[i] = e
	}
	eb.GetBody().(*EBlockBody).EBEntries = entries
	eb.BuildHeader()

	p, err := eb.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	eb2 := NewEBlock()
	err = eb2.UnmarshalBinary(p)
	if err != nil {
		t.Error(err)
	}

	p2, err := eb2.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	if primitives.AreBytesEqual(p, p2) == false {
		t.Logf("eb1 = %x\n", p)
		t.Logf("eb2 = %x\n", p2)
		t.Fail()
	}
}

// TestAddEBEntry checks that you can marshal and unmarshal a test object AFTER adding an entry block
func TestAddEBEntry(t *testing.T) {
	eb := newTestingEntryBlock()
	e := newEntry()
	if err := eb.AddEBEntry(e); err != nil {
		t.Error(err)
	}

	p, err := eb.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	eb2 := NewEBlock()
	if err := eb2.UnmarshalBinary(p); err != nil {
		t.Error(err)
	}
}

// byteof returns a filled 32 byte array with the input byte copied 32 times
func byteof(b byte) []byte {
	r := make([]byte, 0, 32)
	for i := 0; i < 32; i++ {
		r = append(r, b)
	}
	return r
}

//TestEntryBlockMisc checks miscellaineous functions of the entry block
func TestEntryBlockMisc(t *testing.T) {
	e := newEntryBlock()
	hash, err := e.Hash()

	if err != nil {
		t.Error(err)
	}
	if hash.String() != "1ec4c9a52ede96e57f855efc8cb1475e4a449773bad7a5b9a8b9abf4c683a1da" {
		t.Errorf("Returned wrong Hash")
	}
	hash, err = e.KeyMR()

	if err != nil {
		t.Error(err)
	}
	if hash.String() != "a9fc0b656430d8bf71d180760b0b352c08f45a55a8cf157383613484587b4d21" {
		t.Errorf("Returned wrong KeyMR")
	}

	if e.GetEntrySigHashes() != nil {
		t.Errorf("Invalid GetEntrySigHashes")
	}

	if e.GetDatabaseHeight() != 6920 {
		t.Errorf("Invalid GetDatabaseHeight - %v", e.GetDatabaseHeight())
	}
	if e.GetChainID().String() != "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a19746" {
		t.Errorf("Invalid GetChainID - %v", e.GetChainID())
	}
	if e.GetHashOfChainIDHash().String() != "ca8c59c692b660b3e10cc94c7bb1dd893752f496effc867d6f04791a3f364bdd" {
		t.Errorf("Invalid GetHashOfChainIDHash - %v", e.GetHashOfChainIDHash())
	}
	if e.DatabasePrimaryIndex().String() != "a9fc0b656430d8bf71d180760b0b352c08f45a55a8cf157383613484587b4d21" {
		t.Errorf("Invalid DatabasePrimaryIndex - %v", e.DatabasePrimaryIndex())
	}
	if e.DatabaseSecondaryIndex().String() != "1ec4c9a52ede96e57f855efc8cb1475e4a449773bad7a5b9a8b9abf4c683a1da" {
		t.Errorf("Invalid DatabaseSecondaryIndex - %v", e.DatabaseSecondaryIndex())
	}
	if e.GetHash().String() != "1ec4c9a52ede96e57f855efc8cb1475e4a449773bad7a5b9a8b9abf4c683a1da" {
		t.Errorf("Invalid GetHash - %v", e.GetHash())
	}
	if e.BodyKeyMR().String() != "25f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21" {
		t.Errorf("Invalid BodyKeyMR - %v", e.BodyKeyMR())
	}
}

// newTestingEntryBlock creates a new entry block for testing
func newTestingEntryBlock() *EBlock {
	// build an EBlock for testing
	eb := NewEBlock()
	hash := primitives.NewZeroHash()
	hash.SetBytes(byteof(0x11))
	eb.Header.SetChainID(hash)

	hash = primitives.NewZeroHash()
	hash.SetBytes(byteof(0x22))
	eb.Header.SetBodyMR(hash)

	hash = primitives.NewZeroHash()
	hash.SetBytes(byteof(0x33))
	eb.Header.SetPrevKeyMR(hash)

	hash = primitives.NewZeroHash()
	hash.SetBytes(byteof(0x44))
	eb.Header.SetPrevFullHash(hash)

	eb.Header.SetEBSequence(5)
	eb.Header.SetDBHeight(6)
	eb.Header.SetEntryCount(7)
	ha := primitives.NewZeroHash()
	ha.SetBytes(byteof(0xaa))
	hb := primitives.NewZeroHash()
	hb.SetBytes(byteof(0xbb))
	eb.Body.EBEntries = append(eb.Body.EBEntries, ha)
	eb.AddEndOfMinuteMarker(0xcc)
	eb.Body.EBEntries = append(eb.Body.EBEntries, hb)

	return eb
}

// newEntryBlock creates a new entry block for use in tests within this file
func newEntryBlock() *EBlock {
	e := NewEBlock()
	entryStr := "4bf71c177e71504032ab84023d8afc16e302de970e6be110dac20adbf9a1974625f25d9375533b44505964af993212ef7c13314736b2c76a37c73571d89d8b21c6180f7430677d46d93a3e17b68e6a25dc89ecc092cee1459101578859f7f6969d171a092a1d04f067d55628b461c6a106b76b4bc860445f87b0052cdc5f2bfd000002d800001b080000000272d72e71fdee4984ecb30eedcc89cb171d1f5f02bf9a8f10a8b2cfbaf03efe1c0000000000000000000000000000000000000000000000000000000000000001"
	h, err := hex.DecodeString(entryStr)
	if err != nil {
		panic(err)
	}
	err = e.UnmarshalBinary(h)
	if err != nil {
		panic(err)
	}
	return e
}

// TestSameAs checks that the IsSameAs function of two different blocks do not compare the same, and that the same block matches its clone
func TestSameAs(t *testing.T) {
	//block 1000
	eblock1kbytes, _ := hex.DecodeString("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e6041611c693d62887530c5420a48f2ea2d6038745fc493d6b1e531232805dd2149614ef537df0c73df748b12d508b4334fe8d2832a4cd6ea24f64a3363839bd0efa46e835bfed10ded0d756d7ccafd44830cc942799fca43f2505e9d024b0a9dd3c00000221000003e800000002b24d4ee9e2184673a4d7de6fdac1288ea00b7856940341122c34bd50a662340a0000000000000000000000000000000000000000000000000000000000000009")
	eb1k, err := UnmarshalEBlock(eblock1kbytes)
	if err != nil {
		t.Errorf("eb unmarshall failed")
	}

	eblock1k1bytes, _ := hex.DecodeString("df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604f1a3fee8d61b407ae8c43f0ef91c5bc21f9faf560ccc6167f0ae31d13c8e4628f08c42bc44c09ac26c349bef8ee80d2ffb018cfa3e769107b2413792fa9bd64200f9ce481c4e389a83461f5ebff43e10cad5d55e15d58c3afd4fc16006b9519500000222000003e900000002f3dffaa3e03b5520e5876ff1efedf16eacd69620e2d37344c73fd72e478464df0000000000000000000000000000000000000000000000000000000000000004")
	eb1k1, err := UnmarshalEBlock(eblock1k1bytes)
	if err != nil {
		t.Errorf("eb unmarshall failed")
	}

	if eb1k.IsSameAs(eb1k1) {
		t.Errorf("entry block same as comparison failed")
	}

	eb1kclone, err := UnmarshalEBlock(eblock1kbytes)
	if err != nil {
		t.Errorf("eb unmarshall failed")
	}

	if !eb1k.IsSameAs(eb1kclone) {
		t.Errorf("entry block same as comparison failed")
	}
	//fmt.Println(db1k1)
}

// TestEBlock is a helper structure for testing
type TestEBlock struct {
	Raw   string // Raw string of the block
	KeyMR string // Merkle Root
	Hash  string // Hash of the block
}

// TestMarshalUnmarshalStaticEBlock checks an array of two test block strings can be decoded, unmarshalled, and re-marshalled to get the same
// string, and its appropriate hashes
func TestMarshalUnmarshalStaticEBlock(t *testing.T) {
	ts := []TestEBlock{}

	t1 := TestEBlock{}
	t1.Raw = "06a40590f536293bdecc3d7e69a5c21785c6ed454a59caf7b2e083a1a88ac85b900986feee6603c74fc3aa925d3de2371190fee36632bd2f1389cbaa6d62a98b78fe8619ef8af7ddae3de88a5063476d04467a576c4ff40abd157429e19f1748f43a6d9767a02dbbfbec86fe962552f22c55d4f6467dc8cf019d1d0ba06a88d40000f77b0001602100000002370054e235209bda68f934c8d4bd9d84edef4c03ebd890fe7686edca138f61880000000000000000000000000000000000000000000000000000000000000008"
	t1.KeyMR = "1462592f58712147b62617c6fb37380a223cd32ef673345340e94521df3c9aca"
	t1.Hash = "52794022b3da85b58df69cb842b85d45cda70771677fe0011f5af852eb30e930"
	ts = append(ts, t1)

	t2 := TestEBlock{}
	t2.Raw = "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604831da78a07e01a7495719c54a23861d7d2c25a3eac0b5bfda7d7715c6937c108e9a64a371c68e89f5a3e97f6512c3864fa73448b1907de9078f5a9dc0f4b2d0aecfb6f5aee8e04bde68a5b69a4cea55e8b4e7a3a8efbf75d45a6b686874c5bbe0000000d0000001900000004c480b681b113118876e2540b1f9791af555dc2cd9b5806451305167816281710c92715fe2262b22b1b6fad0f4fc81ff7ccf5f9d633fb6815db414ec023719a72c49b069dbc664c2f247b812160c4a902826483df2b47c27dfd2a95c4281dda790000000000000000000000000000000000000000000000000000000000000003"
	t2.KeyMR = "78ac31584a1e526a3739d6eac5129f6a71aefa722792f9afe8b428f34a9f673c"
	t2.Hash = "8180cef1efb75d39fb581c44688a43cae6a4ebab70d7abbe9f2d8864e230e75c"
	ts = append(ts, t2)

	for _, tBlock := range ts {
		rawStr := tBlock.Raw
		raw, err := hex.DecodeString(rawStr)
		if err != nil {
			t.Errorf("%v", err)
		}

		f := new(EBlock)
		rest, err := f.UnmarshalBinaryData(raw)
		if err != nil {
			t.Errorf("%v", err)
		}
		if len(rest) > 0 {
			t.Errorf("Returned too much data - %x", rest)
		}

		b, err := f.MarshalBinary()
		if err != nil {
			t.Errorf("%v", err)
		}

		if primitives.AreBytesEqual(raw, b) == false {
			t.Errorf("Marshalled bytes are not equal - %x vs %x", raw, b)
		}

		//f.CalculateHashes()

		if f.DatabasePrimaryIndex().String() != tBlock.KeyMR {
			t.Errorf("Wrong PrimaryIndex - %v vs %v", f.DatabasePrimaryIndex().String(), tBlock.KeyMR)
		}
		if f.DatabaseSecondaryIndex().String() != tBlock.Hash {
			t.Errorf("Wrong SecondaryIndex - %v vs %v", f.DatabaseSecondaryIndex().String(), tBlock.Hash)
		}
	}
}
