package blockStore

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
)

type BlockStore struct {
	blockIO *BlockIO
	indexIO *IndexIO
}

func NewBlockStore(ctx *t_config.Context) *BlockStore {
	store := new(BlockStore)
	store.blockIO = NewBlockIO(ctx)
	store.indexIO = NewIndexIO(ctx)
	return store
}

func (store *BlockStore) Write(block *block.Block, metadata *BlockMetaData) {
	store.blockIO.Write(block, metadata.Hash)
	store.indexIO.Write(metadata)
}

func (store *BlockStore) Read(hash []byte) (*block.Block, *BlockMetaData) {
	block := store.blockIO.Read(hash)
	meta := store.indexIO.Read(hash)
	return block, &BlockMetaData{Hash: hash, Nonce: meta.Nonce, Height: meta.Height}
}

type ERR_NO_BLOCKS_REMAINING struct {}
func (err ERR_NO_BLOCKS_REMAINING) Error() string {
	return "Blockchain is empty"
}
var ErrNoBlocksRemaining ERR_NO_BLOCKS_REMAINING = ERR_NO_BLOCKS_REMAINING{}

func (store *BlockStore) ReadLast() (*block.Block, *BlockMetaData, error) {
	meta, err := store.indexIO.ReadLast()
	if err == badger.ErrKeyNotFound {
		return nil, nil, ErrNoBlocksRemaining
	} else if err != nil {
		t_error.LogErr(err)
	}
	block := store.blockIO.Read(meta.Hash)
	return block, meta, nil
}

func (store *BlockStore) Delete(hash []byte) {
	store.blockIO.Delete(hash)
	store.indexIO.Delete(hash)
}

func (store *BlockStore) Close() {
	store.indexIO.Close()
}
