package server

import (
	"bytes"

	"github.com/terium-project/terium/internal/transaction"
)

type _TxStream struct {
	inStream  chan *transaction.Tx
	outStream chan *transaction.Tx
}

func NewTxStream() *_TxStream {
	stream := new(_TxStream)
	stream.inStream = make(chan *transaction.Tx)
	stream.outStream = make(chan *transaction.Tx)
	return stream
}

func (stream *_TxStream) InStream() chan *transaction.Tx {
	return stream.inStream
}

func (stream *_TxStream) OutStream() chan *transaction.Tx {
	return stream.outStream
}

func (stream *_TxStream) writeToOut(b []byte) {

	buffer := bytes.Buffer{}
	buffer.Write(b)
	txDec := transaction.NewTxDecoder(nil)
	txDec.Decode(&buffer)
	stream.outStream <- txDec.Out()
}

func (stream *_TxStream) readFromIn() []byte {
	buffer := bytes.Buffer{}
	buffer.WriteByte(TX_MESSAGE)
	txEnc := transaction.NewTxEncoder(&buffer)
	tx := <-stream.inStream
	txEnc.Encode(tx)
	return txEnc.Bytes()
}
