package wallet

import (
	"errors"
	"os"
	"path"

	"github.com/terium-project/terium/internal/client"
	"github.com/terium-project/terium/internal/t_config"
	"github.com/terium-project/terium/internal/t_error"
)

type Wallet struct {
	ClientId *client.ClientId
	Name     string
	dir      string
	ctx      *t_config.Context
}

func NewWallet(ctx *t_config.Context, name string) *Wallet {
	w := new(Wallet)
	w.dir = path.Join(ctx.WalletDir, name)
	w.ctx = ctx
	w.Name = name
	return w
}

func (w *Wallet) Serialize() ([]byte, []byte) {
	return client.MarshalPrivKey(w.ClientId.PrivateKey), client.MarshalPubKey(w.ClientId.PublicKey)
}

func (w *Wallet) Deserialize(priv []byte, pub []byte) {
	w.ClientId = client.GetClientId(priv, pub)
}

func (w *Wallet) Exists() bool {
	if _, err := os.Stat(w.dir); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (w *Wallet) Create() {
	if w.Exists() {
		panic("Wallet with same name exists.")
	}
	err := os.Mkdir(w.dir, os.FileMode(0777))
	t_error.LogErr(err)
	fh, err := os.OpenFile(path.Join(w.dir, ".wallet"), os.O_CREATE|os.O_WRONLY, os.FileMode(0666))
	t_error.LogErr(err)
	w.ClientId = client.NewClientId()
	priv, pub := w.Serialize()
	os.WriteFile(path.Join(w.dir, "priv.der"), priv, os.FileMode(0666))
	os.WriteFile(path.Join(w.dir, "pub.der"), pub, os.FileMode(0666))
	fh.Close()
}

func (w *Wallet) Read() {
	priv, err := os.ReadFile(path.Join(w.dir, "priv.der"))
	t_error.LogErr(err)
	pub, err := os.ReadFile(path.Join(w.dir, "pub.der"))
	t_error.LogErr(err)
	w.Deserialize(priv, pub)
}
