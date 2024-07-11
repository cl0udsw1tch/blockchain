package txMsg

import (
	"encoding/binary"

	"github.com/terium-project/terium/internal/blockchain"
	. "github.com/terium-project/terium/internal/transaction"
)

type Validator struct {
	Blockchain blockchain.Blockchain
	Uid        []byte
	Tx         Tx
}

func New(b blockchain.Blockchain, uid []byte, tx Tx) Validator {

	return Validator{
		Blockchain: b,
		Uid:        uid,
	}

}

func (v Validator) SetTx(tx Tx) {
	v.Tx = tx
}

func (v Validator) Validate() bool {

	v.checkNonEmpty()
	v.checkVal()
	v.checkNotCoinbase()

	return true
}

func (v Validator) checkNonEmpty() bool {
	return len(v.Tx.Inputs) > 0 && len(v.Tx.Outputs) > 0
}

func (v Validator) checkVal() bool {
	var sumIn int64 = 0
	var sumOut int64 = 0
	var inVal int64
	var outVal int64
	for i, tx_in := range v.Tx.Inputs {
		txOut, err := v.Blockchain.FindTxOutByOutPt(tx_in.PrevOutpt)
		if err != nil {
			return false
		}
		inVal = txOut.Value
		outVal = v.Tx.Outputs[i].Value
		if outVal > inVal {
			return false
		}
		sumIn += inVal
		sumOut += outVal
	}
	if sumOut > sumIn-int64(blockchain.TX_FEE) {
		return false
	}

	return true
}

func (v Validator) checkNotCoinbase() bool {

	for _, tx_in := range v.Tx.Inputs {
		if binary.BigEndian.Uint32(tx_in.PrevOutpt.TxId[:]) == 0 && tx_in.PrevOutpt.Idx == 0xFFFFFFFF {
			return false
		}
	}
}

func (v Validator) checkScriptSyntax() bool {
	for i, tx_in := range v.Tx.Inputs {

	}

	for i, tx_out := range v.Tx.Inputs {

	}

	return true
}

func (v Validator) checkVal() bool {

}

func (v Validator) checkVal() bool {

}
