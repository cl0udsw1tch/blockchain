package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"github.com/terium-project/terium/internal/t_error"
	"github.com/terium-project/terium/internal/t_util"
	"golang.org/x/crypto/ripemd160"
)

func MakePrivateKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	t_error.LogErr(err)
	return privKey
}

func GetPublicKey(pk *ecdsa.PrivateKey) *ecdsa.PublicKey {
	return &pk.PublicKey
}

func HashPublicKey(pk *ecdsa.PublicKey) []byte {
	sha256Hasher := sha256.New()
	sha256Hasher.Write(MarshalPubKey(pk))
	pubKeyHash := sha256Hasher.Sum(nil) // sha256
	ripemdHasher := ripemd160.New()
	ripemdHash := ripemdHasher.Sum(pubKeyHash)
	return ripemdHash
}

func MakeAddress(pkhash []byte) string {
	address := append([]byte{0x00, 0x00}, pkhash...) // 2 byte version
	hash := t_util.Hash256(address) // double sha256 hash
	checksum := hash[:4]
	address = append(address, checksum...)
	addressString := hex.EncodeToString(address)
	return addressString
}



func Verify(msghash, sig []byte, pk *ecdsa.PublicKey) bool {
	return ecdsa.VerifyASN1(pk, msghash, sig)
}

func MarshalPubKey(pk *ecdsa.PublicKey) []byte {
	r, err := x509.MarshalPKIXPublicKey(pk)
	t_error.LogErr(err)
	return r
}

func UnMarshalPubKey(b []byte) *ecdsa.PublicKey {
	r, err := x509.ParsePKIXPublicKey(b)
	t_error.LogErr(err)
	key, ok := r.(ecdsa.PublicKey)
	if !ok {
		panic("Bad public key.")
	}
	return &key
}

func MarshalPrivKey(pk *ecdsa.PrivateKey) []byte {
	r, err := x509.MarshalECPrivateKey(pk)
	t_error.LogErr(err)
	return r
}

func UnMarshalPrivKey(b []byte) *ecdsa.PrivateKey {
	r, err := x509.ParseECPrivateKey(b)
	t_error.LogErr(err)
	return r
}