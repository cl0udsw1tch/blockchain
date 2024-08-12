package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tiereum/trmnode/internal/node"
	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"
	"github.com/tiereum/trmnode/internal/wallet"
)

const EMPTY_STRING_ARG string = "NA"

type CommandLine struct {
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
	default:
		cli.PrintUsage()
		os.Exit(1)
	}

}

func (cli *CommandLine) Run() {

	nodeConf, _ := cli.extractConf()
	node := node.NewNode(cli.ctx, nodeConf)
	go node.Run()
}

func (cli *CommandLine) Node() {

	nodeConf, args := cli.extractConf()
	node := node.NewNode(cli.ctx, nodeConf)

	i := 0
	N := len(args)

	for i < N {
		arg := args[i]
		switch arg {
		case "":
			//skip
		case "--readBlk", "-r":
			cli.assertMoreArgs(i, N)
			block := node.ReadTmpBlock(args[i+1])
			node.SetBlock(block)
			i += 2
		case "--writeBlk", "-w":
			cli.getBlockFromArg(&i, args, node)
			node.WriteTmpBlock()
		case "--createBlk", "-c":
			node.CreateBlock(make([]byte, 0))
			i++
		case "--mineBlk", "-m":
			cli.getBlockFromArg(&i, args, node)
			node.Mine()
		case "--validateBlk", "-v":
			cli.getBlockFromArg(&i, args, node)
			node.ValidateBlock()
		case "--addBlk", "-d":
			cli.getBlockFromArg(&i, args, node)
			node.AddBlock()
		case "--updateUtxoSet", "-u":
			cli.getBlockFromArg(&i, args, node)
			node.UpdateUtxoSet()
		case "--broadcastBlk", "-b":
			cli.getBlockFromArg(&i, args, node)
			node.Broadcast()
		case "--validateTx", "-t":
			cli.getTxFromArg(&i, args, node)
			node.ValidateTx()
		case "--addTxToPool", "-o":
			cli.getTxFromArg(&i, args, node)
			node.AddTxToPool()
		case "--addTxToBlk", "-k":
			cli.getTxFromArg(&i, args, node)
			node.AddTxToBlock()
		default:
			cli.PrintUsage()
			os.Exit(1)
		}
	}
}

func (cli *CommandLine) Interactive() {

}

func (cli *CommandLine) Wallet() {
	name := ""
	args := make([]string, len(os.Args)-2)
	copy(args, os.Args[2:])

	i := 0
	N := len(args)
	for i < N {
		arg := args[i]
		switch arg {
		case "--name", "-n":
			name = args[i+1]
			args[i] = ""
			args[i+1] = ""
			i += 2
		default:
			i++
		}
		if name != "" {
			break
		}
	}
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

	i = 0
	N = len(args)
	for i < N {
		arg := args[i]
		switch arg {
		case "":
			i++
		case "--balance", "-b":
			fmt.Printf("Address: %s\nBalance: %d TRM.", w.ClientId.Address, wc.Balance())
			i++
		case "--getAddr", "-a":
			fmt.Printf("Wallet: %s\nAddress: %s", w.Name, w.ClientId.Address)
			i++
		case "--genTx", "-g":
			os.Exit(1)
		case "--sendTx", "-s":
			os.Exit(1)
		default:
			cli.PrintUsage()
			os.Exit(1)
		}
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
	return a, v
}

func (cli *CommandLine) Genesis() {
	if len(os.Args) < 3 {
		cli.PrintUsage()
		os.Exit(1)
	}
	nodeAddr := os.Args[2]

	_numTx := uint8(10)
	_port := uint16(t_config.RpcEndpointPort)
	nodeConf := t_config.Config{
		NumTxInBlock:    &_numTx,
		RpcEndpointPort: &_port,
		ClientAddress:   &nodeAddr,
	}

	node := node.NewNode(cli.ctx, &nodeConf)
	node.Genesis()
}

func (cli *CommandLine) assertNodeBlock(node *node.Node) {
	if node.GetBlock() == nil {
		fmt.Println("Node has no block.")
		os.Exit(1)
	}
}

func (cli *CommandLine) assertNodeTx(node *node.Node) {
	if node.GetTx() == nil {
		fmt.Println("Node has no tx.")
		os.Exit(1)
	}
}

func (cli *CommandLine) extractConf() (*t_config.Config, []string) {
	if len(os.Args) < 3 {
		cli.PrintUsage()
		os.Exit(1)
	}

	var numTx uint8 = t_config.NumTxInBlock
	nodeAddr := ""
	var port uint16 = t_config.RpcEndpointPort

	args := make([]string, len(os.Args)-2)
	copy(args, os.Args[2:])

	i := 0
	N := len(args)

	for i < N {
		arg := args[i]
		switch arg {
		case "--numTx", "-n":
			cli.assertMoreArgs(i, N)
			_numTx, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				cli.PrintUsage()
				os.Exit(1)
			} else {
				numTx = uint8(_numTx)
				args[i] = ""
				args[i+1] = ""
				i += 2
			}
		case "--nodeAddr", "-a":
			cli.assertMoreArgs(i, N)
			nodeAddr = os.Args[i+1]
			args[i] = ""
			args[i+1] = ""
			i += 2
		case "--port", "-p":
			cli.assertMoreArgs(i, N)
			_port, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				cli.PrintUsage()
				os.Exit(1)
			} else {
				port = uint16(_port)
				args[i] = ""
				args[i+1] = ""
				i += 2
			}
		default:
			i++
		}
	}
	nodeConf := t_config.Config{
		NumTxInBlock:    &numTx,
		RpcEndpointPort: &port,
		ClientAddress:   &nodeAddr,
	}
	return &nodeConf, args
}

func (cli *CommandLine) assertMoreArgs(argI, N int) {
	if argI == N {
		cli.PrintUsage()
		os.Exit(1)
	}
}

func (cli *CommandLine) getBlockFromArg(i *int, args []string, node *node.Node) {
	if node.GetBlock() == nil {
		node.SetBlock(node.ReadTmpBlock(args[*i+1]))
		(*i) += 2
	} else {
		(*i)++
	}
}

func (cli *CommandLine) getTxFromArg(i *int, args []string, node *node.Node) {
	if node.GetTx() == nil {
		node.SetTx(node.ReadTmpTx(args[*i+1]))
		(*i) += 2
	} else {
		(*i)++
	}
}
