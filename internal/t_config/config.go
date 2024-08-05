package t_config

import (
	"encoding/json"
	"math/big"
	"os"
	"path"

	"github.com/terium-project/terium/internal/t_error"
)

const (
	Version           int32 = 0x01
	NBits             uint8 = 0x40
	COINBASE_MATURITY uint8 = 100
	BLOCK_REWARD      int64 = 100_000
)

var (
	Target big.Int = *(&big.Int{}).Lsh(big.NewInt(1), uint(NBits))
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
	return ".wallets directory does not exist."
}
func (e DB_DIR_ERR) Error() string {
	return ".db directory does not exist."
}

type Context struct {
	TerieumRoot string
	DataDir     string
	IndexDir    string
	TmpDir      string
	WalletDir   string
	ConfigPath  string
	NodeConfig  *Config
}

type Config struct {
	NumTxInBlock    *uint8  `json:"numTxInBlock"`
	RpcEndpointPort *uint16 `json:"RpcEndpointPort"`
	ClientAddress   *string `json:"clientAddress"`
}

var NumTxInBlock uint8 = 10
var RpcEndpointPort uint16 = 8033

func NewContext() *Context {
	ctx := new(Context)
	t_error.LogErr(ctx.GetPaths())
	ctx.GetConfig()
	return ctx 
}

func (ctx *Context) GetPaths() error {
	root := os.Getenv("TERIUM_ROOT")
	if root == "" {
		return TERIUM_ROOT_ERR{}
	}
	if _, err := os.Stat(root); err == os.ErrNotExist {
		return TERIUM_ROOT_ERR{}
	} else if err != nil {
		return err
	}
	ctx.TerieumRoot = root
	ctx.ConfigPath = path.Join(root, "config.json")

	// these dont have to exist and can be created by the app

	ctx.DataDir = path.Join(root, ".data")
	if _, err := os.Stat(ctx.DataDir); err == os.ErrNotExist {
		os.Mkdir(ctx.DataDir, 0600)
	} else if err != nil {
		return err
	}
	ctx.IndexDir = path.Join(root, ".data", "index")
	if _, err := os.Stat(ctx.IndexDir); err == os.ErrNotExist {
		os.Mkdir(ctx.IndexDir, 0600)
	} else if err != nil {
		return err
	}
	ctx.TmpDir = path.Join(root, ".tmp")
	if _, err := os.Stat(ctx.TmpDir); err == os.ErrNotExist {
		os.Mkdir(ctx.TmpDir, 0600)
	} else if err != nil {
		return err
	}
	ctx.WalletDir = path.Join(root, ".wallets")
	if _, err := os.Stat(ctx.WalletDir); err == os.ErrNotExist {
		os.Mkdir(ctx.WalletDir, 0600)
	} else if err != nil {
		return err
	}
	ctx.ConfigPath = path.Join(root, "config.json")
	return nil
}

func (ctx *Context) GetConfig() {

	conf, err := os.ReadFile(ctx.ConfigPath)
	t_error.LogErr(err)

	err = json.Unmarshal(conf, ctx.NodeConfig)
	t_error.LogErr(err)

	changed := false
	if ctx.NodeConfig.NumTxInBlock == nil {
		ctx.NodeConfig.NumTxInBlock = &NumTxInBlock
		changed = true
	}

	if ctx.NodeConfig.RpcEndpointPort == nil {
		ctx.NodeConfig.RpcEndpointPort = &RpcEndpointPort
		changed = true
	} 

	if changed {
		bytes, err := json.Marshal(ctx.NodeConfig)
		t_error.LogErr(err)
		os.WriteFile(ctx.ConfigPath, bytes, 0600)
	}
}
