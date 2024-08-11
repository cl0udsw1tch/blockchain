package miner

import (
	"encoding/hex"
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

func (miner *Miner) Genesis() *block.Block {

	if miner.blockchain.LastMeta() != nil {
		t_error.LogWarn(errors.New("blockchain already exists, wont create genesis block"))
		os.Exit(1)
	}

	genesis := block.Block{
		Header: block.Header{
			Version:   t_config.Version,
			PrevHash:  make([]byte, 32),
			TimeStamp: uint32(time.Now().Unix()),
			Target:    *t_config.Target,
		},
		TXCount:      1,
		Transactions: []transaction.Tx{miner.CoinbaseTx(uint32(t_config.Version), transaction.NewCompactSize(0), make([]byte, 0))},
	}

	miner.Mine(nil, &genesis)
	miner.blockchain.AddGenesis(&genesis)
	return &genesis
}

func (miner *Miner) CreateBlock(coinbaseSript []byte) *block.Block {


	coinbaseScriptSz := transaction.NewCompactSize(int64(len(coinbaseSript)))
	coinbaseTx := miner.CoinbaseTx(uint32(t_config.Version), coinbaseScriptSz, coinbaseSript)

	header := block.Header{
		Version:   t_config.Version,
		PrevHash:  miner.blockchain.LastMeta().Hash,
		Target:    *t_config.Target,
		TimeStamp: uint32(time.Now().Unix()),
	}

	return &block.Block{
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

func (miner *Miner) MineFromMempool(block *block.Block) {

	for {
		select {
		case <-miner.Signal.MineSignal.Stop:
			<-miner.Signal.MineSignal.Resume
		default:
			time.Sleep(time.Millisecond * 10)
			txs := miner.mempool.GetTxByPriority(int64(*miner.ctx.NodeConfig.NumTxInBlock))
			if len(txs) == int(*miner.ctx.NodeConfig.NumTxInBlock) {
				for _, tx := range txs {
					miner.AddTxToBlock(&tx, block)
				}

				// where this node attempts to mine the block
				if miner.Mine(miner.Signal.SolveSignal.Reset, block) {
					miner.Signal.SolveSignal.SignalReady()
					miner.AddBlock(block)
				}
			}
		}
	}
}

func (miner *Miner) AddTxToBlock(tx *transaction.Tx, block *block.Block) {
	block.Transactions = append(block.Transactions, *tx)
	block.TXCount++
}

// adds block to blockchain, updates UTXO set
func (miner *Miner) AddBlock(block *block.Block) {
	miner.blockchain.AddBlock(block)
}

func (miner *Miner) Mine(quit chan byte, block *block.Block) bool {
	miner.pow = proof.NewPoW()
	defer miner.pow.Close()
	block.Header.TimeStamp = uint32(time.Now().Unix())
	block.Header.MerkleRootHash = block.MerkelRoot()
	go func(){
		for state := range miner.pow.Notifier {
			fmt.Printf("\rNonce: %X\tHash: %s\t Solved: %t", state.Nonce, hex.EncodeToString(state.Hash), state.Solved)
			if state.Solved {
				var s string
				for _, n := range(state.Hash) {
					s += fmt.Sprintf("%08b", n)
				}
				fmt.Printf("\n\nBinary:\n%0*s\n", 256, s)
			}
		}
	}()
	return miner.pow.Solve(block, quit)
}


func (miner *Miner) CoinbaseTx(
	version uint32,
	inScriptSz transaction.CompactSize,
	inSript []byte) transaction.Tx {

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
