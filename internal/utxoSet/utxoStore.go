package utxoSet

import (
	"bytes"
	"container/list"
	"path"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/dgraph-io/badger/v4"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/transaction"
)

type UtxoStore struct {
	ctx *t_config.Context
	db  *badger.DB
}

func NewUtxoStore(ctx *t_config.Context) *UtxoStore {
	store := new(UtxoStore)
	store.ctx = ctx
	var err error
	store.db, err = badger.Open(badger.DefaultOptions(path.Join(ctx.DataDir, "utxoSet.db")))
	t_error.LogErr(err)
	return store
}

func (store *UtxoStore) Close() {
	store.db.Close()
}

func (store *UtxoStore) Read(pt *transaction.OutPoint) (*transaction.Utxo, bool) {
	var utxo transaction.Utxo
	buffer := bytes.Buffer{}
	enc := transaction.NewOutPointEncoder(&buffer)
	enc.Encode(pt)

	err := store.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(buffer.Bytes())

		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			utxobytes := bytes.Buffer{}
			utxobytes.Write(val)
			utxoDec := transaction.NewUtxoDecoder(&utxo)
			utxoDec.Decode(&buffer)
			return nil
		})

		return err
	})
	if err == badger.ErrKeyNotFound {
		return nil, false
	}
	t_error.LogErr(err)
	return &utxo, true
}

func (store *UtxoStore) Write(utxo *transaction.Utxo) {
	buffer := bytes.Buffer{}
	utxoEnc := transaction.NewUtxoEncoder(&buffer)
	utxoEnc.Encode(utxo)

	err := store.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(utxoEnc.Bytes()[:36], utxoEnc.Bytes()[36:])
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

func (store *UtxoStore) FindUTXOsByAddr(addrHex string) *list.List {
	utxos := list.New()
	err := store.db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		utxoDec := transaction.NewUtxoDecoder(nil)
		buffer := bytes.Buffer{}
		for iter.Rewind(); iter.Valid(); iter.Next() {
			item := iter.Item()
			err := item.Value(func(val []byte) error {

				buffer.Write(val)
				utxoDec.Decode(&buffer)
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
	return utxos
}
