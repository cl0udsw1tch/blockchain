package blockchain

import (
	"math/big"
	"time"

	"github.com/terium-project/terium/internal/block"
	"github.com/terium-project/terium/internal/transaction"
	. "github.com/terium-project/terium/internal/transaction"
)

const (
	Version int32 = 0x01
	NBits 	uint8 = 0x40
)

var (
	Target big.Int = *(&big.Int{}).Lsh(big.NewInt(1), uint(NBits))
)

type Blockchain struct {
	LastHash	[]byte
	NewBlock *block.Block
	BlockHeight	big.Int
}

func (b Blockchain) Genesis() {
	genesis := block.Block{
		Header: block.Header{
			Version: Version,
			PrevHash: []byte{},
			TimeStamp: uint32(time.Now().Unix()),
			Target: Target,
		},
		TXCount: 0,
		Transactions: nil,
	}
	genesis.Save()
	b.LastHash = genesis.Hash()
	b.BlockHeight = *big.NewInt(0)
}

func (b Blockchain) CreateBlock( 
	CoinBaseScriptSz CompactSize, 
	CoinBaseSript [][]byte,
	nOut uint8, 
	outputs TxOuts, 
) {
	coinbaseTx := CoinbaseTx(uint32(Version), CoinBaseScriptSz, CoinBaseSript, nOut, outputs)
	header := block.Header{
		Version: Version,
		PrevHash: b.LastHash,
		Target: Target,
	}
	b.NewBlock = &block.Block{
		Header: header,
		TXCount: 1,
		Transactions: []transaction.Tx{coinbaseTx},
	}
}

func (b Blockchain) AddBlock(block block.Block) {
	b.NewBlock = &block
	b.LastHash = block.Hash()
	sum := big.Int{}
	sum.Add(&b.BlockHeight, big.NewInt(1))
	b.BlockHeight = sum
	block.Save()
	b.UpdateDB()
}



func (b Blockchain) FindTxOutByOutPt(tx OutPoint) (TxOut, error) {
	var txout TxOut

	return txout, nil

}
