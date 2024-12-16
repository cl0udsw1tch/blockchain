package node

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/tiereum/trmnode/internal/block"
	"github.com/tiereum/trmnode/internal/blockStore"
	"github.com/tiereum/trmnode/internal/blockchain"
	"github.com/tiereum/trmnode/internal/mempool"
	"github.com/tiereum/trmnode/internal/miner"
	"github.com/tiereum/trmnode/internal/server"
	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"
	"github.com/tiereum/trmnode/internal/transaction"
	"github.com/tiereum/trmnode/internal/utxoSet"
	"github.com/tiereum/trmnode/internal/validator"
	"github.com/tiereum/trmnode/internal/wallet"
)

type Node struct {
	miner          *miner.Miner
	ctx            *t_config.Context
	server         *server.Server
	txValidator    *validator.TxValidator
	blockValidator *validator.BlockValidator
	blockchain     *blockchain.Blockchain
	blockStore     *blockStore.BlockStore
	txIndex        *transaction.TxIndexIO
	mempool        *mempool.MempoolIO
	utxoStore      *utxoSet.UtxoStore
	block          *block.Block
	tx             *transaction.Tx
}

func NewNode(ctx *t_config.Context, newConf *t_config.Config) *Node {

	node := new(Node)
	node.ctx = ctx

	if node.ctx.NodeConfig.ClientAddress == nil {
		if *newConf.ClientAddress == "" {
			fmt.Println("Enter name for new wallet for node: ")
			var name string
			fmt.Scan(&name)
			wallet := wallet.NewWallet(ctx, name)
			wallet.Create()
			node.ctx.NodeConfig.ClientAddress = &wallet.ClientId.Address
		} else {
			valid := wallet.ValidateAddress(*newConf.ClientAddress)
			if !valid {
				panic("Invalid address for node.")
			}
			node.ctx.NodeConfig.ClientAddress = newConf.ClientAddress
		}
	}
	if node.ctx.NodeConfig.NumTxInBlock == nil {
		node.ctx.NodeConfig.NumTxInBlock = newConf.NumTxInBlock
	}
	if node.ctx.NodeConfig.RpcEndpointPort == nil {
		node.ctx.NodeConfig.RpcEndpointPort = newConf.RpcEndpointPort
	}
	node.blockStore = blockStore.NewBlockStore(node.ctx)
	node.txIndex = transaction.NewTxIndexIO(node.ctx)
	node.mempool = mempool.NewMempoolIO(node.ctx)
	node.utxoStore = utxoSet.NewUtxoStore(node.ctx)
	node.server = server.NewServer(node.ctx)

	node.blockchain = blockchain.NewBlockchain(node.ctx, node.blockStore)
	node.txValidator = validator.NewTxValidator(
		node.ctx,
		node.txIndex,
		node.blockStore,
		node.mempool,
		node.utxoStore)
	node.blockValidator = validator.NewBlockValidator(
		ctx,
		node.blockchain,
		node.txValidator)

	node.miner = miner.NewMiner(node.ctx, node.blockchain, node.mempool)

	return node
}

func (node *Node) GetBlock() *block.Block {
	return node.block
}

func (node *Node) SetBlock(block *block.Block) {
	node.block = block
}

func (node *Node) GetTx() *transaction.Tx {
	return node.tx
}

func (node *Node) SetTx(tx *transaction.Tx) {
	node.tx = tx
}

func (node *Node) ReadTmpBlock(hash string) *block.Block {
	p := path.Join(node.ctx.TmpDir, "blocks", hash)
	return node.readBlock(p)
}

func (node *Node) ReadBlock(hash string) *block.Block {
	p := path.Join(node.ctx.DataDir, hash)
	return node.readBlock(p)
}

func (node *Node) readBlock(path string) *block.Block {
	b, err := os.ReadFile(path)
	t_error.LogErr(err)
	buffer := new(bytes.Buffer)
	buffer.Write(b)
	dec := block.NewBlockDecoder(nil)
	err = dec.Decode(buffer)
	t_error.LogErr(err)
	return dec.Out()
}

func (node *Node) WriteTmpBlock() {
	p := path.Join(node.ctx.TmpDir, "blocks", hex.EncodeToString(node.block.Hash()))
	node.writeBlock(p)
}

func (node *Node) WriteBlock() {
	p := path.Join(node.ctx.DataDir, hex.EncodeToString(node.block.Hash()))
	node.writeBlock(p)
}

func (node *Node) writeBlock(path string) {
	b := node.block.Serialize()
	err := os.WriteFile(path, b, 0600)
	t_error.LogErr(err)
}

func (node *Node) DeleteTmpBlock(hash string) {
	p := path.Join(node.ctx.TmpDir, "blocks", hex.EncodeToString(node.block.Hash()))
	os.Remove(p)
}

func (node *Node) DeleteBlock(hash string) {
	p := path.Join(node.ctx.DataDir, hex.EncodeToString(node.block.Hash()))
	os.Remove(p)
}

func (node *Node) ReadTmpTx(hash string) *transaction.Tx {
	path := path.Join(node.ctx.TmpDir, "txs", hash)
	b, err := os.ReadFile(path)
	t_error.LogErr(err)
	buffer := new(bytes.Buffer)
	buffer.Write(b[:len(b)-32])
	dec := transaction.NewTxDecoder(nil)
	err = dec.Decode(buffer)
	t_error.LogErr(err)
	return dec.Out()
}

func (node *Node) WriteTmpTx() {
	path := path.Join(node.ctx.TmpDir, "txs", hex.EncodeToString(node.tx.Hash()))
	b := node.tx.Serialize()
	os.WriteFile(path, b, 0600)
}

func (node *Node) DeleteTmpTx(hash string) {
	p := path.Join(node.ctx.TmpDir, "txs", hex.EncodeToString(node.tx.Hash()))
	os.Remove(p)
}

func (node *Node) Run() {

	node.server.Run()
	node.CreateBlock(make([]byte, 0))
	node.StartMiner()

	for {
		select {

		case <-node.miner.Signal.SolveSignal.Ready:
			// node has mined a block and added it to the blockchain
			node.server.Block().InStream <- node.block
			node.UpdateUtxoSet()
			node.UpdateTxIndex()
			node.AddBlock()
			node.CreateBlock(make([]byte, 0))

		case tx := <-node.server.Tx().OutStream:
			// incoming tx from network
			node.tx = tx
			if !node.ValidateTx() {
				fmt.Println("Invalid tx")
			} else {
				node.AddTxToPool()
			}

		case block := <-node.server.Block().OutStream:
			// incoming block from network
			if node.ValidateBlock() {
				node.block = block
				node.PauseMiner()
				node.AddBlock()
				node.CreateBlock(make([]byte, 0))
				node.ResumeMiner()
			}
		}
	}
}

func (node *Node) StartMiner() {
	go node.miner.MineFromMempool(node.block)
}

func (node *Node) PauseMiner() {
	node.miner.Signal.Pause()
}

func (node *Node) ResumeMiner() {
	node.miner.Signal.Resume()
}

func (node *Node) StartInteractive() {

	for {
		select {
		case <-node.miner.Signal.SolveSignal.Ready:
			os.WriteFile(path.Join(node.ctx.TmpDir, "blocks", hex.EncodeToString(node.block.Hash())), node.block.Serialize(), 0666)

		case tx := <-node.server.Tx().OutStream:
			os.WriteFile(path.Join(node.ctx.TmpDir, "txs", hex.EncodeToString(tx.Hash())), tx.Serialize(), 0666)

		case block := <-node.server.Block().OutStream:
			os.WriteFile(path.Join(node.ctx.TmpDir, "blocks", hex.EncodeToString(block.Hash())), block.Serialize(), 0666)
		}
	}
}

func (node *Node) CreateBlock(coinbaseScript []byte) {
	node.block = node.miner.CreateBlock(coinbaseScript)
}

func (node *Node) Mine() {
	node.miner.Mine(nil, node.block)
}

// Adds block to blockchain, updates UTXO set
func (node *Node) AddBlock() {
	node.miner.AddBlock(node.block)
}

func (node *Node) ValidateBlock() bool {
	if node.block != nil {
		return node.blockValidator.Validate(node.block)
	}
	return node.blockValidator.Validate(node.block)
}

func (node *Node) AddTxToBlock() {
	node.miner.AddTxToBlock(node.tx, node.block)
}

func (node *Node) AddTxToPool() error {
	if !node.txValidator.ValidateTx(node.tx) {
		return errors.New("invalid tx")
	}
	node.mempool.Write(node.tx.Hash(), node.tx, node.blockchain.GetFee(node.tx))
	return nil
}

func (node *Node) ValidateTx() bool {
	return node.txValidator.ValidateTx(node.tx)
}

// goroutines

func (node *Node) TxListen() <-chan *transaction.Tx {
	ch := make(chan *transaction.Tx)
	return ch
}

func (node *Node) UpdateUtxoSet() {

	if node.block.Transactions == nil {
		return
	}
	wg := sync.WaitGroup{}
	for _, tx := range node.block.Transactions {
		wg.Add(1)
		go func(tx *transaction.Tx) {
			defer wg.Done()
			for idx, out := range tx.Outputs {
				wg.Add(1)
				go func(out *transaction.TxOut, idx int32) {
					defer wg.Done()
					utxo := transaction.Utxo{
						OutPoint: transaction.OutPoint{
							TxId: tx.Hash(),
							Idx:  idx,
						},
						Value:             out.Value,
						LockingScriptSize: out.LockingScriptSize,
						LockingScript:     out.LockingScript,
					}
					node.utxoStore.Write(&utxo)
				}(&out, int32(idx))
			}
		}(&tx)
	}
	wg.Wait()
}

func (node *Node) UpdateTxIndex() {
	for i, tx := range node.block.Transactions {
		node.txIndex.Create(tx.Hash(), &transaction.TxMetadata{
			BlockHash:   node.block.Hash(),
			BlockHeight: node.blockchain.Height(),
			Index:       uint8(i),
		})
	}
}

func (node *Node) Broadcast() {
	node.server.Block().InStream <- node.block
}

func (node *Node) Genesis() {
	node.block = node.miner.Genesis()
	node.UpdateUtxoSet()
}
