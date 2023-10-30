package models

import (
	"encoding/json"
	"fmt"
	"walletSynV2/database"
)

type MultiSignWallet struct {
	Addr        string   `json:"addr"`
	Wallets     []string `json:"wallets"`
	Times       []string `json:"times"`
	Creator     string   `json:"creator"`
	OwnersCount int64    `json:"owners_count"`
	Threshold   int64    `json:"threshold"`
	Name        string   `json:"name"`
}

func CreateMultiSig(w *MultiSignWallet) (err error) {
	key := fmt.Sprintf("ToBeMultiSig-%s", w.Creator)
	txLock.Lock(key)
	defer txLock.Unlock(key)

	js, err := json.Marshal(w)
	if err != nil {
		return err
	}
	err = database.Leveldb.Put([]byte(key), js, nil)
	return
}

func GetMultiOwner(creator string) (w *MultiSignWallet) {
	key := fmt.Sprintf("ToBeMultiSig-%s", creator)
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
		return
	}
	defer snapshot.Release()

	mw, _ := snapshot.Get([]byte(key), nil)
	if len(mw) == 0 {
		return
	}

	err = json.Unmarshal(mw, &w)
	if err != nil {
		return
	}
	return
}
