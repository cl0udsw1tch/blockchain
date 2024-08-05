package main

import (
	"github.com/terium-project/terium/cmd/cli"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
)

func main() {

	ctx := t_config.NewContext()

	cli := cli.NewCommandLine(ctx)

}
