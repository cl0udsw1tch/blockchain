package miner

import (
	"math/big"
	"time"
	"github.com/terium-project/terium/internal"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/blockchain/proof"
	"github.com/terium-project/terium/internal/transaction"
)
const (
	Version int32 = 0x01
	NBits 	uint8 = 0x40
)

var (
	Target big.Int = *(&big.Int{}).Lsh(big.NewInt(1), uint(NBits))
)

type Miner struct {
	block *block.Block
	ctx *internal.DirCtx
	blockchain *blockchain.Blockchain
}

func (miner Miner) New(ctx *internal.DirCtx) {
	miner.ctx = ctx
	miner.blockchain = &blockchain.Blockchain{}
	miner.blockchain.New(ctx)
}

func (miner Miner) Genesis() {
	genesis := block.Block{
		Header: block.Header{
			Version: Version,
			PrevHash: []byte{},
			TimeStamp: uint32(time.Now().Unix()),
			Target: Target,
		},
		TXCount: 0,
		Transactions: nil,
	}
	miner.block = &genesis
	miner.Mine()
	miner.blockchain.AddBlock(miner.block)
}

func (miner Miner) CreateBlock( 
	CoinBaseScriptSz transaction.CompactSize, 
	CoinBaseSript [][]byte,
	nOut uint8, 
	outputs transaction.TxOuts, 
	) {

	coinbaseTx := transaction.CoinbaseTx(uint32(Version), CoinBaseScriptSz, CoinBaseSript, nOut, outputs)
	
	header := block.Header{
		Version: Version,
		PrevHash: miner.blockchain.LastHash(),
		Target: Target,
		TimeStamp: uint32(time.Now().Unix()),
	}

	miner.block = &block.Block{
		Header: header,
		TXCount: 1,
		Transactions: []transaction.Tx{coinbaseTx},
	}
}

func (miner Miner) AddBlock() {
	miner.Mine()
	miner.blockchain.AddBlock(miner.block)
}

func (miner Miner) Mine() {
	miner.block.Header.TimeStamp = uint32(time.Now().Unix())
	miner.block.Header.MerkleRootHash = miner.block.MerkelRoot()
	pow := proof.NewPoW(miner.block)
	pow.Solve()
}