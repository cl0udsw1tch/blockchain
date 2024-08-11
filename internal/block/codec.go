package block

import (
	"bytes"
	"encoding/binary"
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

func NewHeaderDecoder(header *Header) *HeaderDecoder {
	dec := new(HeaderDecoder)
	if header == nil {	
		dec.header = new(Header)
	} else {
		dec.header = header
	}
	return dec
}

func (d *HeaderDecoder) Clear() {
	d.header = &Header{}
}

func (d *HeaderDecoder) Decode(buffer *bytes.Buffer) error {

	if buffer.Len() < 4 + 32 + 32 + 4 + 1 + 4 {
		return BAD_HEADER_ERR{}
	}

	d.header.Version = int32(binary.BigEndian.Uint32(buffer.Next(4)))
	d.header.PrevHash = buffer.Next(32)
	d.header.MerkleRootHash = buffer.Next(32)
	d.header.TimeStamp = binary.BigEndian.Uint32(buffer.Next(4))
	d.header.Target = buffer.Next(1)[0]
	d.header.Nonce = binary.BigEndian.Uint32(buffer.Next(4))
	
	return nil
}

func (d *HeaderDecoder) Out() *Header {
	return d.header
}

// **** ENCODER **** //

type HeaderEncoder struct {
	buffer *bytes.Buffer
}

func NewHeaderEncoder(buffer *bytes.Buffer) *HeaderEncoder {
	enc := new(HeaderEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *HeaderEncoder) Clear() {
	e.buffer = new(bytes.Buffer)
}

func (e *HeaderEncoder) Encode(header *Header) {
	binary.Write(e.buffer, binary.BigEndian, header.Version)
	binary.Write(e.buffer, binary.BigEndian, header.PrevHash)
	binary.Write(e.buffer, binary.BigEndian, header.MerkleRootHash)
	binary.Write(e.buffer, binary.BigEndian, header.TimeStamp)
	binary.Write(e.buffer, binary.BigEndian, header.Target)
	binary.Write(e.buffer, binary.BigEndian, header.Nonce)
}

func (e *HeaderEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}

// Header codec ================================== //


// Block codec ================================== //

// **** DECODER **** //
type BlockDecoder struct {
	block *Block
}

func NewBlockDecoder(block *Block) *BlockDecoder {
	dec := new(BlockDecoder)
	if block == nil {	
		dec.block = new(Block)
	} else {
		dec.block = block
	}
	return dec
}

func (d *BlockDecoder) Clear() {
	d.block = new(Block)
}

func (d *BlockDecoder) Decode(buffer *bytes.Buffer) error {
{}
	headerDecoder := NewHeaderDecoder(nil)
	if err := headerDecoder.Decode(buffer); err != nil {
		return BAD_BLOCK_ERR{}
	}
	d.block.Header = *headerDecoder.Out()
	if buffer.Len() < 4 {
		return BAD_BLOCK_ERR{}
	}
	d.block.TXCount = binary.BigEndian.Uint32(buffer.Next(4))

	txDec := transaction.NewTxDecoder(nil)
	
	var i uint32 = 0
	for i < d.block.TXCount {
		txDec.Clear()
		if err := txDec.Decode(buffer); err != nil {
			return BAD_BLOCK_ERR{}
		}
		d.block.Transactions = append(d.block.Transactions, *txDec.Out())
		i++
	}

	return nil
}

func (d *BlockDecoder) Out() *Block {
	return d.block
}

// **** ENCODER **** //

type BlockEncoder struct {
	buffer *bytes.Buffer
}

func NewBlockEncoder(buffer *bytes.Buffer) *BlockEncoder {
	enc := new(BlockEncoder)
	if buffer == nil {	
		enc.buffer = new(bytes.Buffer)
	} else {
		enc.buffer = buffer
	}
	return enc
}

func (e *BlockEncoder) Clear(){
	e.buffer = new(bytes.Buffer)
}

func (e *BlockEncoder) Encode(block *Block) {

	headerEnc := NewHeaderEncoder(e.buffer)
	headerEnc.Encode(&block.Header)

	binary.Write(e.buffer, binary.BigEndian, block.TXCount)

	txEnc := transaction.NewTxEncoder(e.buffer)
	for _, tx := range block.Transactions {
		txEnc.Encode(&tx)
	}
}

func (e *BlockEncoder) Bytes() []byte {
	return e.buffer.Bytes()
}


// Block codec ================================== //