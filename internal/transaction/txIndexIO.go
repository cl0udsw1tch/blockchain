package transaction

import (
	"bytes"
	"encoding/gob"
	"math/big"
	"path"

	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"

	"github.com/dgraph-io/badger/v4"
)

type TxMetadata struct {
	BlockHash   []byte
	BlockHeight big.Int
	Index       uint8
}

type TxIndexIO struct {
	db  *badger.DB
	ctx *t_config.Context
}

func NewTxIndexIO(ctx *t_config.Context) *TxIndexIO {
	io := TxIndexIO{}
	io.ctx = ctx
	var err error
	opts := badger.DefaultOptions(path.Join(ctx.IndexDir, "txIndex"))
	opts.Logger = nil
	io.db, err = badger.Open(opts)
	t_error.LogErr(err)
	return &io
}

func (io *TxIndexIO) Close() {
	io.db.Close()
}

func (io *TxIndexIO) Create(txHash []byte, meta *TxMetadata) {
	err := io.db.Update(func(txn *badger.Txn) error {
		buffer := bytes.Buffer{}
		enc := gob.NewEncoder(&buffer)
		enc.Encode(meta)
		err := txn.Set(txHash, buffer.Bytes())
		return err
	})
	t_error.LogErr(err)
}

func (io *TxIndexIO) Read(txHash []byte) *TxMetadata {
	meta := TxMetadata{}
	err := io.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(txHash)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			buffer := bytes.Buffer{}
			buffer.Write(val)
			dec := gob.NewDecoder(&buffer)
			err := dec.Decode(&meta)
			return err
		})
		return err
	})
	t_error.LogErr(err)
	return &meta
}

func (io *TxIndexIO) Delete(txHash []byte) {
	err := io.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(txHash)
		return err
	})
	t_error.LogErr(err)
}
