package models

import (
	"strconv"
	"walletSynV2/database"
)

func RegisterMySQLModel() {
	//RegisterTestModels()
}

func ReadChainHeight(key string) (int64, error) {
	if !Has(key) {
		return 0, nil
	}
	height, err := database.Leveldb.Get([]byte(key), nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(height), 10, 64)
}

func UpdateChainHeight(key string, height int64) error {
	h := strconv.FormatInt(height, 10)
	return database.Leveldb.Put([]byte(key), []byte(h), nil)
}

func Has(key string) bool {
	has, _ := database.Leveldb.Has([]byte(key), nil)
	return has
}
