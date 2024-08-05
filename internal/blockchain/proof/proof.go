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
	block *block.Block
	notifier PoWStateNotifier
}

func NewPoW(block *block.Block) PoW {
	return PoW{block: block, notifier: make(chan PoWState)}
}

func (pow PoW) Solve(quitSig <-chan byte, ackChan chan<- byte) {

	var hash big.Int
	state := PoWState{Nonce: 0, Solved: false}

	for state.Nonce < (1<<32 - 1) {
		select {
		case <-quitSig :
			ackChan<-0x00
			return
		default : 
			pow.block.Header.Nonce = state.Nonce
			hash.SetBytes(pow.block.Hash())

			state.Hash = hash.Bytes()

			if pow.verifyHash(hash) {
				state.Solved = true
				pow.notifier <- state
				return
			}
			pow.notifier <- state
			state.Nonce++
		}
	}
	
	t_error.LogErr(NoNonceError{})
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