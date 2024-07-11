package proof

import (
	"encoding/binary"
	"math"
	"math/big"

	"github.com/terium-project/terium/internal/block"
)

type NoNonceError struct {}

func (n NoNonceError) Error() string {
	return "Nonce not found."
}

type PoWState struct {
	Hash []byte
	Nonce uint32
	Solved bool
} 

type PoWStateNotifier chan PoWState


type PoW struct {
	block *block.Block
	notifier PoWStateNotifier
}

func NewPoW(block *block.Block) PoW {
	return PoW{block: block, notifier: make(chan PoWState)}
}

func (pow PoW) Solve() error {

	var hash big.Int
	state := PoWState{Nonce: 0, Solved: false}

	for state.Nonce < (1<<32 - 1) {
		pow.block.Header.Nonce = state.Nonce
		hash.SetBytes(pow.block.Hash())

		state.Hash = hash.Bytes()

		if pow.verifyHash(hash) {
			state.Solved = true
			pow.notifier <- state
			return nil
		}
		pow.notifier <- state
		state.Nonce++
	}

	return NoNonceError{}
}

func (pow PoW) Validate() bool {

	var hash big.Int
	hash.SetBytes(pow.block.Hash())
	pow.notifier<-PoWState{Hash: hash.Bytes()}
	return pow.verifyHash(hash)
} 

func (pow PoW) verifyHash(hash big.Int) bool {
	return pow.block.Header.Target.Cmp(&hash) == 1
}