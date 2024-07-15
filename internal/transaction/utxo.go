package transaction

import (
)

type Utxo struct {
	TXID              [32]byte
	TXPoint           OutPoint
	Value             int64
	LockingScriptSize CompactSize
	LockingScript     [][]byte
}

func (utxo *Utxo) getUtxo() (*Utxo, error) {

	var err error

	return nil, err
}
