package blockchain

import (
	"errors"
	"math/big"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockStore"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/utxoSet"
)

type BlockchainWriter struct {
	ctx *t_config.Context
}

func NewBlockchainWriter(ctx *t_config.Context) *BlockchainWriter {

	return &BlockchainWriter{ctx: ctx}
}

func (writer *BlockchainWriter) AddBlock(block *block.Block, lastHash []byte) {
	lastIO := blockStore.NewIndexIO(writer.ctx, lastHash)
	defer lastIO.Close()
	lastIO.Read()
	currHeight := lastIO.MetaData().Height

	sum := big.Int{}
	sum.Add(&currHeight, big.NewInt(1))

	metadata := blockStore.BlockMetaData{
		Hash:   block.Hash(),
		Height: sum,
		Nonce:  block.Header.Nonce,
	}
	_blockStore := blockStore.NewBlockStore(writer.ctx, metadata.Hash)
	defer _blockStore.Close()
	_blockStore.Write(block, &metadata)

	lastIO = blockStore.NewIndexIO(writer.ctx, metadata.Hash)
	lastIO.WriteLastHash()

	txIO := transaction.NewTxIndexIO(writer.ctx)
	defer txIO.Close()

	for i, tx := range block.Transactions {
		txIO.Create(tx.Hash(), &transaction.TxMetadata{
			BlockHash:   metadata.Hash,
			BlockHeight: metadata.Height,
			Index:       uint8(i),
		})
	}
}

func (writer *BlockchainWriter) DeleteBlock(hash []byte) {
	blockStore := blockStore.NewBlockStore(writer.ctx, hash)
	defer blockStore.Close()
	blockStore.Delete()
}

type BlockchainIterator struct {
	ctx      *t_config.Context
	block    *block.Block
	store    *blockStore.BlockStore
	metadata *blockStore.BlockMetaData
}

func NewBlockchainIterator(ctx *t_config.Context) *BlockchainIterator {
	iter := BlockchainIterator{}
	indexIO := blockStore.NewIndexIO(ctx, nil)
	defer indexIO.Close()
	indexIO.ReadLastHash()
	meta := indexIO.MetaData()
	store := blockStore.NewBlockStore(ctx, meta.Hash)
	defer store.Close()
	store.Read()
	iter.ctx = ctx
	iter.metadata = meta
	iter.block = store.Block()
	iter.store = store
	return &iter
}

func (iter *BlockchainIterator) Next() {
	iter.store = blockStore.NewBlockStore(iter.ctx, iter.block.Header.PrevHash)
	iter.store.Read()
	iter.metadata = iter.store.Metadata()
	iter.block = iter.store.Block()
}

func (iter *BlockchainIterator) Close() {
	iter.store.Close()
}

func (iter *BlockchainIterator) Block() *block.Block {
	return iter.block
}

func (iter *BlockchainIterator) Metadata() *blockStore.BlockMetaData {
	return iter.metadata
}

type Blockchain struct {
	writer   *BlockchainWriter
	ctx      *t_config.Context
	lastHash []byte
}

func NewBlockchain(ctx *t_config.Context) *Blockchain {

	blockchain := Blockchain{}
	blockchain.ctx = ctx

	blockchain.writer = NewBlockchainWriter(ctx)
	reader := NewBlockchainIterator(ctx)

	blockchain.lastHash = reader.metadata.Hash
	return &blockchain
}

func (blockchain *Blockchain) AddBlock(block *block.Block) {
	blockchain.writer.AddBlock(block, blockchain.lastHash)
	blockchain.lastHash = block.Hash()
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	iter := NewBlockchainIterator(blockchain.ctx)
	return iter
}

func (blockchain *Blockchain) LastHash() []byte {
	return blockchain.lastHash
}

func (blockchain *Blockchain) Height() big.Int {
	return blockchain.LastBlock().Height
}

func (blockchain *Blockchain) BlockMeta(hash []byte) *blockStore.BlockMetaData {
	store := blockStore.NewBlockStore(blockchain.ctx, hash)
	defer store.Close()
	store.Read()
	return store.Metadata()
}

func (blockchain *Blockchain) Block(hash []byte) *block.Block {
	store := blockStore.NewBlockStore(blockchain.ctx, hash)
	defer store.Close()
	store.Read()
	return store.Block()
}

func (blockchain *Blockchain) TxMeta(hash []byte) *transaction.TxMetadata {
	store := transaction.NewTxIndexIO(blockchain.ctx)
	defer store.Close()
	meta := store.Read(hash)
	return meta
}

func (blockchain *Blockchain) Tx(hash []byte) *transaction.Tx {
	meta := blockchain.TxMeta(hash)
	store := blockStore.NewBlockStore(blockchain.ctx, meta.BlockHash)
	defer store.Close()
	store.Read()
	return &store.Block().Transactions[meta.Index]
}

func (blockchain *Blockchain) LastBlock() *blockStore.BlockMetaData {
	lastIO := blockStore.NewIndexIO(blockchain.ctx, blockchain.lastHash)
	defer lastIO.Close()
	lastIO.Read()
	return lastIO.MetaData()
}

func (blockchain *Blockchain) FindUTXO(outpt *transaction.OutPoint) (*transaction.Utxo, error) {

	store := utxoSet.NewUtxoStore(blockchain.ctx)
	defer store.Close()
	utxo, ok := store.Read(outpt) // handle error
	if !ok {
		return nil, errors.New("key doesn't exit")
	}
	return utxo, nil
}

func (blockchain *Blockchain) GetFee(tx *transaction.Tx) int64 {
	var sumIn int64 = 0
	var sumOut int64 = 0

	for _, in := range tx.Inputs {
		utxo, err := blockchain.FindUTXO(&in.PrevOutpt)
		t_error.LogErr(err)
		sumIn += utxo.Value
	}

	for _, out := range tx.Outputs {
		sumOut += out.Value
	}

	return sumOut - sumIn
}
