package transaction

import (
	"bytes"
	"encoding/binary"

	"github.com/tiereum/trmnode/internal/t_util"
)

type OutPoint struct {
	TxId []byte
	Idx  int32
}

func (pt OutPoint) Copy() OutPoint {
	var _txId []byte
	copy(_txId[:], pt.TxId[:])
	_pt := OutPoint{
		TxId: _txId,
		Idx:  pt.Idx,
	}
	return _pt
}

const (
	COMPACT_SZ1 byte = 0x01
	COMPACT_SZ2 byte = 0x02
	COMPACT_SZ3 byte = 0x03
)

type CompactSize struct {
	Type byte
	Size []byte
}

func NewCompactSize(val int64) CompactSize {
	buffer := bytes.Buffer{}
	binary.Write(&buffer, binary.BigEndian, val)
	sz := 0
	bts := buffer.Bytes()
	for sz < len(bts) && bts[len(buffer.Bytes())-1-sz] != 0x00 {
		sz++
	}
	if sz == 0 {
		return CompactSize{
			Type: byte(1),
			Size: []byte{0},
		}
	}
	return CompactSize{
		Type: byte(sz),
		Size: bts[len(bts)-sz:],
	}
}

type TxOut struct {
	Value             int64
	LockingScriptSize CompactSize
	LockingScript     []byte
}

func (s TxOut) Copy() TxOut {
	lockscript := bytes.Clone(s.LockingScript)
	o := TxOut{
		Value:             s.Value,
		LockingScriptSize: s.LockingScriptSize,
		LockingScript:     lockscript,
	}
	return o
}

type TxOuts []TxOut

func (s TxOuts) Copy() TxOuts {
	var o TxOuts = make(TxOuts, len(s))
	for i, it := range s {
		o[i] = it.Copy()
	}
	return o
}

type TxIn struct {
	PrevOutpt           OutPoint
	UnlockingScriptSize CompactSize
	UnlockingScript     []byte
}

type TxIns []TxIn

func (s TxIn) Copy() TxIn {
	unlockscript := bytes.Clone(s.UnlockingScript)
	o := TxIn{
		PrevOutpt:           s.PrevOutpt.Copy(),
		UnlockingScriptSize: s.UnlockingScriptSize,
		UnlockingScript:     unlockscript,
	}
	return o
}

func (s TxIns) Copy() TxIns {
	var o TxIns = make(TxIns, len(s))
	for i, it := range s {
		o[i] = it.Copy()
	}
	return o
}

type Tx struct {
	Version    int32
	NumInputs  uint8
	Inputs     TxIns
	NumOutputs uint8
	Outputs    TxOuts
	LockTime   uint32 // number of blocks until spending is allowed
}
type Utxo struct {
	OutPoint          OutPoint
	Value             int64
	LockingScriptSize CompactSize
	LockingScript     []byte
}

func (s *Tx) Copy() Tx {

	o := Tx{
		Version:    s.Version,
		NumInputs:  s.NumInputs,
		Inputs:     s.Inputs.Copy(),
		NumOutputs: s.NumOutputs,
		Outputs:    s.Outputs.Copy(),
		LockTime:   s.LockTime,
	}
	return o
}

func (tx *Tx) Serialize() []byte {
	e := NewTxEncoder(nil)
	e.Encode(tx)
	return e.Bytes()
}

func (tx *Tx) Hash() []byte {
	return t_util.Hash256(tx.Serialize())
}

func Coinbase(scriptSz CompactSize, script []byte) TxIn {
	return TxIn{
		PrevOutpt:           OutPoint{TxId: make([]byte, 32), Idx: -1},
		UnlockingScriptSize: scriptSz,
		UnlockingScript:     script,
	}
}

func (tx *Tx) IsCoinbase() bool {
	return len(tx.Inputs) == 1 &&
		len(tx.Outputs) == 1 &&
		tx.Inputs[0].PrevOutpt.Idx == -1 &&
		bytes.Equal(tx.Inputs[0].PrevOutpt.TxId, make([]byte, 32))
}

func (tx *Tx) HasCoinbases() bool {
	for _, in := range tx.Inputs {
		if in.PrevOutpt.Idx == -1 ||
			bytes.Equal(in.PrevOutpt.TxId, make([]byte, 32)) {
			return true
		}
	}
	return false
}

func (tx *Tx) Preimage(inIdx uint8, inUTXO *Utxo, sigHashFlag byte) []byte {
	txCopy := tx.Copy()

	for _, tx_in := range txCopy.Inputs {
		tx_in.UnlockingScript = []byte{0x00}
		tx_in.UnlockingScriptSize = NewCompactSize(1)
	}
	txCopy.Inputs[inIdx].UnlockingScriptSize = inUTXO.LockingScriptSize
	txCopy.Inputs[inIdx].UnlockingScript = inUTXO.LockingScript

	switch SigHashFlag(sigHashFlag & 0b11) {
	case SIGHASH_ALL:
		break
	case SIGHASH_NONE:
		txCopy.Outputs = []TxOut{}
	case SIGHASH_SINGLE:
		txCopy.Outputs = txCopy.Outputs[:inIdx+1]
		for _, tx_out := range txCopy.Outputs[:inIdx] {
			tx_out.Value = -1
			tx_out.LockingScriptSize = NewCompactSize(1)
			tx_out.LockingScript = []byte{0x00}
		}
	}

	if sigHashFlag&uint8(SIGHASH_ANYONECANPAY) != 0 {
		txCopy.Inputs = []TxIn{txCopy.Inputs[inIdx]}
	}

	var txBuffer bytes.Buffer
	encoder := NewTxEncoder(&txBuffer)
	encoder.Encode(&txCopy)
	txBuffer.Write([]byte{sigHashFlag})
	return t_util.Hash256(txBuffer.Bytes())
}
