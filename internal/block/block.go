package block

import (
	"math/big"
	"github.com/terium-project/terium/internal/t_util"
	"github.com/terium-project/terium/internal/transaction"
)

// 80 bytes
type Header struct{
	Version			int32
	PrevHash	 	[]byte // 32 bytes
	MerkleRootHash	[]byte // 32 bytes
	TimeStamp		uint32
	Target 			big.Int
	Nonce			uint32
}

type Block struct{
	Header 					Header
	TXCount					uint32
	Transactions			[]transaction.Tx
}

func MerkelRoot(nodes [][]byte) []byte {
	switch len(nodes){
	case 0:
		return []byte{}
	case 1: 
		b := t_util.Hash256(nodes[0])
		b = t_util.Hash256(append(b[:], b[:]...))
		return b[:]
	case 2: 
		b0 := t_util.Hash256(nodes[0])
		b1 := t_util.Hash256(nodes[1])
		b := t_util.Hash256(append(b0[:], b1[:]...))
		return b[:]
	default:
		l := MerkelRoot(nodes[:len(nodes) / 2])
		r := MerkelRoot(nodes[len(nodes) / 2:])
		b := t_util.Hash256(append(l[:], r[:]...))
		return b[:]
	}
}

func (block Block) MerkelRoot() []byte {
	bits := make([][]byte, len(block.Transactions))
	e := transaction.NewTxEncoder(nil)
	for i, tx := range block.Transactions {
		e.Encode(&tx)
		bits[i] = e.Bytes()
		e.Clear()
	}
	return MerkelRoot(bits)
}

func (block Block) Hash() []byte {
	e := NewHeaderEncoder(nil)
	e.Encode(&block.Header)
	return t_util.Hash256(e.Bytes()) 
}

func (block Block) Serialize() []byte {
	e := NewBlockEncoder(nil)
	e.Encode(&block)
	return e.Bytes()
}







	