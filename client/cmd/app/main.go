package main

import (
	"github.com/tiereum/trmclient/cmd/cli"
	"github.com/tiereum/trmclient/internal/t_config"
)


func main() {
	ctx := t_config.NewContext()
	cli := cli.NewCommandLine(ctx)
	cli.ValidateArgs()
}


