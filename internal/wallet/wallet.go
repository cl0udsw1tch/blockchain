package wallet

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io/fs"
	"log"
	"os"
	"path"

	"github.com/terium-project/terium/internal/client"
	"github.com/terium-project/terium/internal/t_error"
)

type Wallet struct {
	ClientId client.ClientId
	Name     string
	Path	 string
}

func GetWalletDir() string {
	var cwd string
	var err error
	var walletDir string

	cwd, err = os.Getwd()
	t_error.LogErr(&err)

	walletDir = path.Join(cwd, "wallets")
	return walletDir
}

func WalletExists(name string) bool {
	walletDir := GetWalletDir()
	if walletDir == "" {
		return false
	}
	walletPath := path.Join(walletDir, name + ".terium.wallet")
	if _, err := os.Stat(walletPath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func CreateWalletFile(name string) (Wallet, *os.File) {
	walletFile := path.Join(GetWalletDir(), name + ".terium.wallet")
	if WalletExists(name) {
		log.Panic("Wallet already exists.")
	}
	
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	
	
	fh, err := os.OpenFile(walletFile, os.O_CREATE | os.O_RDWR, 0600)

	t_error.LogErr(&err)
	
	wallet := Wallet{
		ClientId: 	client.MakeClient(),
		Name:     	name,
		Path: 		walletFile,
	}

	encoder.Encode(wallet)

	os.WriteFile(name + ".terium.wallet", buffer.Bytes(), 0600)

	return wallet, fh
}

func ReadWalletFile(name string) Wallet {
	fileName := path.Join(GetWalletDir(), name)
	fileBytes, err := os.ReadFile(fileName)
	t_error.LogErr(&err)
	
	var wallet Wallet
	var buffer bytes.Buffer

	_, err = buffer.Write(fileBytes)
	t_error.LogErr(&err)

	decoder := gob.NewDecoder(&buffer)
	err = decoder.Decode(wallet)
	t_error.LogErr(&err)

	return wallet

}

func ReadWalletList() []Wallet{
	wallets := []Wallet{}
	return wallets
}
