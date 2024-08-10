package cli

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockStore"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/mempool"
	"github.com/terium-project/terium/internal/miner"
	"github.com/terium-project/terium/internal/node"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/wallet"
)

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
	fmt.Println("command [(--arg_name | -arg_flag) arg_value] ...")

	fmt.Println("run")
	fmt.Printf("%-20s%-30s%s", "--numTx", "| -n <int>", "Number of txs to include before mining current block. Default 10\n")
	fmt.Printf("%-20s%-30s%s", "--port", "| -p <int>", "Port to listen and send on. Default "+fmt.Sprint(t_config.RpcEndpointPort)+"\n")
	fmt.Printf("%-20s%-30s%s", "--nodeAddr", "| -a", "Address for block rewards\n")
	fmt.Printf("%-20s%-30s%s", "--interactive", "| -i", "Run in interactive mode. Default false\n")
		fmt.Printf("%-20s%-30s%s", "--createBlock", "| -c", "Creates block\n")
		fmt.Printf("%-20s%-30s%s", "--mineBlock", "| -n <block_hash>", "Mines block\n")
		fmt.Printf("%-20s%-30s%s", "--validateBlock", "| -v <block_hash>", "Validates block\n")
		fmt.Printf("%-20s%-30s%s", "--addBlock", "| -a <block_hash>", "Adds a mined and validated block to the local blockchain\n")
		fmt.Printf("%-20s%-30s%s", "--validateTx", "| -e <tx_hash>", "Validates a transaction\n")
		fmt.Printf("%-20s%-30s%s", "--addTxToMem", "| -m <tx_hash>", "Adds tx to mempool. Validate first\n")
		fmt.Printf("%-20s%-30s%s", "--addTx", "| -t <tx_hash>", "Adds a validated transaction to block\n")

	fmt.Println("blockchain")
	fmt.Printf("%-20s%-30s%s", "--print", "| -p", "Print the block header hashes of the main branch\n")
	fmt.Printf("%-20s%-30s%s", "--utxo", "| -u", "Print the utxo outpoints in the utxo set\n")
	fmt.Printf("%-20s%-30s%s", "--mempool", "| -m", "Print the tx hashes in the mempool\n")

	fmt.Println("wallet")
	fmt.Printf("%-20s%-30s%s", "--name", "| -n <string>", "name of wallet to use, creates one if it doesnt exist\n")
	fmt.Printf("%-20s%-30s%s", "--balance", "| -b", "print balance\n")
	fmt.Printf("%-20s%-30s%s", "--tx", "| -t <address:value,...>", "create a transaction. arg is hex encoded address:value,...\n")
	fmt.Printf("%-20s%-30s%s", "--sendTx", "| -s <tx_hash>", "broadcast transaction. Arg is the txid.\n")

	fmt.Println("update")
	fmt.Printf("%-20s%-30s%s", "--print", "| -p", "Print the block header hashes of the main branch\n")

	os.Exit(1)
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) == 1 {
		cli.PrintUsage()
	}
	if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			cli.PrintUsage()
		}
	}
	switch os.Args[1] {
	case "run":
		cli.Run()
	case "blockchain":
		cli.Blockchain()
	case "wallet":
		cli.Wallet()
	case "update":
		cli.Update()
	case "genesis":
		cli.Genesis()
	default : 
		cli.PrintUsage()
		os.Exit(1)
	}
	
}

func (cli *CommandLine) Run() {
	var numTx int
	var port int
	var addr string
	var interactive bool
	flag.IntVar(&numTx, "numTx", 10, "")
	flag.StringVar(&addr, "nodeAddr", "", "")
	flag.IntVar(&port, "port", int(t_config.RpcEndpointPort), "")
	flag.BoolVar(&interactive, "interactive", false, "")
	flag.CommandLine.Parse(os.Args[2:])


	if len(flag.Args()) != 1 {
		cli.PrintUsage()
	}
	_numTx := uint8(numTx)
	_port := uint16(port)
	nodeConf := t_config.Config{
		NumTxInBlock: &_numTx,
		RpcEndpointPort: &_port,
		ClientAddress: &addr,

	}
	node := node.NewNode(cli.ctx, &nodeConf)
	if !interactive {
		node.Run()
	} else {
		go node.StartInteractive()
		scanner := bufio.NewScanner(os.Stdin)
		outerCommandLoop:
		for {
			fmt.Print("tierium: ")
			if !scanner.Scan() {
				break
			}
			input := scanner.Text()
			if input == "exit" {
				fmt.Println("Exiting...")
				break
			}
			args := strings.Split(input, " ")
			argN := 0
			for argN < len(args) {
				switch args[argN] {
				case "--createBlock", "-c" :
					node.CreateBlock(make([]byte, 0))
				case "--mineBlock", "-n" : 
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.Mine(cli.GetBlock(arg))
					} else if arg == "node" {
						node.Mine(nil)
					}
				case "--validateBlock", "-v":
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.ValidateBlock(cli.GetBlock(arg))
					} else if arg == "node" {
						node.ValidateBlock(nil)
					}
				case "--addBlock", "-a":
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.AddBlock(cli.GetBlock(arg))
					} else if arg == "node" {
						node.AddBlock(nil)
					}
				case "--broadcast", "-b":
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.Broadcast(cli.GetBlock(arg))
					} else if arg == "node" {
						node.Broadcast(nil)
					}
				case "--validateTx", "-e":
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.ValidateTx(cli.GetTx(arg))
					}
				case "--addTxToMem", "-m":
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.AddTxToPool(cli.GetTx(arg))
					}
				case "--addTx", "-t": 
					if argN+1==len(args) {
						cli.PrintUsage()
						break outerCommandLoop
					}
					arg := args[argN+1]
					if len(arg) == 64 {
						node.AddTxToBlock(cli.GetTx(arg))
					}
				}
				argN++
			}
		}
	}
}
func (cli *CommandLine) GetBlock(hexHash string) *block.Block {
	p := path.Join(cli.ctx.TmpDir, "blocks", hexHash)
	b, err := os.ReadFile(p)
	t_error.LogErr(err)
	buffer := new(bytes.Buffer)
	buffer.Write(b)
	dec := block.NewBlockDecoder(nil)
	err = dec.Decode(buffer)
	t_error.LogErr(err)
	return dec.Out()
}
func (cli *CommandLine) GetTx(hexHash string) *transaction.Tx {
	p := path.Join(cli.ctx.TmpDir, "txs", hexHash)
	b, err := os.ReadFile(p)
	t_error.LogErr(err)
	buffer := new(bytes.Buffer)
	buffer.Write(b)
	dec := transaction.NewTxDecoder(nil)
	err = dec.Decode(buffer)
	t_error.LogErr(err)
	return dec.Out()
}


func (cli *CommandLine) Blockchain() {

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
		t_error.LogErr(errors.New("name of wallet required"))
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
func (cli *CommandLine) Update() {

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
	
	_blockStore := blockStore.NewBlockStore(cli.ctx)
	_blockchain := blockchain.NewBlockchain(cli.ctx, _blockStore)
	_mempool := mempool.NewMempoolIO(cli.ctx)
	miner := miner.NewMiner(cli.ctx, _blockchain, _mempool)
	miner.Genesis()
}