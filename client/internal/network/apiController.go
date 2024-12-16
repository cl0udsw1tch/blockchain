package network

import (
	"bytes"

	"github.com/tiereum/trmclient/internal/t_config"
	"github.com/tiereum/trmclient/internal/transaction"
)


type ApiController struct {
	Network *Network
	ApiInputChannel chan []byte
	ApiResponseChannel chan []byte
}

func NewApiController(ctx *t_config.Context) *ApiController {
	return &ApiController{
		Network: NewNetwork(ctx),
		ApiInputChannel: make(chan []byte),
		ApiResponseChannel: make(chan []byte),
	}
}

func (a *ApiController) BroadCastTx(tx *transaction.Tx) {
	b := bytes.Buffer{}
	encoder := transaction.NewTxEncoder(&b)
	encoder.Encode(tx)

	a.Network.Broadcast(b.Bytes(), a.ApiInputChannel, a.ApiResponseChannel)
}

func (a *ApiController) RequestBalance() int64 {


	return -1
}

