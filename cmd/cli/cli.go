package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/terium-project/terium/internal/node"
	"github.com/terium-project/terium/internal/t_config"
)

type CommandLine struct{
	ctx *t_config.Context
	node *node.Node
}

func NewCommandLine(ctx *t_config.Context) *CommandLine {
	cli := new(CommandLine)
	cli.ctx = ctx
	cli.ValidateArgs()

	return cli
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage:")
	fmt.Println("command [(--arg_name | -arg_flag) arg_value] ...")

	fmt.Println("run")
	fmt.Printf("%-20s%s", "--numTx | -n arg", "Number of txs to include before mining current block. Default 10\n")
	fmt.Printf("%-20s%s", "--port | -p arg", "Port to listen and send on. Default "+fmt.Sprint(t_config.RpcEndpointPort)+"\n.")
	fmt.Printf("%-20s%s", "--nodeAddr | -a", "Address for block rewards\n")
	fmt.Printf("%-20s%s", "--interactive | -i", "Run in interactive mode. Default false\n")
		fmt.Printf("%-20s%s", "--createBlock | -c", "Creates block\n")
		fmt.Printf("%-20s%s", "--mineBlock | -n", "Mines block\n")
		fmt.Printf("%-20s%s", "--validateBlock | -v", "Validates block\n")
		fmt.Printf("%-20s%s", "--addBlock | -a", "Adds a mined and validated block to the local blockchain\n")
		fmt.Printf("%-20s%s", "--addTx | -t", "Adds a validated transaction\n")
		fmt.Printf("%-20s%s", "--validateTx | -e", "Validates a transaction\n")

	fmt.Println("blockchain")
	fmt.Printf("%-20s%s", "--print | -p", "Print the block header hashes of the main branch\n")
	fmt.Printf("%-20s%s", "--utxo | -u", "Print the utxo outpoints in the utxo set\n")
	fmt.Printf("%-20s%s", "--mempool | -m", "Print the tx hashes in the mempool\n")

	fmt.Println("wallet")
	fmt.Printf("%-20s%s", "--name | -n arg", "name of wallet to use, creates one if it doesnt exist\n")
	fmt.Printf("%-20s%s", "--balance | -b", "print balance\n")
	fmt.Printf("%-20s%s", "--tx | -t arg", "create a transaction. arg is hex encoded address:value,...\n")
	fmt.Printf("%-20s%s", "--sendTx | -s arg", "broadcast transaction. Arg is the txid.\n")

	fmt.Println("update")
	fmt.Printf("%-20s%s", "--print | -p", "Print the block header hashes of the main branch\n")
}

func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) == 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			cli.PrintUsage()
		}
	}
	switch os.Args[1] {
	case "run":
		var numTx int
		var port int
		var addr string
		var interactive bool
		flag.IntVar(&numTx, "numTx", 10, "")
		flag.StringVar(&addr, "nodeAddr", "", "")
		flag.IntVar(&port, "port", int(t_config.RpcEndpointPort), "")
		flag.BoolVar(&interactive, "interactive", false, "")
		flag.Parse()

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
		cli.Run(node, interactive)
		
	case "blockchain":
		cli.Blockchain()
	case "wallet":
		cli.Wallet()
	case "update":
		cli.Update()
	}
	addBlock := false
	flag.BoolVar(&addBlock, "addBlock", false, "Adds a mined and validated block to the local blockchain")
}

func (cli *CommandLine) Run(node *node.Node, interactive bool) {
	if !interactive {
		node.Run()
	} else {

	}
}
func (cli *CommandLine) Blockchain() {

}
func (cli *CommandLine) Wallet() {

}
func (cli *CommandLine) Update() {

}
