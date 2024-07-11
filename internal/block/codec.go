package block

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/terium-project/terium/internal/transaction"
)

type (
	BAD_HEADER_ERR struct {}
	BAD_BLOCK_ERR struct {}
)

func (e BAD_HEADER_ERR) Error() string {
	return "Incorrect block header format."
}

func (e BAD_BLOCK_ERR) Error() string {
	return "Incorrect block header format."
}

// Header codec ================================== //

// **** DECODER **** //
type HeaderDecoder struct {
	header *Header
}

func (d HeaderDecoder) New(header *Header) {
	d.header = header
}

func (d HeaderDecoder) Clear() {
	d.header = nil
}

func (d HeaderDecoder) Decode(buffer bytes.Buffer) error {

	if buffer.Len() < 4 + 32 + 32 + 4 + 256 + 4 {
		return BAD_HEADER_ERR{}
	}

	d.header.Version = int32(binary.BigEndian.Uint32(buffer.Next(4)))
	d.header.PrevHash = buffer.Next(32)
	d.header.MerkleRootHash = buffer.Next(32)
	d.header.TimeStamp = binary.BigEndian.Uint32(buffer.Next(4))
	d.header.Target = big.Int{}
	d.header.Target.SetBytes(buffer.Next(256))
	d.header.Nonce = binary.BigEndian.Uint32(buffer.Next(4))
	
	return nil
}

func (d HeaderDecoder) Out() *Header {
	return d.header
}

// **** ENCODER **** //

type HeaderEncoder struct {
	buffer *bytes.Buffer
}

func (e HeaderEncoder) New(buffer *bytes.Buffer) {
	e.buffer = buffer
}

func (e HeaderEncoder) Clear() {
	e.buffer = nil
}

func (e HeaderEncoder) Encode(header Header) {
	binary.Write(e.buffer, binary.BigEndian, header.Version)
	binary.Write(e.buffer, binary.BigEndian, header.PrevHash)
	binary.Write(e.buffer, binary.BigEndian, header.MerkleRootHash)
	binary.Write(e.buffer, binary.BigEndian, header.TimeStamp)
	binary.Write(e.buffer, binary.BigEndian, header.Target)
	binary.Write(e.buffer, binary.BigEndian, header.Nonce)
}

func (e HeaderEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// Header codec ================================== //


// Block codec ================================== //

// **** DECODER **** //
type BlockDecoder struct {
	block *Block
}

func (d BlockDecoder) New(block *Block) {
	d.block = block
}

func (d BlockDecoder) Clear() {
	d.block = nil
}

func (d BlockDecoder) Decode(buffer bytes.Buffer) error {
	headerDecoder := HeaderDecoder{}
	headerDecoder.New(&d.block.Header)
	if err := headerDecoder.Decode(buffer); err != nil {
		return BAD_BLOCK_ERR{}
	}
	if buffer.Len() < 4 {
		return BAD_BLOCK_ERR{}
	}
	d.block.TXCount = binary.BigEndian.Uint32(buffer.Next(4))

	txDec := transaction.TxDecoder{}
	
	var i uint32 = 0
	for i < d.block.TXCount {
		txDec.Clear()
		tx := transaction.Tx{}
		txDec.New(&tx)
		if err := txDec.Decode(buffer); err != nil {
			return BAD_BLOCK_ERR{}
		}
		d.block.Transactions = append(d.block.Transactions, tx)
		i++
	}

	return nil
}

func (d BlockDecoder) Out() *Block {
	return d.block
}

// **** ENCODER **** //

type BlockEncoder struct {
	buffer *bytes.Buffer
}

func (e BlockEncoder) New(buffer *bytes.Buffer) {
	e.buffer = buffer
}

func (e BlockEncoder) Clear(){
	e.buffer = nil
}

func (e BlockEncoder) Encode(block Block) {
	headerEnc := HeaderEncoder{}
	headerEnc.New(e.buffer)
	headerEnc.Encode(block.Header)

	binary.Write(e.buffer, binary.BigEndian, block.TXCount)

	txEnc := transaction.TxEncoder{}
	for _, tx := range block.Transactions {
		txEnc.Clear()
		txEnc.New(e.buffer)
		txEnc.Encode(tx)
	}
}

func (e BlockEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}


// Block codec ================================== //