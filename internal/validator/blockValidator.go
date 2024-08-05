package validator

import (
	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/t_config"
)

type BlockValidator struct {
	ctx        *t_config.Context
	blockchain *blockchain.Blockchain
	block      *block.Block
}

func NewBlockValidator(ctx *t_config.Context) *BlockValidator {
	b := new(BlockValidator)
	b.ctx = ctx
	b.blockchain = blockchain.NewBlockchain(ctx)
	return b
}

func (validator *BlockValidator) Validate(block *block.Block) bool {
	validator.block = block

	return true
}
