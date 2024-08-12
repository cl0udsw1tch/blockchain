package main

import (
	"github.com/tiereum/trmnode/cmd/cli"
	"github.com/tiereum/trmnode/internal/t_config"
)

func main() {

	ctx := t_config.NewContext()
	cli := cli.NewCommandLine(ctx)
	cli.ValidateArgs()
}
