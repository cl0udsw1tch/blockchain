package validator

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/blockStore"
	"github.com/terium-project/terium/internal/blockchain"
	"github.com/terium-project/terium/internal/mempool"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/utxoSet"
)

type TxValidator struct {
	ctx        *t_config.Context
	blockchain *blockchain.Blockchain
	tx         *transaction.Tx
	txStore 	*transaction.TxIndexIO
	blockStore	*blockStore.BlockStore
	mempool 	*mempool.MempoolIO
	utxoStore 	*utxoSet.UtxoStore
}

func NewTxValidator(
	ctx *t_config.Context, 
	_txStore *transaction.TxIndexIO, 
	_blockStore *blockStore.BlockStore, 
	_mempool *mempool.MempoolIO,
	_utxoStore 	*utxoSet.UtxoStore,
	) *TxValidator {
	v := new(TxValidator)
	v.ctx = ctx
	v.blockchain = blockchain.NewBlockchain(ctx, _blockStore)
	v.txStore = _txStore
	v.blockStore = _blockStore
	v.mempool = _mempool
	v.utxoStore = _utxoStore
	return v
}

func (v *TxValidator) ValidateTx(tx *transaction.Tx) bool {
	v.tx = tx
	if !v.assertNonEmpty() ||
		!v.assertNoCoinbases() ||
		!v.assertVal() ||
		!v.assertSpentCoinbaseMaturity() ||
		!v.assertSigScriptSyntax() ||
		!v.assertTxNotInPool() ||
		!v.assertTxInUTXOs() ||
		!v.validate() {
		return false
	}

	return true
}

func (v *TxValidator) assertNonEmpty() bool {
	return len(v.tx.Inputs) > 0 && len(v.tx.Outputs) > 0
}

func (v *TxValidator) assertNoCoinbases() bool {
	for _, in := range v.tx.Inputs {
		if in.PrevOutpt.Idx == -1 ||
			bytes.Equal(in.PrevOutpt.TxId, make([]byte, 32)) {
			return false
		}
	}
	return true
}

func (v *TxValidator) assertVal() bool {
	var sumIn int64 = 0
	var sumOut int64 = 0
	var inVal int64
	var outVal int64
	for i, tx_in := range v.tx.Inputs {
		utxo, err := v.blockchain.FindUTXO(&tx_in.PrevOutpt)
		if err != nil {
			return false
		}
		inVal = utxo.Value
		outVal = v.tx.Outputs[i].Value
		if outVal > inVal {
			return false
		}
		sumIn += inVal
		sumOut += outVal
	}
	return sumOut <= sumIn-int64(blockchain.TX_FEE)
}

func (v *TxValidator) assertSpentCoinbaseMaturity() bool {

	for _, in := range v.tx.Inputs {

		txMeta := v.txStore.Read(in.PrevOutpt.TxId)
		block, _ := v.blockStore.Read(txMeta.BlockHash)

		prevTx := block.Transactions[txMeta.Index]

		if (&prevTx).IsCoinbase() {
			h := v.blockchain.Height()
			if h.Cmp(new(big.Int).Add(&txMeta.BlockHeight, big.NewInt(int64(t_config.COINBASE_MATURITY)))) == -1 {
				return false
			}
		}
	}
	return true
}

func (v *TxValidator) assertSigScriptSyntax() bool {
	for _, in := range v.tx.Inputs {
		if len(in.UnlockingScript)%3 != 0 { // push, sz, bytes ...
			return false
		}
		i := 0
		for i < len(in.UnlockingScript) {
			code := in.UnlockingScript[i]
			if len(code) != 1 {
				return false
			}
			_code := transaction.OpCode(in.UnlockingScript[i][0])
			nSz, ok := transaction.OpPushMap[_code]
			if !ok {
				return false
			}
			if len(in.UnlockingScript[i+1]) != nSz {
				return false
			}
			sz := binary.BigEndian.Uint32(in.UnlockingScript[i+1])
			b := in.UnlockingScript[i+2]
			if len(b) != int(sz) {
				return false
			}
			i += 3
		}
	}
	return true
}

func (v *TxValidator) assertTxNotInPool() bool {
	_, _, ok := v.mempool.Read(v.tx.Hash())
	return !ok
}

func (v *TxValidator) assertTxInUTXOs() bool {
	for _, in := range v.tx.Inputs {
		outpt := in.PrevOutpt
		_, ok := v.utxoStore.Read(&outpt)
		if !ok {
			return false
		}
	}
	return true
}

func (v *TxValidator) validate() bool {

	for i, in := range v.tx.Inputs {
		outpt := in.PrevOutpt
		utxo, ok := v.utxoStore.Read(&outpt)
		if !ok {
			panic("Utxo not found")
		}
		opctx := transaction.OpCtx{
			Tx:        v.tx,
			Stack:     transaction.OpStack{},
			State:     transaction.OP_OK,
			TxIn:      &in,
			InUtxo:    utxo,
			InIdx:     uint8(i),
			Script:    append(in.UnlockingScript, utxo.LockingScript...),
			ScriptPtr: 0,
		}
		interpreter := transaction.NewInterpreter(&opctx)
		if interpreter.Execute() != transaction.OP_OK {
			return false
		}
	}
	return false
}
