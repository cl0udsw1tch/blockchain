package mempool

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/tiereum/trmnode/internal/t_config"
	"github.com/tiereum/trmnode/internal/t_error"
	"github.com/tiereum/trmnode/internal/transaction"

	_ "github.com/mattn/go-sqlite3"
)

type MempoolMetadata struct {
	Tx  *transaction.Tx
	Fee int64
}

type MempoolIO struct {
	ctx *t_config.Context
	db  *sql.DB
}

func NewMempoolIO(ctx *t_config.Context) *MempoolIO {
	store := new(MempoolIO)
	store.ctx = ctx
	var err error

	store.db, err = sql.Open("sqlite3", ":memory:")
	t_error.LogErr(err)
	bytes, err := os.ReadFile("./internal/mempool/mempool.sql")
	t_error.LogErr(err)
	sqlCommand := string(bytes)
	_, err = store.db.Exec(sqlCommand)
	t_error.LogErr(err)
	return store
}

func (store *MempoolIO) Exists(hash []byte) bool {
	query := fmt.Sprintf("SELECT EXISTS(SELECT txid FROM tierium.mempool WHERE txid='%s') AS row_exists;", hex.EncodeToString(hash))
	row := store.db.QueryRow(query)
	var v int
	row.Scan(&v)
	return v == 1
}

func (store *MempoolIO) Read(hash []byte) (*transaction.Tx, int64, bool) {

	query := fmt.Sprintf("SELECT tx, fee FROM tierium.mempool WHERE txid='%s';", hex.EncodeToString(hash))
	row := store.db.QueryRow(query)

	if row.Err() != nil {
		return nil, 0, false
	}

	var tx string
	var fee int64
	err := row.Scan(&tx, &fee)
	t_error.LogErr(err)
	txBytes, err := hex.DecodeString(tx)
	t_error.LogErr(err)

	buffer := bytes.Buffer{}
	buffer.Write(txBytes)
	dec := transaction.NewTxDecoder(nil)
	err = dec.Decode(&buffer)
	t_error.LogErr(err)
	return dec.Out(), fee, true

}

func (store *MempoolIO) Write(hash []byte, tx *transaction.Tx, fee int64) {
	cmd := "INSERT INTO tierium.mempool (txid, tx, fee) VALUES (?, ?, ?);"
	buffer := new(bytes.Buffer)
	enc := transaction.NewTxEncoder(buffer)
	enc.Encode(tx)
	_, err := store.db.Exec(cmd, hex.EncodeToString(hash), hex.EncodeToString(enc.Bytes()), fee)
	t_error.LogErr(err)
}

func (store *MempoolIO) Delete(hash []byte) {
	cmd := "DELETE FROM tierium.mempool WHERE txid = ?;"
	_, err := store.db.Exec(cmd, hex.EncodeToString(hash))
	t_error.LogErr(err)
}

func (store *MempoolIO) Close() {
	store.db.Close()
}

func (store *MempoolIO) GetTxByPriority(required int64) []transaction.Tx {
	var count int64
	err := store.db.QueryRow("SELECT COUNT(*) FROM tierium.mempool;").Scan(&count)
	t_error.LogErr(err)
	txs := make([]transaction.Tx, count)
	if count < required {
		return []transaction.Tx{}
	}

	cmd := "SELECT txid,tx FROM tierium.mempool ORDER BY fee DESC LIMIT ?;"
	rows, err := store.db.Query(cmd, required)
	t_error.LogErr(err)
	txids := make([]string, count)
	buffer := new(bytes.Buffer)
	dec := transaction.NewTxDecoder(nil)

	i := 0
	for rows.Next() {
		var currTxHex string
		rows.Scan(&txids[i], &currTxHex)
		b, err := hex.DecodeString(currTxHex)
		t_error.LogErr(err)
		buffer.Write(b)
		err = dec.Decode(buffer)
		t_error.LogErr(err)
		txs[i] = *dec.Out()
		i++
	}

	return txs

}

func (store *MempoolIO) GetTxWithLargestFee() string {
	var txid string
	err := store.db.QueryRow("SELECT txid FROM tierium.mempool ORDER BY fee DESC LIMIT 1;").Scan(&txid)
	t_error.LogErr(err)
	return txid
}
