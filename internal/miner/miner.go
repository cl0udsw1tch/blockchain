package miner

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/blockchain/proof"
	"github.com/terium-project/terium/internal/mempool"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/transaction"
)

type Miner struct {
	mempool     *mempool.MempoolIO
	block       *block.Block
	ctx         *t_config.Context
	blockchain  *blockchain.Blockchain
	pow         *proof.PoW
	Signal      *MinerSignal
}

func NewMiner(ctx *t_config.Context, blockchain *blockchain.Blockchain, mempool *mempool.MempoolIO) *Miner {

	miner := new(Miner)
	miner.ctx = ctx
	miner.mempool = mempool
	miner.Signal = NewMinerSignal()
	miner.blockchain = blockchain

	return miner
}

func (miner *Miner) SetBlock(block *block.Block) {
	miner.block = block
}

func (miner *Miner) Block() *block.Block {
	return miner.block
}

func (miner *Miner) Genesis() {

	if miner.blockchain.LastMeta() != nil {
		t_error.LogWarn(errors.New("blockchain already exists, wont create genesis block"))
		os.Exit(1)
	}
	genesis := block.Block{
		Header: block.Header{
			Version:   t_config.Version,
			PrevHash:  []byte{},
			TimeStamp: uint32(time.Now().Unix()),
			Target:    *t_config.Target,
		},
		TXCount:      0,
		Transactions: nil,
	}
	miner.block = &genesis
	miner.Mine(nil)
	miner.blockchain.AddGenesis(miner.block)
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
		PrevHash:  miner.blockchain.LastMeta().Hash,
		Target:    *t_config.Target,
		TimeStamp: uint32(time.Now().Unix()),
	}

	miner.block = &block.Block{
		Header:       header,
		TXCount:      1,
		Transactions: []transaction.Tx{coinbaseTx},
	}
}

type MineSignal struct {
	Stop   chan byte
	Resume chan byte
}

func (s *MineSignal) SignalStop() {
	s.Stop <- 0x00
}

func (s *MineSignal) SignalResume() {
	s.Resume <- 0x00
}

type SolveSignal struct {
	Ready chan byte
	Reset chan byte
}

func (s *SolveSignal) SignalReady() {
	s.Ready <- 0x00
}

func (s *SolveSignal) SignalReset() {
	s.Reset <- 0x00
}

type MinerSignal struct {
	MineSignal  MineSignal
	SolveSignal SolveSignal
}

func NewMinerSignal() *MinerSignal {

	m := &MinerSignal{
		MineSignal: MineSignal{
			Stop:   make(chan byte, 1),
			Resume: make(chan byte, 1)},
		SolveSignal: SolveSignal{
			Ready: make(chan byte, 1),
			Reset: make(chan byte, 1)},
	}
	return m
}

func (s *MinerSignal) Pause() {
	s.MineSignal.SignalStop()
	s.SolveSignal.SignalReset()
}
func (s *MinerSignal) Resume() {
	s.MineSignal.SignalResume()
}

func (miner *Miner) MineFromMempool() {

	for {
		select {
		case <-miner.Signal.MineSignal.Stop:
			<-miner.Signal.MineSignal.Resume
		default:
			time.Sleep(time.Millisecond * 10)
			txs := miner.mempool.GetTxByPriority(int64(*miner.ctx.NodeConfig.NumTxInBlock))
			if len(txs) == int(*miner.ctx.NodeConfig.NumTxInBlock) {
				for _, tx := range txs {
					miner.AddTxToBlock(&tx)
				}

				// where this node attempts to mine the block
				if miner.Mine(miner.Signal.SolveSignal.Reset) {
					miner.Signal.SolveSignal.SignalReady()
					miner.AddBlock(miner.block)
				}
			}
		}
	}
}

func (miner *Miner) AddTxToBlock(tx *transaction.Tx) {
	miner.block.Transactions = append(miner.block.Transactions, *tx)
	miner.block.TXCount++
}

// adds block to blockchain, updates UTXO set
func (miner *Miner) AddBlock(block *block.Block) {
	miner.blockchain.AddBlock(block)

}

func (miner *Miner) Mine(quit chan byte) bool {
	miner.pow = proof.NewPoW()
	defer miner.pow.Close()
	miner.block.Header.TimeStamp = uint32(time.Now().Unix())
	miner.block.Header.MerkleRootHash = miner.block.MerkelRoot()
	go func(){
		for state := range miner.pow.Notifier {
			fmt.Printf("\rNonce: %X\tHash: %X\t Solved: %t", state.Nonce, state.Hash, state.Solved)
			if state.Solved {
				var s string
				for _, n := range(state.Hash) {
					s += fmt.Sprintf("%08b", n)
				}
				fmt.Printf("\n\nBinary:\n%0*s", 256, s)
			}
		}
	}()
	return miner.pow.Solve(miner.block, quit)
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
