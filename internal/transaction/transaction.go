package transaction

import (
	"github.com/terium-project/terium/internal/t_util"
)

type OutPoint struct {
	TxId []byte
	Idx  uint32
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


type TxOut struct {
	Value             int64
	LockingScriptSize CompactSize
	LockingScript     [][]byte
}

func (s TxOut) Copy() TxOut {
	lockscript := t_util.CopySlice2D(s.LockingScript)
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
	UnlockingScript     [][]byte
}

type TxIns []TxIn

func (s TxIn) Copy() TxIn {
	unlockscript := t_util.CopySlice2D(s.UnlockingScript)
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
	Version    uint32
	NumInputs  uint8
	Inputs     TxIns
	NumOutputs uint8
	Outputs    TxOuts
	LockTime   uint32 // number of blocks until spending is allowed
}

func (s Tx) Copy() Tx {

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

func (tx Tx) Serialize() []byte {
	e := TxEncoder{}
	e.Encode(tx)
	return e.Bytes()
}


func Coinbase(scriptSz CompactSize, script [][]byte) TxIn {
	return TxIn{
		PrevOutpt: OutPoint{TxId: []byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		}, Idx: 0xFFFFFFFF},
		UnlockingScriptSize: scriptSz,
		UnlockingScript: script,
	}
}

func CoinbaseTx(
	version uint32, 
	inScriptSz CompactSize, 
	inSript [][]byte,
	nOut uint8, 
	outputs TxOuts, ) Tx {

	return Tx{
		Version: version,
		NumInputs: 1,
		Inputs: []TxIn{Coinbase(inScriptSz, inSript)},
		NumOutputs: nOut,
		Outputs: outputs,
		LockTime: 100,
	}
}




