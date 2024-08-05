package server

import (
	"errors"

	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/network"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/transaction"
)

const (
	TX_MESSAGE    byte = 0x00
	BLOCK_MESSAGE byte = 0x01
)

type Server struct {
	ctx         *t_config.Context
	network     *network.Network
	txStream    *_TxStream
	blockStream *_BlockStream
}

func NewServer(ctx *t_config.Context) *Server {
	server := new(Server)
	server.ctx = ctx
	server.network = network.NewNetwork(ctx)
	server.txStream = NewTxStream()
	server.blockStream = NewBlockStream()
	return server
}

func (server *Server) Run() {
	go server.Broadcast()
	go server.Listen()
}

func (server *Server) Listen() {
	go server.network.Listen()
	for data := range server.network.IStream() {
		switch data[0] {
		case TX_MESSAGE:
			server.txStream.writeToOut(data[1:])
		case BLOCK_MESSAGE:
			server.blockStream.writeToOut(data[1:])
		default:
			t_error.LogWarn(errors.New("undefined message code"))
			continue
		}

	}
}

func (server *Server) Broadcast() {

	go server.network.Broadcast()

	go func() {
		server.network.OStream() <- server.blockStream.readFromIn()
	}()

	go func() {
		server.network.OStream() <- server.txStream.readFromIn()
	}()

}

type TxStream struct {
	InStream  chan<- *transaction.Tx
	OutStream <-chan *transaction.Tx
}

type BlockStream struct {
	InStream  chan<- *block.Block
	OutStream <-chan *block.Block
}

func (server *Server) Tx() *TxStream {
	return &TxStream{InStream: server.txStream.inStream, OutStream: server.txStream.outStream}
}
func (server *Server) Block() *BlockStream {
	return &BlockStream{InStream: server.blockStream.inStream, OutStream: server.blockStream.outStream}
}
