package blockStore

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math/big"
	"path"
	"github.com/dgraph-io/badger/v4"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
)

type __metadata__ struct {
	Nonce  uint32
	Height big.Int
}
type IndexIO struct {
	db       *badger.DB
	ctx      *t_config.Context
}

func NewIndexIO(ctx *t_config.Context) *IndexIO {

	i := new(IndexIO)
	opts := badger.DefaultOptions(path.Join(ctx.IndexDir, "blockIndex"))
	opts.Logger = nil
	db, err := badger.Open(opts)
	t_error.LogErr(err)
	i.ctx = ctx
	i.db = db
	return i
}

func (store *IndexIO) ReadLast() (*BlockMetaData, error) {
	var hash []byte
	err := store.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lastHash"))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			buffer := bytes.Buffer{}
			buffer.Write(val)
			hash = buffer.Bytes()
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err == badger.ErrKeyNotFound {
		return nil, err
	}
	meta := store.Read(hash)
	return &BlockMetaData{Hash: hash, Nonce: meta.Nonce, Height: meta.Height}, nil
}

func (store *IndexIO) Write(meta *BlockMetaData) {
	hash := meta.Hash
	__meta := __metadata__{
		Nonce: meta.Nonce,
		Height: meta.Height,
	}
	err := store.db.Update(func(txn *badger.Txn) error {

		buffer := bytes.Buffer{}
		enc := gob.NewEncoder(&buffer)
		enc.Encode(__meta)
		err := txn.Set(hash, buffer.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	t_error.LogErr(err)
	store.WriteLastHash(hash)
}

func (store *IndexIO) WriteLastHash(hash []byte) {
	err := store.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("lastHash"), hash)
		if err != nil {
			return err
		}
		return nil
	})
	t_error.LogErr(err)
}

func (store *IndexIO) Read(hash []byte) *__metadata__ {
	metadata := new(__metadata__)
	err := store.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(hash)

		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			buffer := bytes.Buffer{}
			buffer.Write(val)
			dec := gob.NewDecoder(&buffer)
			dec.Decode(metadata)
			return nil
		})

		return nil
	})

	t_error.LogErr(err)
	return metadata
}


func (store *IndexIO) Delete(hash []byte) {
	err := store.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(hash)
	})
	if err != badger.ErrKeyNotFound {
		t_error.LogErr(err)
	} else {
		t_error.LogWarn(errors.New("key not found, nothing to do"))
	}
}

func (store *IndexIO) Close() {
	if err := store.db.Close(); err != nil {
		t_error.LogErr(err)
	}
}

func (store *IndexIO) MetaData(hash []byte, metadata *__metadata__) *BlockMetaData {
	_metadata := BlockMetaData{
		Hash:   hash,
		Nonce:  metadata.Nonce,
		Height: metadata.Height,
	}
	return &_metadata
}
