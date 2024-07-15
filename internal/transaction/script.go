package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"github.com/terium-project/terium/internal/t_util"
	"golang.org/x/crypto/ripemd160"
)

type OpCode byte
type OpState byte
type OpFunc func(ctx *OpCtx)
type OpCtx struct {
	Tx     Tx
	Stack  OpStack
	State  OpState
	TxIn   TxIn
	InUtxo Utxo
	InIdx  uint8
}

type OpMap_T map[OpCode]OpFunc

const (
	OP_OK          OpState = 0x00
	OP_PANIC       OpState = 0x01
	OP_DUP         OpCode  = 0x02
	OP_HASH160     OpCode  = 0x03
	OP_EQUALVERIFY OpCode  = 0x04
	OP_EQUAL       OpCode  = 0x05
	OP_VERIFY      OpCode  = 0x06
	OP_CHECKSIG    OpCode  = 0x07
	OP_PUSHDATA1	OpCode = 0x08
	OP_PUSHDATA2	OpCode = 0x09
	OP_PUSHDATA4	OpCode = 0x0A
)

type SigHashFlag byte

const (
	SIGHASH_ALL          SigHashFlag = 0x01
	SIGHASH_NONE         SigHashFlag = 0x02
	SIGHASH_SINGLE       SigHashFlag = 0x03
	SIGHASH_ANYONECANPAY SigHashFlag = 0x80
)

var (
	OpMap OpMap_T = OpMap_T{
		OP_DUP:         OpDup,
		OP_HASH160:     OpHash160,
		OP_EQUALVERIFY: OpEqualVerify,
		OP_EQUAL:       OpEqual,
		OP_VERIFY:      OpVerify,
		OP_CHECKSIG:    OpCheckSig,
	}

	OpDup OpFunc = func(ctx *OpCtx) {
		ctx.Stack.Push(ctx.Stack.Peek())
	}

	OpHash160 OpFunc = func(ctx *OpCtx) {
		sha256Hasher := sha256.New()
		ripemd160Hasher := ripemd160.New()
		sha256Hasher.Write(ctx.Stack.Peek())
		ripemd160Hasher.Write(sha256Hasher.Sum(nil))
		ctx.Stack.Push(ripemd160Hasher.Sum(nil))
	}

	OpEqualVerify OpFunc = func(ctx *OpCtx) {
		OpEqual(ctx)
		OpVerify(ctx)
	}

	OpEqual OpFunc = func(ctx *OpCtx) {
		first, second := ctx.Stack.Pop(), ctx.Stack.Pop()

		if t_util.SliceCompare(first, second) {
			ctx.Stack.Push([]byte{0x01})
		} else {
			ctx.Stack.Push([]byte{0x00})
		}
	}

	OpVerify OpFunc = func(ctx *OpCtx) {
		top := ctx.Stack.Peek()
		if top[0] == byte(0x00) {
			ctx.Stack.Pop()
		} else {
			ctx.State = OP_PANIC
		}
	}

	OpCheckSig OpFunc = func(ctx *OpCtx) {
		pubKey := ctx.Stack.Pop()
		sig := ctx.Stack.Pop()

		sigHashFlag := uint8(binary.BigEndian.Uint16(sig[len(sig)-8:]))
		sig = sig[:len(sig)-8]

		txCopy := ctx.Tx.Copy()

		for _, tx_in := range txCopy.Inputs {
			tx_in.UnlockingScript = [][]byte{{0x00}}
			tx_in.UnlockingScriptSize = CompactSize{Type: COMPACT_SZ1, Size: []byte{0x00}}
		}
		txCopy.Inputs[ctx.InIdx].UnlockingScriptSize = ctx.InUtxo.LockingScriptSize
		txCopy.Inputs[ctx.InIdx].UnlockingScript = ctx.InUtxo.LockingScript

		switch SigHashFlag(sigHashFlag & 0b11) {
		case SIGHASH_ALL:
			break
		case SIGHASH_NONE:
			txCopy.Outputs = []TxOut{}
		case SIGHASH_SINGLE:
			txCopy.Outputs = txCopy.Outputs[:ctx.InIdx+1]
			for _, tx_out := range txCopy.Outputs[:ctx.InIdx] {
				tx_out.Value = -1
				tx_out.LockingScriptSize = CompactSize{Type: COMPACT_SZ1, Size: []byte{0x00}}
				tx_out.LockingScript = [][]byte{}
			}
		}

		if sigHashFlag&uint8(SIGHASH_ANYONECANPAY) != 0 {
			txCopy.Inputs = []TxIn{txCopy.Inputs[ctx.InIdx]}
		}

		var txBuffer bytes.Buffer
		encoder := gob.NewEncoder(&txBuffer)
		encoder.Encode(txCopy)

		txBuffer.Write([]byte{sigHashFlag})

		var pubKeyObj ecdsa.PublicKey
		var bufferPK bytes.Buffer
		bufferPK.Write(pubKey)
		decoder := gob.NewDecoder(&bufferPK)
		decoder.Decode(&pubKeyObj)

		valid := ecdsa.VerifyASN1(&pubKeyObj, txBuffer.Bytes(), sig)

		if valid {
			ctx.State = OP_OK
			return
		}
		ctx.State = OP_PANIC

	}
)
