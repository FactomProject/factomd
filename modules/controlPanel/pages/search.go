package pages

type Search struct {
	Title          string
	Content        string
	Term           string
	DirectoryBlock DirectoryBlockInfo
}

type DirectoryBlockInfo struct {
	KeyMerkleRoot                    string
	BodyKeyMerkleRoot                string
	Hash                             string
	TimeStamp                        string
	BlockHeight                      string
	PreviousDirectoryBlockMerkleRoot string
	PreviousDirectoryBlockHash       string
}
