package blockStore

import (
	"bytes"
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
	Height big.Int
}


type BlockIO struct {
	ctx *internal.DirCtx
	block *block.Block
	bytes []byte
	sum []byte
	hash []byte
	path string
}

func (b *BlockIO) New(ctx *internal.DirCtx, hash []byte) {
	b.ctx = ctx
	b.hash = hash
	b.path = path.Join(b.ctx.DataDir, string(b.hash))
	b.ctx = &internal.T_DirCtx
}

func (b *BlockIO) Write(block *block.Block) error {



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

func (b *BlockIO) Read() error {

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

	blockDecoder := block.BlockDecoder{}
	buffer := bytes.Buffer{}
	buffer.Write(b.bytes)
	blockDecoder.New(b.block)
	blockDecoder.Decode(buffer)

	return nil
}

func (b *BlockIO) Checksum() {
	sum := sha256.Sum256(b.bytes)
	b.sum = sum[:]
}

func (b *BlockIO) Check() bool {
	expected := sha256.Sum256(b.bytes)
	expectedSum := big.Int{}
	expectedSum.SetBytes(expected[:])

	actualSum := big.Int{}
	actualSum.SetBytes(b.sum)

	return expectedSum.Cmp(&actualSum) == 0

}

func (b *BlockIO) Update(block *block.Block) error {
	b.Write(block)
	return nil
}

func (b *BlockIO) Delete() error {
	os.Remove(b.path)
	return nil
}

func (b *BlockIO) Block() *block.Block {
	return b.block
}