package blockStore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"os"
	"path"

	"github.com/tiereum/trmnode/internal/block"
	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"
)

type CorruptBlockErr struct{}

func (c CorruptBlockErr) Error() string {
	return "Block is corrupt."
}

type BlockMetaData struct {
	Hash   []byte
	Nonce  uint32
	Height big.Int
}

type BlockIO struct {
	ctx *t_config.Context
}

func NewBlockIO(ctx *t_config.Context) *BlockIO {
	return &BlockIO{
		ctx: ctx,
	}
}

func (b *BlockIO) Write(block *block.Block, hash []byte) error {

	f, err := os.OpenFile(path.Join(b.ctx.DataDir, hex.EncodeToString(hash)), os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return err
	}
	_bytes := block.Serialize()
	n, err := f.Write(_bytes)
	t_error.LogErr(err)

	sum := b.Checksum(_bytes)
	f.WriteAt(sum, int64(n))
	return nil
}

func (b *BlockIO) Read(hash []byte) *block.Block {

	allBytes, err := os.ReadFile(path.Join(b.ctx.DataDir, hex.EncodeToString(hash)))
	t_error.LogErr(err)
	dataBytes := allBytes[:len(allBytes)-32]
	sum := allBytes[len(allBytes)-32:]

	valid := b.Check(dataBytes, sum)

	if !valid {
		t_error.LogErr(CorruptBlockErr{})
	}

	blockDecoder := block.NewBlockDecoder(nil)
	buffer := bytes.Buffer{}
	buffer.Write(dataBytes)
	blockDecoder.Decode(&buffer)

	return blockDecoder.Out()
}

func (b *BlockIO) Checksum(blockBytes []byte) []byte {
	sum := sha256.Sum256(blockBytes)
	return sum[:]
}

func (b *BlockIO) Check(_bytes, sum []byte) bool {
	expected := sha256.Sum256(_bytes)

	return bytes.Equal(expected[:], sum)

}

func (b *BlockIO) Update(block *block.Block, hash []byte) {
	b.Delete(hash)
	b.Write(block, hash)
}

func (b *BlockIO) Delete(hash []byte) {
	path := path.Join(b.ctx.DataDir, hex.EncodeToString(hash))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t_error.LogWarn(errors.New("file does not exist, cant delete"))
		return
	}
	t_error.LogErr(os.Remove(path))
}
