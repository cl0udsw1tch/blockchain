package transaction

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"

	"github.com/terium-project/terium/internal/client"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/t_util"
	"golang.org/x/crypto/ripemd160"
)

const P2PKH_LOCK_SCRIPT_SZ int64 = 26 

type OpCode byte
type OpState byte
type OpFunc func(ctx *OpCtx)
type OpCtx struct {
	Tx     		*Tx
	Stack  		OpStack
	State  		OpState
	TxIn   		*TxIn
	InUtxo 		*Utxo
	InIdx  		uint8
	Script 		[][]byte
	ScriptPtr	uint8
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

var OpPushMap map[OpCode]int = map[OpCode]int {
	OP_PUSHDATA1: 1,
	OP_PUSHDATA2: 2,
	OP_PUSHDATA4: 4,
}

type SigHashFlag byte

const (
	SIGHASH_ALL          SigHashFlag = 0x01
	SIGHASH_NONE         SigHashFlag = 0x02
	SIGHASH_SINGLE       SigHashFlag = 0x03
	SIGHASH_ANYONECANPAY SigHashFlag = 0x80
)

var (
	OpMap OpMap_T = OpMap_T{
		OP_PUSHDATA1:	OpPushData,
		OP_PUSHDATA2:	OpPushData,
		OP_PUSHDATA4:	OpPushData,
		OP_DUP:         OpDup,
		OP_HASH160:     OpHash160,
		OP_EQUALVERIFY: OpEqualVerify,
		OP_EQUAL:       OpEqual,
		OP_VERIFY:      OpVerify,
		OP_CHECKSIG:    OpCheckSig,
	}

	OpPushData OpFunc = func(ctx *OpCtx) {
		nSz := OpPushMap[OpCode(ctx.Script[ctx.ScriptPtr][0])]
		ctx.ScriptPtr++
		sz := binary.BigEndian.Uint32(ctx.Script[ctx.ScriptPtr][:nSz])
		ctx.ScriptPtr++
		ctx.Stack.Push(ctx.Script[ctx.ScriptPtr][:sz])
		ctx.ScriptPtr++
	}

	OpDup OpFunc = func(ctx *OpCtx) {
		ctx.Stack.Push(ctx.Stack.Peek())
		ctx.ScriptPtr++
	}

	OpHash160 OpFunc = func(ctx *OpCtx) {
		sha256Hasher := sha256.New()
		ripemd160Hasher := ripemd160.New()
		sha256Hasher.Write(ctx.Stack.Peek())
		ripemd160Hasher.Write(sha256Hasher.Sum(nil))
		ctx.Stack.Push(ripemd160Hasher.Sum(nil))
		ctx.ScriptPtr++
	}

	OpEqualVerify OpFunc = func(ctx *OpCtx) {
		OpEqual(ctx)
		OpVerify(ctx)
		ctx.ScriptPtr--
	}

	OpEqual OpFunc = func(ctx *OpCtx) {
		first, second := ctx.Stack.Pop(), ctx.Stack.Pop()

		if t_util.SliceCompare(first, second) {
			ctx.Stack.Push([]byte{0x01})
		} else {
			ctx.Stack.Push([]byte{0x00})
		}
		ctx.ScriptPtr++
	}

	OpVerify OpFunc = func(ctx *OpCtx) {
		top := ctx.Stack.Peek()
		if top[0] == byte(0x00) {
			ctx.Stack.Pop()
		} else {
			ctx.State = OP_PANIC
		}
		ctx.ScriptPtr++
	}

	OpCheckSig OpFunc = func(ctx *OpCtx) {
		pubKey := ctx.Stack.Pop()
		sig := ctx.Stack.Pop()

		sigHashFlag := uint8(binary.BigEndian.Uint16(sig[len(sig)-8:]))
		sig = sig[:len(sig)-8]

		preimage := ctx.Tx.Preimage(ctx.InIdx, ctx.InUtxo, sigHashFlag)

		pubKeyObj := client.UnMarshalPubKey(pubKey)
		valid := ecdsa.VerifyASN1(pubKeyObj, preimage, sig)

		if valid {
			ctx.State = OP_OK
			ctx.ScriptPtr++
			return
		}
		ctx.State = OP_PANIC
		ctx.ScriptPtr++

	}
)

type Interpreter struct {
	ctx *OpCtx
}

func NewInterpreter(ctx *OpCtx) *Interpreter {
	i := new(Interpreter)
	i.ctx = ctx
	return i
}

func (i *Interpreter) Execute() OpState {
	for i.ctx.ScriptPtr	< uint8(len(i.ctx.Script)) {
		OpMap[OpCode(i.ctx.Script[i.ctx.ScriptPtr][0])](i.ctx)
	}
	return OpState(i.ctx.Stack.Peek()[0])
}


func GetAddrFromP2PKHLockScript(script [][]byte) string {
	scriptStr := t_util.ConvertToHexString(script)


	pattern := fmt.Sprintf(`^%x%x%x%x([0-9a-fA-F]{40})%x%x$`, 
		OP_DUP, 
		OP_HASH160,
		OP_PUSHDATA1,
		0x14,
		OP_EQUALVERIFY,
		OP_CHECKSIG,
	)
	re := regexp.MustCompile(pattern)
	if !re.MatchString(scriptStr) {
		t_error.LogErr(errors.New("bad scriptPubKey"))
	}
	matches := re.FindAllStringSubmatch(scriptStr, 1)
	return matches[0][0]


}

func P2PKH_LockScript(addrHex string) [][]byte {
	addrBytes, err := hex.DecodeString(addrHex)
	t_error.LogErr(err)
	return [][]byte{
		{byte(OP_DUP)}, // 1
		{byte(OP_HASH160)}, // 1 
		{byte(OP_PUSHDATA1)}, // 1 
		{byte(0x14)}, // 1
		addrBytes[1:21], // 20
		{byte(OP_EQUALVERIFY)}, // 1
		{byte(OP_CHECKSIG)}, // 1
	}
}