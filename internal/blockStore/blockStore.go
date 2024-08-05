package blockStore

import (
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/t_config"
)

type BlockStore struct {
	blockIO *BlockIO
	indexIO *IndexIO
}

func NewBlockStore(ctx *t_config.Context, hash []byte) *BlockStore {
	store := BlockStore{}
	store.blockIO = NewBlockIO(ctx, hash)
	store.indexIO = NewIndexIO(ctx, hash)
	return &store
}

func (store *BlockStore) Write(block *block.Block, metadata *BlockMetaData) {
	store.blockIO.Write(block)
	store.indexIO.Create(metadata)
}

func (store *BlockStore) Read() {
	store.blockIO.Read()
	store.indexIO.Read()
}

func (store *BlockStore) Block() *block.Block {
	return store.blockIO.Block()
}

func (store *BlockStore) Metadata() *BlockMetaData {
	return store.indexIO.MetaData()
}

func (store *BlockStore) Update(block *block.Block, metadata *BlockMetaData) {
	store.blockIO.Update(block)
	store.indexIO.Update(metadata)
}

func (store *BlockStore) Delete() {
	store.blockIO.Delete()
	store.indexIO.Delete()
}

func (store *BlockStore) Close() {
	store.indexIO.Close()
}
