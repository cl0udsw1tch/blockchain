package server

import (
	"bytes"

	"github.com/terium-project/terium/internal/block"
)

type _BlockStream struct {
	inStream  chan *block.Block
	outStream chan *block.Block
}

func NewBlockStream() *_BlockStream {
	stream := new(_BlockStream)
	stream.inStream = make(chan *block.Block)
	stream.outStream = make(chan *block.Block)
	return stream
}

func (stream *_BlockStream) InStream() chan *block.Block {
	return stream.inStream
}

func (stream *_BlockStream) OutStream() chan *block.Block {
	return stream.outStream
}

func (stream *_BlockStream) writeToOut(b []byte) {

	buffer := bytes.Buffer{}
	buffer.Write(b)
	blockDec := block.NewBlockDecoder(nil)
	blockDec.Decode(&buffer)
	stream.outStream <- blockDec.Out()
}

func (stream *_BlockStream) readFromIn() []byte {

	buffer := bytes.Buffer{}
	buffer.WriteByte(BLOCK_MESSAGE)
	blockEnc := block.NewBlockEncoder(&buffer)
	block := <-stream.inStream
	blockEnc.Encode(block)
	return blockEnc.Bytes()
}
