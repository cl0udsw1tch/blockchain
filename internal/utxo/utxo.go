package utxo

import (
	"github.com/terium-project/terium/internal/transaction"
)

type Utxo struct {
	TXID              [32]byte
	TXPoint           transaction.OutPoint
	Value             int64
	LockingScriptSize int8
	LockingScript     []byte
}

func (utxo *Utxo) getUtxo() (*Utxo, error) {

	var err error

	return nil, err
}
