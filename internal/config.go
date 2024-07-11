package internal

import (
	"os"
	"path"
)

type TERIUM_ROOT_ERR struct {}
type DATA_DIR_ERR struct {}
type TMP_DIR_ERR struct {}
type WALLET_DIR_ERR struct {}
type DB_DIR_ERR struct {}

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

type DirCtx struct {
	TerieumRoot string
	DataDir 	string
	IndexDir	string
	TmpDir		string
	WalletDir	string
}

var T_DirCtx DirCtx

func (ctx *DirCtx) Config() error {
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
	return nil
}