package miner

import (
	"sync"
	"time"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/blockchain/proof"
	"github.com/terium-project/terium/internal/mempool"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/utxoSet"
	"github.com/terium-project/terium/internal/validator"
)

type Miner struct {
	mempool     *mempool.MempoolIO
	block       *block.Block
	ctx         *t_config.Context
	blockchain  *blockchain.Blockchain
	txValidator *validator.TxValidator
}

func NewMiner(ctx *t_config.Context) *Miner {

	miner := new(Miner)
	miner.ctx = ctx
	miner.blockchain = blockchain.NewBlockchain(ctx)
	miner.txValidator = validator.NewTxValidator(ctx)
	miner.mempool = mempool.NewMempoolIO(ctx)
	return miner
}

func (miner *Miner) Block() *block.Block {
	return miner.block
}

func (miner *Miner) Genesis() {
	genesis := block.Block{
		Header: block.Header{
			Version:   t_config.Version,
			PrevHash:  []byte{},
			TimeStamp: uint32(time.Now().Unix()),
			Target:    t_config.Target,
		},
		TXCount:      0,
		Transactions: nil,
	}
	miner.block = &genesis
	miner.Mine()
	miner.blockchain.AddBlock(miner.block)
}

func (miner *Miner) CreateBlock(coinbaseSript [][]byte) {

	var n int64 = 0
	for _, s := range coinbaseSript {
		n += int64(len(s))
	}
	coinbaseScriptSz := transaction.NewCompactSize(n)
	coinbaseTx := miner.CoinbaseTx(uint32(t_config.Version), coinbaseScriptSz, coinbaseSript)

	header := block.Header{
		Version:   t_config.Version,
		PrevHash:  miner.blockchain.LastHash(),
		Target:    t_config.Target,
		TimeStamp: uint32(time.Now().Unix()),
	}

	miner.block = &block.Block{
		Header:       header,
		TXCount:      1,
		Transactions: []transaction.Tx{coinbaseTx},
	}
}

func (miner *Miner) MonitorMempool() {
	txs := []transaction.Tx{}
	for len(txs) == 0 {
		time.Sleep(time.Millisecond * 10)
		miner.mempool.GetTxByPriority(int64(*miner.ctx.NodeConfig.NumTxInBlock))
	}
	for _, tx := range txs {
		miner.AddTxToBlock(&tx)
	}
}

func (miner *Miner) AddTxToBlock(tx *transaction.Tx) {
	miner.block.Transactions = append(miner.block.Transactions, *tx)
	miner.block.TXCount++
}

func (miner *Miner) AddBlock(quitSig <-chan byte, ackChan chan<- byte) {
	
	miner.Mine(quitSig, ackChan)
	miner.blockchain.AddBlock(miner.block)
	miner.UpdateUtxoSet()
}

func (miner *Miner) Mine(quitSig <-chan byte, ackChan chan<- byte) {
	miner.block.Header.TimeStamp = uint32(time.Now().Unix())
	miner.block.Header.MerkleRootHash = miner.block.MerkelRoot()
	pow := proof.NewPoW(miner.block)
	pow.Solve(quitSig, ackChan)
}

func (miner *Miner) UpdateUtxoSet() {
	if miner.block.Transactions == nil {
		return
	}
	wg := sync.WaitGroup{}
	utxoStore := utxoSet.NewUtxoStore(miner.ctx)
	defer utxoStore.Close()
	for _, tx := range miner.block.Transactions {
		wg.Add(1)
		go func(tx *transaction.Tx) {
			defer wg.Done()
			for idx, out := range tx.Outputs {
				wg.Add(1)
				go func(out *transaction.TxOut, idx int32) {
					defer wg.Done()
					utxo := transaction.Utxo{
						OutRef: transaction.OutPoint{
							TxId: tx.Hash(),
							Idx:  idx,
						},
						Value:             out.Value,
						LockingScriptSize: out.LockingScriptSize,
						LockingScript:     out.LockingScript,
					}
					utxoStore.Write(&utxo)
				}(&out, int32(idx))
			}
		}(&tx)
	}
	wg.Wait()
}

func (miner *Miner) CoinbaseTx(
	version uint32,
	inScriptSz transaction.CompactSize,
	inSript [][]byte) transaction.Tx {

	output := transaction.TxOut{
		Value:             t_config.BLOCK_REWARD,
		LockingScriptSize: transaction.NewCompactSize(transaction.P2PKH_LOCK_SCRIPT_SZ),
		LockingScript:     transaction.P2PKH_LockScript(*miner.ctx.NodeConfig.ClientAddress),
	}

	return transaction.Tx{
		Version:    t_config.Version,
		NumInputs:  1,
		Inputs:     []transaction.TxIn{transaction.Coinbase(inScriptSz, inSript)},
		NumOutputs: 1,
		Outputs:    []transaction.TxOut{output},
		LockTime:   100,
	}
}
