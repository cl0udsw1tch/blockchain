package node

import (
	"bufio"
	"fmt"
	"os"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/mempool"
	"github.com/terium-project/terium/internal/miner"
	"github.com/terium-project/terium/internal/server"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/validator"
	"github.com/terium-project/terium/internal/wallet"
)

type Node struct {
	miner *miner.Miner
	ctx   *t_config.Context
	server *server.Server
	txValidator *validator.TxValidator
	blockValidator *validator.BlockValidator
	mempool *mempool.MempoolIO
	blockchain *blockchain.Blockchain
}

func NewNode(ctx *t_config.Context, newConf *t_config.Config) *Node {
	
	
	node := new(Node)
	node.ctx = ctx
	
	if node.ctx.NodeConfig.ClientAddress == nil {
		if newConf.ClientAddress == nil {
			fmt.Print("Enter name for new wallet for node: ")
			scan := bufio.NewScanner(os.Stdin)
			scan.Scan()
			wallet := wallet.NewWallet(ctx, scan.Text())
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
	
	node.server = server.NewServer(node.ctx)
	node.txValidator = validator.NewTxValidator(node.ctx)
	node.mempool = mempool.NewMempoolIO(node.ctx)
	node.miner = miner.NewMiner(node.ctx)
	node.blockchain = blockchain.NewBlockchain(node.ctx)


	node.server.Run()
	return node
}

func (node *Node) Run() {
	node.CreateBlock(make([]byte, 0))
	go node.Miner.MonitorMempool()
	sigChan := make(chan byte, 1)
	ackChan := make(chan byte, 1)
	isMining := false
	for {
		select { 
		case tx := <- node.server.Tx().OutStream :
			if !node.ValidateTx() {
				fmt.Println("Invalid tx")
				continue
			} else {
				node.mempool.Write(tx.Hash(), tx, node.blockchain.GetFee(tx))
			}
			if node.miner.Block().TXCount == uint32(*node.ctx.NodeConfig.NumTxInBlock) {
				isMining = true
				node.miner.AddBlock(sigChan, ackChan)
				isMining = false
				node.server.Block().InStream <- node.miner.Block()
			}
		case block := <- node.server.Block().OutStream : 
			if isMining {
				isMining = false
			}
			sigChan<-0x00
			<-ackChan
			<-sigChan
		}
	}
}




func (node *Node) CreateBlock(coinbaseScript []byte) {
	node.miner.CreateBlock([][]byte{coinbaseScript})
}

func (node *Node) AddBlock() {
	node.miner.AddBlock()
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
	node.miner.
}
// goroutines

func (node *Node) TxListen() <-chan *transaction.Tx {
	ch := make(chan *transaction.Tx)
	return ch
}

func (node *Node) Broadcast() {

}
