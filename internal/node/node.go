package node

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockStore"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/mempool"
	"github.com/terium-project/terium/internal/miner"
	"github.com/terium-project/terium/internal/server"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/utxoSet"
	"github.com/terium-project/terium/internal/validator"
	"github.com/terium-project/terium/internal/wallet"
)

type Node struct {
	miner *miner.Miner
	ctx   *t_config.Context
	server *server.Server
	txValidator *validator.TxValidator
	blockValidator *validator.BlockValidator
	blockchain *blockchain.Blockchain
	blockStore *blockStore.BlockStore
	txIndex *transaction.TxIndexIO
	mempool *mempool.MempoolIO
	utxoStore *utxoSet.UtxoStore
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
	node.server.Run()
	return node
}

func (node *Node) Run() {

	node.CreateBlock(make([]byte, 0))
	node.StartMiner()

	for {
		select { 

		case <-node.miner.Signal.SolveSignal.Ready :
			// node has mined a block and added it to the blockchain
			block := node.miner.Block()
			node.server.Block().InStream<- block
			node.UpdateUtxoSet(block)
			node.UpdateTxIndex(block)
			node.CreateBlock(make([]byte, 0))

		case tx := <-node.server.Tx().OutStream :
			// incoming tx from network
			if !node.ValidateTx(tx) {
				fmt.Println("Invalid tx")
			} else {
				node.UpdateMempool(tx)
			}

		case block := <-node.server.Block().OutStream : 
			// incoming block from network
			if node.ValidateBlock(block) {
				node.PauseMiner()
				node.AddBlock(block)
				node.CreateBlock(make([]byte, 0))
				node.ResumeMiner()
			}
		}
	}
}

func (node *Node) StartMiner() {
	go node.miner.MineFromMempool()
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
		case <-node.miner.Signal.SolveSignal.Ready :
			os.WriteFile(path.Join(node.ctx.TmpDir, "blocks", hex.EncodeToString(node.miner.Block().Hash())), node.miner.Block().Serialize(), 0666)

		case tx := <-node.server.Tx().OutStream :
			os.WriteFile(path.Join(node.ctx.TmpDir, "txs", hex.EncodeToString(tx.Hash())), tx.Serialize(), 0666)

		case block := <-node.server.Block().OutStream : 
			os.WriteFile(path.Join(node.ctx.TmpDir, "blocks", hex.EncodeToString(block.Hash())), block.Serialize(), 0666)
		}
	}
}


func (node *Node) CreateBlock(coinbaseScript []byte) {
	node.miner.CreateBlock([][]byte{coinbaseScript})
}

// Adds block to blockchain, updates UTXO set
func (node *Node) AddBlock(block *block.Block) {
	if block != nil {
		node.miner.AddBlock(block)
		return
	}
	node.miner.AddBlock(node.miner.Block())
}


func (node *Node) ValidateBlock(block *block.Block) bool {
	if block != nil {
		return node.blockValidator.Validate(block)
	}
	return node.blockValidator.Validate(node.miner.Block())
}

func (node *Node) AddTxToBlock(tx *transaction.Tx) {
	node.miner.AddTxToBlock(tx)
}

func (node *Node) AddTxToPool(tx *transaction.Tx) error {
	if !node.txValidator.ValidateTx(tx) {
		return errors.New("invalid tx")
	}
	node.mempool.Write(tx.Hash(), tx, node.blockchain.GetFee(tx))
	return nil
}

func (node *Node) ValidateTx(tx *transaction.Tx) bool {
	return node.txValidator.ValidateTx(tx)
}
// goroutines

func (node *Node) TxListen() <-chan *transaction.Tx {
	ch := make(chan *transaction.Tx)
	return ch
}

func (node *Node) UpdateUtxoSet(block *block.Block) {
	if block.Transactions == nil {
		return
	}
	wg := sync.WaitGroup{}
	defer node.utxoStore.Close()
	for _, tx := range block.Transactions {
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

func (node *Node) UpdateTxIndex(block *block.Block) {
	for i, tx := range block.Transactions {
		node.txIndex.Create(tx.Hash(), &transaction.TxMetadata{
			BlockHash:   block.Hash(),
			BlockHeight: node.blockchain.Height(),
			Index:       uint8(i),
		})
	}
}

func (node *Node) UpdateMempool(tx *transaction.Tx) {
	node.mempool.Write(tx.Hash(), tx, node.blockchain.GetFee(tx))
}

func (node *Node) Broadcast(block *block.Block) {
	if block != nil {
		node.server.Block().InStream<- block
		return
	}
	node.server.Block().InStream<- node.miner.Block()
}

func (node *Node) Mine(block *block.Block) {
	if block != nil {
		node.miner.SetBlock(block)
	} 
	node.miner.Mine(nil)
	
}