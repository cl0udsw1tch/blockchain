package t_config

import (
	"encoding/json"
	"os"
	"path"

	"github.com/tiereum/trmclient/internal/t_error"
)

const (
	Version           int32 = 0x01
)


type TERIUM_ROOT_ERR struct{}
type DATA_DIR_ERR struct{}
type TMP_DIR_ERR struct{}
type WALLET_DIR_ERR struct{}
type DB_DIR_ERR struct{}

func (e TERIUM_ROOT_ERR) Error() string {

	return "TERIUM_ROOT environment variable invalid."
}
func (e DATA_DIR_ERR) Error() string {
	return ".data directory does not exist."
}
func (e TMP_DIR_ERR) Error() string {
	return ".tmp directory does not exist."
}
func (e WALLET_DIR_ERR) Error() string {
	return "wallets directory does not exist."
}
func (e DB_DIR_ERR) Error() string {
	return ".db directory does not exist."
}


type Context struct {
	ClientRoot 		string
	WalletDir   	string
	ConfigPath  	string
	NetworkConfig  *NetworkConfig
}

type NetworkConfig struct {
	ClientPort *uint16 `json:"clientPort"`
	NodeTable	[]string `json:"nodeTable"`
}

var defaultClientPort uint16 = 7033


func NewContext() *Context {
	ctx := new(Context)
	t_error.LogErr(ctx.GetPaths())
	ctx.GetConfig()
	return ctx
}

func (ctx *Context) GetPaths() error {
	home := os.Getenv("HOME")
	if home == "" {
		os.Exit(1)
	}

	ctx.ClientRoot = path.Join(home, "tiereum/client")
	ctx.ConfigPath = path.Join(ctx.ClientRoot, "config.json")

	
	ctx.WalletDir = path.Join(ctx.ClientRoot, "wallets")
	if _, err := os.Stat(ctx.WalletDir); os.IsNotExist(err) {
		os.Mkdir(ctx.WalletDir, os.FileMode(0777))
	} else if err != nil {
		return err
	}
	ctx.ConfigPath = path.Join(ctx.ClientRoot, "config.json")
	if _, err := os.Stat(ctx.ConfigPath); os.IsNotExist(err) {
		fs, err := os.Create(ctx.ConfigPath)
		t_error.LogErr(err)
		_, err = fs.WriteString("{}")
		t_error.LogErr(err)
		err = fs.Close()
		t_error.LogErr(err)
	} else if err != nil {
		return err
	}
	return nil
}

func (ctx *Context) GetConfig() {

	conf, err := os.ReadFile(ctx.ConfigPath)
	t_error.LogErr(err)

	err = json.Unmarshal(conf, ctx.NetworkConfig)
	t_error.LogErr(err)

	var changed bool = false

	if ctx.NetworkConfig.ClientPort == nil {
		ctx.NetworkConfig.ClientPort = &defaultClientPort
		changed = true
	}

	if changed {
		bytes, err := json.Marshal(ctx.NetworkConfig)
		t_error.LogErr(err)
		os.WriteFile(ctx.ConfigPath, bytes, os.FileMode(0777))
	}
}
