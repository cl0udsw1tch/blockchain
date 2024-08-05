package client

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/terium-project/terium/internal/t_error"
)

type ClientId struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey 
	PubKeyHash []byte // 20 bytes
	Address    string // 25 bytes
}

func NewClientId() *ClientId {
	a := ClientId{}
	a.PrivateKey = MakePrivateKey()
	a.PublicKey = GetPublicKey(a.PrivateKey)
	a.PubKeyHash = HashPublicKey(a.PublicKey)
	a.Address = MakeAddress(a.PubKeyHash)
	return &a
}


func (a *ClientId) Init(priv []byte) {
	a.PrivateKey = UnMarshalPrivKey(priv)
	a.PublicKey = GetPublicKey(a.PrivateKey)
	a.PubKeyHash = HashPublicKey(a.PublicKey)
	a.Address = MakeAddress(a.PubKeyHash)
}
func (a *ClientId) Sign(msgHash []byte) []byte {
	r, err := ecdsa.SignASN1(rand.Reader, a.PrivateKey, msgHash)
	t_error.LogErr(err)
	return r
	
}

