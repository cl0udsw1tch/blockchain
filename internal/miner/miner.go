package miner

import (
	"time"

	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/blockchain/proof"
)


func Mine(block *block.Block) error {
	block.Header.TimeStamp = uint32(time.Now().Unix())
	block.Header.MerkleRootHash = block.MerkelRoot()
	pow := proof.NewPoW(block)
	return pow.Solve()
}