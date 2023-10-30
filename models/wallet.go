package models

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"strconv"
	"strings"
	"walletSynV2/database"
	"walletSynV2/utils"
)

var (
	txLock utils.DBKeyLock
)

const TxNft = 99

type Tx struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Contract    string  `json:"contract"`
	Status      uint64  `json:"status"`
	Amount      big.Int `json:"amount"`
	Hash        string  `json:"hash"`
	ChainId     big.Int `json:"chain_id"`
	Gas         big.Int `json:"gas"`
	Time        uint64  `json:"time"`
	TxType      int     `json:"tx_type"`
	BlockNumber big.Int `json:"block_number"`
	GasPrice    big.Int `json:"gas_price"`
}

type NftTx struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	Contract    string  `json:"contract"`
	TokenId     string  `json:"token_id"`
	Status      uint64  `json:"status"`
	Hash        string  `json:"hash"`
	ChainId     big.Int `json:"chain_id"`
	Gas         big.Int `json:"gas"`
	Time        uint64  `json:"time"`
	TxType      int     `json:"tx_type"`
	BlockNumber big.Int `json:"block_number"`
	GasPrice    big.Int `json:"gas_price"`
}

type Ftswap struct {
	From        string  `json:"from"`
	ToAddr      string  `json:"to_addr"`
	FromToken   string  `json:"from_token"`
	ToToken     string  `json:"to_token"`
	AmountIn    big.Int `json:"amount_in"`
	AmountOut   big.Int `json:"amount_out"`
	Contract    string  `json:"contract"`
	Status      uint64  `json:"status"`
	Hash        string  `json:"hash"`
	ChainId     big.Int `json:"chain_id"`
	Gas         big.Int `json:"gas"`
	Time        uint64  `json:"time"`
	TxType      int     `json:"tx_type"`
	BlockNumber big.Int `json:"block_number"`
	GasPrice    big.Int `json:"gas_price"`
}

type OwnNfts struct {
	Contract string `json:"contract"`
	TokenIds string `json:"token_ids"`
}

func HasContract(key string) bool {
	has, _ := database.Leveldb.Has([]byte("contract-"+key), nil)
	return has
}

func HasNft(key string) bool {
	has, _ := database.Leveldb.Has([]byte("nft-"+key), nil)
	return has
}

func HasWallet(key string) bool {
	has, _ := database.Leveldb.Has([]byte("wallet-"+key), nil)
	return has
}

func NewTx(wallet string, tx *Tx) error {
	tkey := wallet + "-tx-"
	txLock.Lock(tkey)
	defer txLock.Unlock(tkey)
	iter := database.Leveldb.NewIterator(util.BytesPrefix([]byte(tkey)), nil)
	num := getNum(&iter)
	key := fmt.Sprintf("%s-tx-%08d", wallet, num)
	js, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func NewNftTx(wallet string, tx *NftTx) error {
	iter := database.Leveldb.NewIterator(util.BytesPrefix([]byte(wallet+"-nft-")), nil)
	num := getNum(&iter)
	key := fmt.Sprintf("%s-nft-%08d", wallet, num)
	js, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func NewFtswap(wallet string, tx *Ftswap) error {
	iter := database.Leveldb.NewIterator(util.BytesPrefix([]byte(wallet+"-ftswap-")), nil)
	num := getNum(&iter)
	key := fmt.Sprintf("%s-ftswap-%08d", wallet, num)
	js, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func getNum(iter *iterator.Iterator) (num int64) {
	if (*iter).Last() {
		vA := strings.Split(string((*iter).Key()), "-")
		num, _ = strconv.ParseInt(vA[len(vA)-1], 10, 64)
		num++
	} else {
		num = 0
	}
	(*iter).Release()
	return num
}

func Put(key string, value string) error {
	return database.Leveldb.Put([]byte(key), []byte(value), nil)
}

func GetTx(key string, hash string) (re string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()

	iter := snapshot.NewIterator(util.BytesPrefix([]byte(key)), nil)
	defer iter.Release()
	re = ""
	for iter.Next() {
		if strings.Contains(string(iter.Value()), hash) {
			re = string(iter.Value())
		}
	}
	return
}

func GetSubset(key string, token string, page int, size int, aType uint) string {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()

	iter := snapshot.NewIterator(util.BytesPrefix([]byte(key)), nil)
	defer iter.Release()

	wallet := strings.Split(key, "-")[0]

	var jsonA []string
	for iter.Next() {
		isSwap := strings.Contains(string(iter.Key()), "-ftswap-")
		if aType == TxNft && isSwap {
			continue
		}
		if string(iter.Value()) == "" || string(iter.Value()) == " " ||
			strings.Contains(string(iter.Key()), "-ownNft-") {
			continue
		}
		//fmt.Println(string(iter.Value()))
		if token == "" {
			jsonA = append(jsonA, string(iter.Value()))
		} else if token == "0x0000000000000000000000000000000000000000" {
			s := "\"contract\":\"\","
			if strings.Contains(string(iter.Value()), s) {
				jsonA = append(jsonA, string(iter.Value()))
			}
			if isSwap {
				f := "\"from_token\":\"\","
				fw := fmt.Sprintf("\"from\":\"%s\",", wallet)
				t := "\"to_token\":\"\","
				tw := fmt.Sprintf("\"to_addr\":\"%s\",", wallet)
				if (strings.Contains(string(iter.Value()), f) && strings.Contains(string(iter.Value()), fw)) ||
					(strings.Contains(string(iter.Value()), t) && strings.Contains(string(iter.Value()), tw)) {
					jsonA = append(jsonA, string(iter.Value()))
				}
			}
		} else {
			s := fmt.Sprintf("\"contract\":\"%s\",", token)
			if strings.Contains(string(iter.Value()), s) {
				jsonA = append(jsonA, string(iter.Value()))
			}
			if isSwap {
				f := fmt.Sprintf("\"from_token\":\"%s\",", token)
				fw := fmt.Sprintf("\"from\":\"%s\",", wallet)
				t := fmt.Sprintf("\"to_token\":\"%s\",", token)
				tw := fmt.Sprintf("\"to_addr\":\"%s\",", wallet)
				if (strings.Contains(string(iter.Value()), f) && strings.Contains(string(iter.Value()), fw)) ||
					(strings.Contains(string(iter.Value()), t) && strings.Contains(string(iter.Value()), tw)) {
					jsonA = append(jsonA, string(iter.Value()))
				}
			}
		}
	}
	if len(jsonA) == 0 {
		return ""
	}
	if len(jsonA) < size*page {
		return ""
	}

	//reverse
	for i, j := 0, len(jsonA)-1; i < j; i, j = i+1, j-1 {
		jsonA[i], jsonA[j] = jsonA[j], jsonA[i]
	}

	if size != 0 {
		if len(jsonA) >= size*(page+1) {
			jsonA = jsonA[size*page : size*(page+1)]
		} else {
			jsonA = jsonA[size*page:]
		}
	}
	//js, _ := json.Marshal(jsonA)
	return "[" + strings.Join(jsonA, ",") + "]"
}

func GetSub(key string) string {
	iter := database.Leveldb.NewIterator(util.BytesPrefix([]byte(key)), nil)
	defer iter.Release()
	var jsonA []string
	for iter.Next() {
		jsonA = append(jsonA, string(iter.Value()))
	}
	if len(jsonA) == 0 {
		return ""
	}
	return "[" + strings.Join(jsonA, ",") + "]"
}

func GetNfts(key string) string {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()

	iter := snapshot.NewIterator(util.BytesPrefix([]byte(key+"-ownNft-")), nil)
	defer iter.Release()
	var nfts []OwnNfts
	for iter.Next() {
		nfts = append(nfts, OwnNfts{
			Contract: strings.Split(string(iter.Key()), "-")[2],
			TokenIds: string(iter.Value()),
		})
	}
	re, _ := json.Marshal(nfts)
	return string(re)
}

func GetToBeNft() (contract string, blockNum int64) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()
	iter := snapshot.NewIterator(util.BytesPrefix([]byte("tobe-nft-")), nil)
	defer iter.Release()
	for iter.Next() {
		contract = strings.Split(string(iter.Key()), "-")[2]
		blockNum, _ = strconv.ParseInt(string(iter.Value()), 10, 64)
	}
	return contract, blockNum
}

func GetAll() {
	iter := database.Leveldb.NewIterator(nil, nil)
	for iter.Next() {
		fmt.Println("key ", string(iter.Key()), "value ", string(iter.Value()))
	}
}

func Get(key string) (v string, err error) {
	if !Has(key) {
		return v, nil
	}
	va, err := database.Leveldb.Get([]byte(key), nil)
	return string(va), err
}

func Del(key string) error {
	return database.Leveldb.Delete([]byte(key), nil)
}

func UpdateNfts(keys []string, values []string, tp int) error {
	batch := new(leveldb.Batch)
	for i, key := range keys {
		if tp == 1 {
			batch.Put([]byte(key), []byte(values[i]))
		} else if tp == 0 {
			batch.Delete([]byte(key))
		}
	}
	return database.Leveldb.Write(batch, nil)
}

func GetSnap() (keys []string, values []string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()
	iter := snapshot.NewIterator(util.BytesPrefix([]byte("snap-")), nil)
	defer iter.Release()
	for iter.Next() {
		keys = append(keys, string(iter.Key())[5:])
		values = append(values, string(iter.Value()))
	}
	return
}
