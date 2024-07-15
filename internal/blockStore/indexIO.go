package blockStore

import (
	"bytes"
	"encoding/gob"
	"math/big"
	"path"
	"github.com/dgraph-io/badger/v4"
	"github.com/terium-project/terium/internal"
	"github.com/terium-project/terium/internal/t_error"
)

type __metadata__ struct {
	Nonce uint32
	Height big.Int
}
type IndexIO struct {
	db *badger.DB
	hash []byte
	metadata *__metadata__
	ctx *internal.DirCtx
}


func (store *IndexIO) New(ctx *internal.DirCtx, hash []byte) {
	
	
	store.hash = hash
	store.ctx = ctx
	db, err := badger.Open(badger.DefaultOptions(path.Join(ctx.IndexDir, "index.db")))
	if err != nil {
		t_error.LogErr(err)
	}
	
	store.db = db
	
}

func (store *IndexIO) ReadLastHash() {
	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lastHash"))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			buffer := bytes.Buffer{}
			buffer.Write(val)
			store.hash = buffer.Bytes()
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t_error.LogErr(err)
	}
	store.Read()
}

func (store *IndexIO) WriteLastHash() {
	err := store.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("lastHash"), store.hash)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t_error.LogErr(err)
	}
}


func (store *IndexIO) Create(metadata *BlockMetaData) {
	store.writeMeta()
}

func (store *IndexIO) Read() {
	err := store.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(store.hash)

		if err != nil {
			return err
		}

		item.Value(func(val []byte) error {
			buffer := bytes.Buffer{}
			buffer.Write(val)
			dec := gob.NewDecoder(&buffer)
			dec.Decode(store.metadata)
			return nil
		})

		return nil
	})

	if err != nil {
		t_error.LogErr(err)
	}
}

func (store *IndexIO) Update(metadata *BlockMetaData) {
	store.setMeta(metadata)
	store.writeMeta()
}

func (store *IndexIO) Delete() {
	err := store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(store.hash)
	})

	if err != nil {
		t_error.LogErr(err)
	}
}

func (store *IndexIO) Close(){
	if err := store.db.Close(); err != nil {
		t_error.LogErr(err)
	}
}

func (store *IndexIO) MetaData() *BlockMetaData {
	metadata := BlockMetaData{
		Hash: store.hash,
		Nonce: store.metadata.Nonce,
		Height: store.metadata.Height,
	}
	return &metadata
}


// private

func (store *IndexIO) setMeta(metadata *BlockMetaData) {

	store.metadata = &__metadata__{
		Nonce: metadata.Nonce,
		Height: metadata.Height,
	}
}
func (store *IndexIO) writeMeta() {
	err := store.db.Update(func(txn *badger.Txn) error {
		
		buffer := bytes.Buffer{}
		enc := gob.NewEncoder(&buffer)
		enc.Encode(store.metadata)
		err := txn.Set(store.hash, buffer.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t_error.LogErr(err)
	}
}