package blockStore

import (
	"crypto/sha256"
	"math/big"
	"os"
	"path"

	"github.com/terium-project/terium/internal"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/t_error"
)

type CorruptBlockErr struct {}

func (c CorruptBlockErr) Error() string {
	return "Block is corrupt."
}

type BlockMetaData struct {
	Hash []byte
	Nonce uint32
}

type BlockIO struct {
	dirs *internal.DirCtx
	block *block.Block
	bytes []byte
	sum []byte
	metadata *BlockMetaData
	path string
	file *os.File
}

func (b BlockIO) New(meta *BlockMetaData) {
	b.metadata = meta
	b.path = path.Join(b.dirs.DataDir, string(b.metadata.Hash))
	b.dirs = &internal.T_DirCtx
}

func (b BlockIO) Write(block *block.Block) error {
	b.block = block

	f, err := os.OpenFile(b.path, os.O_CREATE | os.O_RDWR, 0644)

	if err != nil {
		return err
	}
	b.bytes = b.block.Serialize()
	n, err := f.Write(b.bytes)
	if err != nil {
		t_error.LogErr(err)
	}

	b.Checksum()
	f.WriteAt(b.sum, int64(n))
	return nil
}

func (b BlockIO) Read(hash []byte) error {

	_bytes, err := os.ReadFile(b.path)
	if err != nil {
		t_error.LogErr(err)
	}
	b.bytes = _bytes[:len(_bytes) - 32]
	b.sum = _bytes[len(_bytes) - 32:]

	valid := b.Check()

	if !valid {
		return CorruptBlockErr{}
	}

	return nil
}

func (b BlockIO) Checksum() {
	sum := sha256.Sum256(b.bytes)
	b.sum = sum[:]
}

func (b BlockIO) Check() bool {
	expected := sha256.Sum256(b.bytes)
	expectedSum := big.Int{}
	expectedSum.SetBytes(expected[:])

	actualSum := big.Int{}
	actualSum.SetBytes(b.sum)

	return expectedSum.Cmp(&actualSum) == 0

}

func (b BlockIO) Update() error {

	// not sure yet if Update() should be separate from Write()
	return nil
}

func (b BlockIO) Delete() error {
	os.Remove(b.path)
	return nil
}