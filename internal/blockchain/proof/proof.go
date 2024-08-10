package proof

import (
	"math/big"

	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/t_error"
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
	Notifier PoWStateNotifier
}

func NewPoW() *PoW {
	return &PoW{Notifier: make(chan PoWState)}
}

func (pow *PoW) Solve(block *block.Block, restart chan byte) bool {

	state := PoWState{Nonce: 0, Solved: false}

	for state.Nonce < (1<<32 - 1) {
		select {
		case <-restart :
			return false
		default : 
			block.Header.Nonce = state.Nonce
			state.Hash = block.Hash()

			if pow.Validate(block.Header.Target, state.Hash) {
				state.Solved = true
				pow.Notifier <- state
				return  true
			}
			pow.Notifier <- state
			state.Nonce++
		}
	}
	
	t_error.LogErr(NoNonceError{})
	return false
}

func (pow *PoW) Validate(target big.Int, hash []byte) bool {

	return target.Cmp(new(big.Int).SetBytes(hash)) == 1
}

func (pow *PoW) Close() {
	close(pow.Notifier)
}
