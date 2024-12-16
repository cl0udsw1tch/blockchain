package cli

import (
	"fmt"
	"os"

	"github.com/tiereum/trmclient/internal/t_config"
	"github.com/tiereum/trmclient/internal/wallet"
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

	fmt.Println("\nwallet")
	fmt.Printf("%-20s%-30s%s", "--name", "<name>", "Name of wallet to use, creates one if it doesnt exist\n")
	fmt.Printf("%-50s%s", "--balance", "Print balance\n")
	fmt.Printf("%-20s%-30s%s", "--tx", "<address:value,...>", "Create a transaction. arg is hex encoded address:value,...\n")
	fmt.Printf("%-20s%-30s%s", "--sendTx", "[tx_hash]", "Broadcast transaction. Arg is the txid.\n")

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
	case "wallet":
		cli.Wallet()
	default:
		cli.PrintUsage()
		os.Exit(1)
	}

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
			// generate tx
			os.Exit(1)
		case "--sendTx", "-s":
			os.Exit(1)
		default:
			cli.PrintUsage()
			os.Exit(1)
		}
	}
}