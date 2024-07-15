package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"

	"github.com/mr-tron/base58"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/t_util"
	"golang.org/x/crypto/ripemd160"
)

type ClientId struct {
	PrivateKey []byte
	PublicKey  []byte
	PubKeyHash []byte
	Address    string
}

func (a *ClientId) MakeClientId() {
	privKey := a.MakePrivateKey()
	a.MakePublicKey(&privKey)
	a.HashPublicKey()
	a.MakeAddress()

}

func (a *ClientId) MakePrivateKey() ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

	t_error.LogErr(err)

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(*privKey)

	t_error.LogErr(err)

	privKeyBytes := buffer.Bytes()
	a.PrivateKey = privKeyBytes

	return *privKey

}

func (a *ClientId) MakePublicKey(pk *ecdsa.PrivateKey) {

	pubKey := pk.PublicKey
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(pubKey)

	t_error.LogErr(err)

	pubKeyBytes := buffer.Bytes()
	a.PublicKey = pubKeyBytes
}

func (a *ClientId) HashPublicKey() {
	sha256Hasher := sha256.New()
	sha256Hasher.Write(a.PublicKey)
	pubKeyHash := sha256Hasher.Sum(nil) // sha256
	ripemdHasher := ripemd160.New()
	ripemdHash := ripemdHasher.Sum(pubKeyHash)
	a.PubKeyHash = ripemdHash
}

func (a *ClientId) MakeAddress() {

	address := append([]byte{0x00, 0x00}, a.PubKeyHash...) // 2 byte version


	address = t_util.Hash256(address) // double sha256 hash

	checksum := address[:4]

	address = append(checksum, address...)
	addressString := base58.Encode(address)
	a.Address = addressString
}
