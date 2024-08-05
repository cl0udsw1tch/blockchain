package wallet

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"path"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/client"
	"github.com/terium-project/terium/internal/t_error"
)

type Wallet struct {
	ClientId *client.ClientId
	Name     string
	path     string
	ctx      *t_config.Context
}

func NewWallet(ctx *t_config.Context, name string) *Wallet {
	w := new(Wallet)
	w.path = path.Join(ctx.WalletDir, name, ".wallet")
	w.ctx = ctx
	w.Name = name
	return w
}

func (w *Wallet) Serialize() []byte {
	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(w.ClientId)
	return buffer.Bytes()
}

func (w *Wallet) Deserialize(b []byte) {
	client := client.ClientId{}
	buffer := bytes.Buffer{}
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&client)
	t_error.LogErr(err)
	w.ClientId = &client
}

func (w *Wallet) Exists() bool {
	if _, err := os.Stat(w.path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (w *Wallet) Create() {
	if w.Exists() {
		panic("Wallet with same name exists.")
	}
	fh, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY, 0600)
	t_error.LogErr(err)
	w.ClientId = client.NewClientId()
	os.WriteFile(w.path, w.Serialize(), 0600)
	fh.Close()
}

func (w *Wallet) Read() {
	fileBytes, err := os.ReadFile(w.path)
	t_error.LogErr(err)
	w.Deserialize(fileBytes)
}
