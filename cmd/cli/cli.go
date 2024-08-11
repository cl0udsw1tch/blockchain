package cli

import (
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"github.com/terium-project/terium/internal/node"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/wallet"
)

const EMPTY_STRING_ARG string = "NA"

type CommandLine struct{
	ctx *t_config.Context
}

func NewCommandLine(ctx *t_config.Context) *CommandLine {
	cli := new(CommandLine)
	cli.ctx = ctx
	return cli
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("command [--arg_name [arg_value]] ...")

	fmt.Println("\nrun | node")
	fmt.Printf("%-20s%-30s%s", "--numTx", "[num]", "Number of txs to include before mining current block. Default 10\n")
	fmt.Printf("%-20s%-30s%s", "--port", "[port]", "Port to listen and send on. Default "+fmt.Sprint(t_config.RpcEndpointPort)+"\n")
	fmt.Printf("%-20s%-30s%s", "--nodeAddr", "[address]", "Address for block rewards\n")

	fmt.Println("\nnode")
	fmt.Printf("%-20s%-30s%s", "--readBlk", "[block_hash]", "Reads block from .tmp into node\n")
	fmt.Printf("%-50s%s", "--writeBlk", "Writes block from node into .tmp\n")
	fmt.Printf("%-50s%s", "--createBlk", "Creates block\n")
	fmt.Printf("%-20s%-30s%s", "--mineBlk", "[block_hash]", "Mines block\n")
	fmt.Printf("%-20s%-30s%s", "--validateBlk", "[block_hash]", "Validates block\n")
	fmt.Printf("%-20s%-30s%s", "--addBlk", "[block_hash]", "Adds a mined and validated block to the local blockchain\n")
	fmt.Printf("%-20s%-30s%s", "--broadcastBlk", "[block_hash]", "Broadcasts a mined and validated block to the local blockchain\n")
	fmt.Printf("%-20s%-30s%s", "--validateTx", "[tx_hash]", "Validates a transaction\n")
	fmt.Printf("%-20s%-30s%s", "--addTxToMem", "[tx_hash]", "Adds tx to mempool. Validate first\n")
	fmt.Printf("%-20s%-30s%s", "--addTxToBlk", "[tx_hash]", "Adds a validated transaction to block\n")
	fmt.Printf("%-20s%-30s%s", "--addTxsToUtxos", "[tx_hash]", "Updates utxoset with txs from node's current block\n")

	fmt.Println("\nwallet")
	fmt.Printf("%-20s%-30s%s", "--name", "<name>", "Name of wallet to use, creates one if it doesnt exist\n")
	fmt.Printf("%-50s%s", "--balance", "Print balance\n")
	fmt.Printf("%-20s%-30s%s", "--tx", "<address:value,...>", "Create a transaction. arg is hex encoded address:value,...\n")
	fmt.Printf("%-20s%-30s%s", "--sendTx", "[tx_hash]", "Broadcast transaction. Arg is the txid.\n")

	fmt.Println("\nblockchain")
	fmt.Printf("%-50s%s", "--print", "Print the block header hashes of the main branch\n")
	fmt.Printf("%-50s%s", "--utxo", "Print the utxo outpoints in the utxo set\n")
	fmt.Printf("%-50s%s", "--mempool", "Print the tx hashes in the mempool\n")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) == 1 {
		cli.PrintUsage()
		os.Exit(1)
	}
	if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			cli.PrintUsage()
			os.Exit(1)
		}
	}
	switch os.Args[1] {
	case "run":
		cli.Run()
	case "node":
		cli.Node()
	case "genesis":
		cli.Genesis()
	case "wallet":
		cli.Wallet()
	case "blockchain":
		os.Exit(1)
		cli.Blockchain()
	case "interactive":
		os.Exit(1)
		cli.Interactive()
	default : 
		cli.PrintUsage()
		os.Exit(1)
	}
	
}

func (cli *CommandLine) Run() {

	var numTx int
	var port int
	var addr string

	flag.IntVar(&numTx, "numTx", 10, "")
	flag.StringVar(&addr, "nodeAddr", "", "")
	flag.IntVar(&port, "port", int(t_config.RpcEndpointPort), "")

	_numTx := uint8(numTx)
	_port := uint16(port)
	nodeConf := t_config.Config{
		NumTxInBlock: &_numTx,
		RpcEndpointPort: &_port,
		ClientAddress: &addr,

	}

	node := node.NewNode(cli.ctx, &nodeConf)
	go node.Run()
}

func (cli *CommandLine) Node() {
	var numTx int
	var port int
	var addr string

	flag.IntVar(&numTx, "numTx", 10, "")
	flag.StringVar(&addr, "nodeAddr", "", "")
	flag.IntVar(&port, "port", int(t_config.RpcEndpointPort), "")

	var readBlk string
	var writeBlk bool
	var createBlk bool
	var mineBlk string
	var valideBlk string
	var addBlk string
	var broadcastBlk string
	var validateTx string
	var addTxToMem string
	var addTxToBlk string
	var addTxsToUtxos string

	flag.StringVar(&readBlk, "readBlk", "", "")
	flag.BoolVar(&writeBlk, "writeBlk", false, "")
	flag.BoolVar(&createBlk, "createBlk", false, "")
	flag.StringVar(&mineBlk, "mineBlk", EMPTY_STRING_ARG, "")
	flag.StringVar(&valideBlk, "validateBlk", EMPTY_STRING_ARG, "")
	flag.StringVar(&addBlk, "addBlk", EMPTY_STRING_ARG, "")
	flag.StringVar(&broadcastBlk, "broadcastBlk", EMPTY_STRING_ARG, "")
	flag.StringVar(&validateTx, "validateTx", EMPTY_STRING_ARG, "")
	flag.StringVar(&addTxToMem, "addTxToMem", EMPTY_STRING_ARG, "")
	flag.StringVar(&addTxToBlk, "addTxToBlk", EMPTY_STRING_ARG, "")
	flag.StringVar(&addTxsToUtxos, "addTxsToUtxos", EMPTY_STRING_ARG, "")

	if len(os.Args) < 3 {
		cli.PrintUsage()
		os.Exit(1)
	}

	flag.CommandLine.Parse(os.Args[2:])

	if len(flag.Args()) > 0 {
		cli.PrintUsage()
		os.Exit(1)
	}

	_numTx := uint8(numTx)
	_port := uint16(port)
	nodeConf := t_config.Config{
		NumTxInBlock: &_numTx,
		RpcEndpointPort: &_port,
		ClientAddress: &addr,
	}

	node := node.NewNode(cli.ctx, &nodeConf)

	if readBlk != "" {
		node.ReadTmpBlock(readBlk)
	}

	if createBlk {
		node.CreateBlock(make([]byte, 0))
	}

	if mineBlk != EMPTY_STRING_ARG {
		if mineBlk == "" {
			cli.checkNodeBlock(node)
		} else {
			node.SetBlock(node.ReadTmpBlock(mineBlk))
		}
		node.Mine()
	}

	if valideBlk != EMPTY_STRING_ARG {
		if valideBlk == "" {
			cli.checkNodeBlock(node)
		} else {
			node.SetBlock(node.ReadTmpBlock(valideBlk))
		}
		node.ValidateBlock()
	}

	if addBlk != EMPTY_STRING_ARG {
		if addBlk == "" {
			cli.checkNodeBlock(node)
		} else {
			node.SetBlock(node.ReadTmpBlock(addBlk))
		}
		node.AddBlock()
	}

	if broadcastBlk != EMPTY_STRING_ARG {
		if broadcastBlk == "" {
			cli.checkNodeBlock(node)
		} else {
			node.SetBlock(node.ReadTmpBlock(broadcastBlk))
		}
		node.Broadcast()
	}

	if validateTx != EMPTY_STRING_ARG {
		if validateTx == "" {
			cli.checkNodeTx(node)
		} else {
			node.SetTx(node.ReadTmpTx(validateTx))
		}
		node.ValidateTx()
	}

	if addTxToMem != EMPTY_STRING_ARG {
		if addTxToMem == "" {
			cli.checkNodeTx(node)
		} else {
			node.SetTx(node.ReadTmpTx(addTxToMem))
		}
		node.AddTxToPool()
	}

	if addTxToBlk != EMPTY_STRING_ARG {
		if addTxToBlk == "" {
			cli.checkNodeTx(node)
		} else {
			node.SetTx(node.ReadTmpTx(addTxToBlk))
		}
		cli.checkNodeBlock(node)
		node.AddTxToBlock()
	}

	if addTxsToUtxos != EMPTY_STRING_ARG {
		if addTxsToUtxos == "" {
			cli.checkNodeBlock(node)
		} else {
			node.SetBlock(node.ReadTmpBlock(addTxsToUtxos))
		}
		node.UpdateUtxoSet()
	}

	if writeBlk {
		cli.checkNodeBlock(node)
		node.WriteTmpBlock()
	}
}


func (cli *CommandLine) Interactive() {

}


func (cli *CommandLine) Wallet() {
	var name string
	var balance bool
	var tx string
	var sendTx string
	var getAddr bool

	flag.StringVar(&name, "name", "", "")
	flag.BoolVar(&balance, "balance", false, "")
	flag.StringVar(&tx, "tx", "", "")
	flag.StringVar(&sendTx, "sendTx", "", "")
	flag.BoolVar(&getAddr, "getAddr", false, "")

	flag.CommandLine.Parse(os.Args[2:])

	if name == "" {
		fmt.Println("name of wallet required")
		cli.PrintUsage()
		os.Exit(1)
	}

	w := wallet.NewWallet(cli.ctx, name)
	if !w.Exists() {
		w.Create()
	} else {
		w.Read()
	}
	wc := wallet.NewWalletController(w, cli.ctx)
	
	if balance {
		fmt.Printf("%s balance: %d terium.", w.ClientId.Address, wc.Balance())
	}

	if getAddr {
		fmt.Printf("Address: %s", w.ClientId.Address)
	}

	if tx != "" {
		addrs, amnts := cli.ParseTx(tx)
		tx := wc.GenP2PKH(nil, addrs, amnts, nil, 0)
		os.WriteFile(path.Join(cli.ctx.TmpDir, "txs", hex.EncodeToString(tx.Hash())), tx.Serialize(), 0666)
	} 

	if sendTx != "" {
		b, err := os.ReadFile(path.Join(cli.ctx.TmpDir, "txs", sendTx))
		t_error.LogErr(err)
		dec := transaction.NewTxDecoder(nil)
		buffer := bytes.Buffer{}
		buffer.Write(b)
		dec.Decode(&buffer)
		wc.BroadcastTx(dec.Out())
	} 

}

func (cli *CommandLine) Blockchain() {

}

func (cli *CommandLine) ParseTx(tx string) ([]string, []int64) {
	frags := strings.Split(tx, ",")
	a := make([]string, len(frags))
	v := make([]int64, len(frags))

	for i, frag := range frags {
		s := strings.Split(frag, ":")
		if len(s) != 2 {
			t_error.LogErr(errors.New("bad tx"))
		}
		a[i] = s[0]
		var err error
		v[i], err = strconv.ParseInt(s[1], 10, 64)
		t_error.LogErr(err)
	}
	return a,v
}

func (cli *CommandLine) Genesis() {
	var nodeAddr string
	flag.StringVar(&nodeAddr, "nodeAddr", "", "")

	flag.CommandLine.Parse(os.Args[2:])

	_numTx := uint8(10)
	_port := uint16(t_config.RpcEndpointPort)
	nodeConf := t_config.Config{
		NumTxInBlock: &_numTx,
		RpcEndpointPort: &_port,
		ClientAddress: &nodeAddr,
	}

	node := node.NewNode(cli.ctx, &nodeConf)
	node.Genesis()
}

func (cli *CommandLine) checkNodeBlock(node *node.Node) {
	if node.GetBlock() == nil {
		fmt.Println("Node has no block.")
		os.Exit(1)
	}
}

func (cli *CommandLine) checkNodeTx(node *node.Node) {
	if node.GetTx() == nil {
		fmt.Println("Node has no tx.")
		os.Exit(1)
	}
}