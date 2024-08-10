package main

import (
	"github.com/terium-project/terium/cmd/cli"
	"github.com/terium-project/terium/internal/t_config"
)

func main() {

	ctx := t_config.NewContext()
	cli := cli.NewCommandLine(ctx)
	cli.ValidateArgs()
}
