package validator

import (
	"encoding/hex"

	"github.com/tiereum/trmnode/internal/block"
	"github.com/tiereum/trmnode/internal/blockchain"
	"github.com/tiereum/trmnode/internal/blockchain/proof"
	"github.com/tiereum/trmnode/internal/t_config"
)

type BlockValidator struct {
	ctx         *t_config.Context
	blockchain  *blockchain.Blockchain
	txValidator *TxValidator
}

func NewBlockValidator(ctx *t_config.Context, blockchain *blockchain.Blockchain, txValidator *TxValidator) *BlockValidator {
	b := new(BlockValidator)
	b.ctx = ctx
	b.blockchain = blockchain
	b.txValidator = txValidator
	return b
}

func (validator *BlockValidator) Validate(block *block.Block) bool {
	if !validator.AssertNonEmpty(block) ||
		!validator.AssertNonce(block) ||
		!validator.AssertMerkelHash(block) ||
		!validator.AssertCoinbaseFirst(block) ||
		!validator.AssertValidTxs(block) {
		return false
	}
	return true
}

func (validator *BlockValidator) AssertNonEmpty(block *block.Block) bool {
	return len(block.Transactions) > 0
}

func (validator *BlockValidator) AssertNonce(block *block.Block) bool {
	pow := proof.NewPoW()
	return pow.Validate(block.Header.Target, block.Hash())
}
func (validator *BlockValidator) AssertMerkelHash(block *block.Block) bool {
	a_root := hex.EncodeToString(block.MerkelRoot())
	e_root := hex.EncodeToString(block.Header.MerkleRootHash)
	return a_root == e_root
}

func (validator *BlockValidator) AssertCoinbaseFirst(block *block.Block) bool {
	first := block.Transactions[0]
	return first.IsCoinbase()
}
func (validator *BlockValidator) AssertValidTxs(block *block.Block) bool {
	for _, tx := range block.Transactions {
		if !validator.txValidator.ValidateTx(&tx) {
			return false
		}
	}
	return true
}
