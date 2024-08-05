package wallet

import (
	"bytes"
	"encoding/hex"
	"errors"
	"sync"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/client"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/t_util"
	"github.com/terium-project/terium/internal/transaction"
	"github.com/terium-project/terium/internal/utxoSet"
)

type WalletController struct {
	wallet *Wallet
	ctx    *t_config.Context
}

func NewWalletController(wallet *Wallet, ctx *t_config.Context) *WalletController {
	w := new(WalletController)
	w.wallet = wallet
	w.ctx = ctx
	return w
}

func (w *WalletController) GenP2PKH(
	utxoOutPoints []transaction.OutPoint,
	recipientAddrs []string,
	recipientVal []int64,
	sigHashFlags []byte,
	locktime uint32) *transaction.Tx {

	nIn := len(utxoOutPoints)
	nOut := len(recipientAddrs)
	inputs := make([]transaction.TxIn, nIn)
	outputs := make([]transaction.TxOut, nOut)
	utxos := make([]*transaction.Utxo, nIn)
	store := utxoSet.NewUtxoStore(w.ctx)
	defer store.Close()

	for i, outPoint := range utxoOutPoints {
		u, ok := store.Read(&outPoint)
		if !ok {
			panic("Utxo doesnt exist.")
		}
		utxos[i] = u
		inputs[i] = transaction.TxIn{
			PrevOutpt:           outPoint,
			UnlockingScriptSize: transaction.NewCompactSize(0),
			UnlockingScript:     [][]byte{{0x00}},
		}
	}
	wg := sync.WaitGroup{}

	for i := range nOut {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outputs[i] = transaction.TxOut{
				Value:             recipientVal[i],
				LockingScriptSize: transaction.NewCompactSize(transaction.P2PKH_LOCK_SCRIPT_SZ),
				LockingScript: [][]byte{
					{byte(transaction.OP_DUP)},         // 1
					{byte(transaction.OP_HASH160)},     // 1
					{byte(transaction.OP_PUSHDATA1)},   // 1
					{byte(0x14)},                       // 1
					[]byte(recipientAddrs[i])[1:21],    // 20
					{byte(transaction.OP_EQUALVERIFY)}, // 1
					{byte(transaction.OP_CHECKSIG)},    // 1
				},
			}
		}(i)
	}
	wg.Wait()

	tx := &transaction.Tx{
		Version:    t_config.Version,
		NumInputs:  uint8(nIn),
		Inputs:     inputs,
		NumOutputs: uint8(nOut),
		Outputs:    outputs,
		LockTime:   locktime,
	}

	w.SignTx(tx, utxos, sigHashFlags)
	return tx

}

func (w *WalletController) SignTxIn(
	tx *transaction.Tx,
	inIdx uint8,
	inUTXO *transaction.Utxo,
	sigHashFlag byte) {
	sig := w.wallet.ClientId.Sign(tx.Preimage(inIdx, inUTXO, sigHashFlag))
	pk := client.MarshalPubKey(w.wallet.ClientId.PublicKey)
	nPk := len(pk)
	tx.Inputs[inIdx].UnlockingScript = [][]byte{
		{byte(transaction.OP_PUSHDATA1)},
		{byte(0x20)},
		sig,
		{byte(transaction.OP_PUSHDATA1)},
		{byte(nPk)},
		pk,
	}
}

func (w *WalletController) SignTx(
	tx *transaction.Tx,
	inUTXO []*transaction.Utxo,
	sigHashFlags []byte,
) {
	for i := range inUTXO {
		w.SignTxIn(tx, uint8(len(inUTXO)), inUTXO[i], sigHashFlags[i])
	}
}

func (w *WalletController) Balance() int64 {
	var sum int64 = 0
	store := utxoSet.NewUtxoStore(w.ctx)
	defer store.Close()
	utxos := store.FindUTXOsByAddr(w.wallet.ClientId.Address)
	for elem := utxos.Front(); elem != nil; elem.Next() {
		utxo, ok := elem.Value.(transaction.Utxo)
		if !ok {
			t_error.LogErr(errors.New("bad utxo"))
		}
		sum += utxo.Value
	}
	return sum
}

func ValidateAddress(hexxAddr string) bool {
	if len(hexxAddr) != 50 {
		return false
	}
	addrbytes, err := hex.DecodeString(hexxAddr)
	t_error.LogErr(err)
	return bytes.Equal(t_util.Hash256(addrbytes[:len(addrbytes)-4])[:4], addrbytes[len(addrbytes)-4:])
}
