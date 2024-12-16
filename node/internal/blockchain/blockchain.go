package blockchain

import (
	"errors"
	"math/big"

	"github.com/tiereum/trmnode/internal/block"
	"github.com/tiereum/trmnode/internal/blockStore"
	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"
	"github.com/tiereum/trmnode/internal/transaction"
	"github.com/tiereum/trmnode/internal/utxoSet"
)

type Blockchain struct {
	ctx        *t_config.Context
	blockStore *blockStore.BlockStore
	lastMeta   *blockStore.BlockMetaData
	iter       *BlockchainIterator
}

func NewBlockchain(ctx *t_config.Context, store *blockStore.BlockStore) *Blockchain {

	b := new(Blockchain)
	b.ctx = ctx
	b.blockStore = store
	_, meta, err := b.blockStore.ReadLast()
	if err == blockStore.ErrNoBlocksRemaining {
		b.lastMeta = nil
	} else {
		b.lastMeta = meta
	}
	b.iter = NewBlockchainIterator(b.blockStore)
	return b
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	return blockchain.iter
}

func (blockchain *Blockchain) AddGenesis(block *block.Block) {

	metadata := blockStore.BlockMetaData{
		Hash:   block.Hash(),
		Height: *big.NewInt(0),
		Nonce:  block.Header.Nonce,
	}
	blockchain.blockStore.Write(block, &metadata)
	blockchain.lastMeta = &metadata
}

func (blockchain *Blockchain) AddBlock(block *block.Block) {
	var currHeight *big.Int
	if blockchain.lastMeta == nil {
		currHeight = big.NewInt(0)
	} else {
		currHeight = &blockchain.lastMeta.Height
	}

	sum := big.Int{}
	sum.Add(currHeight, big.NewInt(1))

	metadata := blockStore.BlockMetaData{
		Hash:   block.Hash(),
		Height: sum,
		Nonce:  block.Header.Nonce,
	}
	blockchain.blockStore.Write(block, &metadata)
	blockchain.lastMeta = &metadata
}

func (blockchain *Blockchain) Block(hash []byte) (*block.Block, *blockStore.BlockMetaData) {
	return blockchain.blockStore.Read(hash)
}

func (blockchain *Blockchain) LastMeta() *blockStore.BlockMetaData {
	return blockchain.lastMeta
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

type BlockchainIterator struct {
	valid    bool
	block    *block.Block
	metadata *blockStore.BlockMetaData
	store    *blockStore.BlockStore
	start    bool
}

func NewBlockchainIterator(store *blockStore.BlockStore) *BlockchainIterator {
	iter := BlockchainIterator{}
	iter.valid = false
	iter.store = store
	iter.start = true
	return &iter
}

func (iter *BlockchainIterator) Valid() bool {
	return iter.valid
}

func (iter *BlockchainIterator) Next() {
	if iter.start {
		iter.start = false
		block, meta, err := iter.store.ReadLast()
		if err == blockStore.ErrNoBlocksRemaining {
			return
		}
		iter.block = block
		iter.metadata = meta
		iter.valid = true
		return
	} else if iter.valid {
		if len(iter.block.Header.PrevHash) > 0 {
			iter.block, iter.metadata = iter.store.Read(iter.block.Header.PrevHash)
		} else {
			iter.valid = false
			iter.block = nil
			iter.metadata = nil
		}
	}
}

func (iter *BlockchainIterator) Block() *block.Block {
	return iter.block
}

func (iter *BlockchainIterator) Metadata() *blockStore.BlockMetaData {
	return iter.metadata
}

func (blockchain *Blockchain) Height() big.Int {
	return blockchain.lastMeta.Height
}
