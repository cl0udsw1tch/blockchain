package utxoSet

import (
	"bytes"
	"container/list"
	"errors"
	"path"

	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"
	"github.com/tiereum/trmnode/internal/transaction"

	"github.com/dgraph-io/badger/v4"
)

type UtxoStore struct {
	ctx *t_config.Context
	db  *badger.DB
}

func NewUtxoStore(ctx *t_config.Context) *UtxoStore {
	store := new(UtxoStore)
	store.ctx = ctx
	var err error
	opts := badger.DefaultOptions(path.Join(ctx.DataDir, "utxoSet"))
	opts.Logger = nil
	store.db, err = badger.Open(opts)
	t_error.LogErr(err)
	return store
}

func (store *UtxoStore) Close() {
	store.db.Close()
}

func (store *UtxoStore) Read(pt *transaction.OutPoint) (*transaction.Utxo, bool) {
	var utxo *transaction.Utxo
	enc := transaction.NewOutPointEncoder(nil)
	enc.Encode(pt)

	err := store.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(enc.Bytes())

		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			utxobytes := new(bytes.Buffer)
			utxobytes.Write(val)
			utxoDec := transaction.NewUtxoDecoder(nil)
			utxoDec.Decode(utxobytes)
			utxo = utxoDec.Out()
			return nil
		})

		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, false
	}
	t_error.LogErr(err)
	return utxo, true
}

func (store *UtxoStore) Write(utxo *transaction.Utxo) {

	utxoEnc := transaction.NewUtxoEncoder(nil)
	utxoEnc.Encode(utxo)

	outptEnc := transaction.NewOutPointEncoder(nil)
	outptEnc.Encode(&utxo.OutPoint)

	err := store.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(outptEnc.Bytes(), utxoEnc.Bytes())
		return err
	})

	t_error.LogErr(err)
}

func (store *UtxoStore) Delete(pt *transaction.OutPoint) {
	buffer := bytes.Buffer{}
	ptEnc := transaction.NewOutPointEncoder(&buffer)
	ptEnc.Encode(pt)

	err := store.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(ptEnc.Bytes())
		return err
	})

	t_error.LogErr(err)
}

func (store *UtxoStore) FindUTXOsByAddr(addrHex string) []transaction.Utxo {
	utxos := list.New()
	err := store.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		utxoDec := transaction.NewUtxoDecoder(nil)
		buffer := new(bytes.Buffer)
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			err := item.Value(func(val []byte) error {

				buffer.Write(val)
				err := utxoDec.Decode(buffer)
				t_error.LogErr(err)
				lockaddr := transaction.GetAddrFromP2PKHLockScript(utxoDec.Out().LockingScript)
				if lockaddr == addrHex {
					utxos.PushBack(utxoDec.Out())
				}
				buffer.Reset()
				utxoDec.Clear()
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	t_error.LogErr(err)
	utxoSlice := make([]transaction.Utxo, utxos.Len())
	curr := utxos.Front()
	i := 0
	for curr != nil {

		currVal, ok := curr.Value.(*transaction.Utxo)
		if !ok {
			t_error.LogErr(errors.New("bad utxo format in mempool"))
		}
		utxoSlice[i] = *currVal
		curr = curr.Next()
		i++
	}
	return utxoSlice
}
