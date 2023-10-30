package models

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"walletSynV2/database"
)

type BTransfer struct {
	From        string    `json:"from"`
	To          []string  `json:"to"`
	Contract    string    `json:"contract"`
	Status      uint64    `json:"status"`
	Amount      []big.Int `json:"amount"`
	Total       big.Int   `json:"total"`
	Hash        string    `json:"hash"`
	ChainId     big.Int   `json:"chain_id"`
	Gas         big.Int   `json:"gas"`
	Time        uint64    `json:"time"`
	TxType      int       `json:"tx_type"`
	BlockNumber big.Int   `json:"block_number"`
	GasPrice    big.Int   `json:"gas_price"`
}

func NewBatchTx(wallet string, tx *BTransfer) error {
	iter := database.Leveldb.NewIterator(util.BytesPrefix([]byte(wallet+"-batchTransfer-")), nil)
	num := getNum(&iter)
	key := fmt.Sprintf("%s-batchTransfer-%08d", wallet, num)
	js, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}
