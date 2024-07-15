package blockchain

import (
	"math/big"


	"github.com/terium-project/terium/internal"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockStore"

	. "github.com/terium-project/terium/internal/transaction"
)


type BlockchainWriter struct {
	ctx 		*internal.DirCtx
}

func (writer BlockchainWriter) New(ctx *internal.DirCtx) {
	writer.ctx = ctx
}

func (writer BlockchainWriter) AddBlock(block *block.Block, lastHash []byte) {
	

	lastIO := blockStore.IndexIO{}
	lastIO.New(writer.ctx, lastHash)
	lastIO.Read()
	currHeight := lastIO.MetaData().Height
	
	sum := big.Int{}
	sum.Add(&currHeight, big.NewInt(1))
	
	metadata := blockStore.BlockMetaData{
		Hash: block.Hash(),
		Height: sum,
		Nonce: block.Header.Nonce,
	}
	blockStore := &blockStore.BlockStore{}
	blockStore.New(writer.ctx, metadata.Hash)
	blockStore.Write(block, &metadata)
	blockStore.Close()
	
	lastIO.New(writer.ctx, metadata.Hash)
	lastIO.WriteLastHash()
	lastIO.Close()
	
}



func (writer BlockchainWriter) DeleteBlock(hash []byte) {
	blockStore := &blockStore.BlockStore{}
	blockStore.New(writer.ctx, hash)
	blockStore.Delete()
	blockStore.Close()
}


type BlockchainIterator struct {
	ctx *internal.DirCtx
	block *block.Block
	store *blockStore.BlockStore
	metadata *blockStore.BlockMetaData
}


func (iter BlockchainIterator) New(ctx *internal.DirCtx) {
	indexIO := blockStore.IndexIO{}
	indexIO.New(ctx, nil)
	indexIO.ReadLastHash()
	store := blockStore.BlockStore{}
	meta := indexIO.MetaData()
	store.New(ctx, meta.Hash)
	store.Read()
	iter.ctx = ctx
	iter.metadata = meta
	iter.block = store.Block()
	iter.store = &store
}

func (iter BlockchainIterator) Next() {
	iter.store.New(iter.ctx, iter.block.Header.PrevHash)
	iter.store.Read()
	iter.metadata = iter.store.Metadata()
	iter.block = iter.store.Block()
}

func (iter BlockchainIterator) Close() {
	iter.store.Close()
}

func (iter BlockchainIterator) Block() *block.Block {
	return iter.block
}

func (iter BlockchainIterator) Metadata() *blockStore.BlockMetaData {
	return iter.metadata
}



type Blockchain struct {
	writer *BlockchainWriter
	ctx *internal.DirCtx
	lastHash []byte
}

func (blockchain Blockchain) New(ctx *internal.DirCtx) {
	blockchain.ctx = ctx
	
	blockchain.writer = &BlockchainWriter{}
	blockchain.writer.New(ctx)

	reader := &BlockchainIterator{}
	reader.New(ctx)

	blockchain.lastHash = reader.metadata.Hash
}

func (blockchain Blockchain) AddBlock(block *block.Block) {
	blockchain.writer.AddBlock(block, blockchain.lastHash)
	blockchain.lastHash = block.Hash()
}

func (blockchain Blockchain) Iterator() *BlockchainIterator {
	iter := BlockchainIterator{}
	iter.New(blockchain.ctx)
	return &iter
}

func (blockchain Blockchain) LastHash() []byte {
	return blockchain.lastHash
}

func (blockchain Blockchain) FindTxOutByOutPt(tx OutPoint) (TxOut, error) {
	var txout TxOut

	return txout, nil

}
