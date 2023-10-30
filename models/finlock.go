package models

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"walletSynV2/database"
)

type Lock struct {
	Id         string `json:"id"`
	Token      string `json:"token"`
	Owner      string `json:"owner"`
	Name       string `json:"name"'`
	Symbol     string `json:"symbol"`
	Decimals   uint8  `json:"decimals"`
	Amount     string `json:"amount"`
	LockDate   uint64 `json:"lockDate"`
	UnlockDate uint64 `json:"unlockDate"`
	LockStates uint64 `json:"lockStates"`
}

type AddLocked struct {
	Id         *big.Int
	Token      common.Address
	Owner      common.Address
	Amount     *big.Int
	UnlockDate *big.Int
}

type LockUpdated struct {
	Id            *big.Int
	Token         common.Address
	Owner         common.Address
	NewAmount     *big.Int
	NewUnlockDate *big.Int
}

type LockRemoved struct {
	Id         *big.Int
	Token      common.Address
	Owner      common.Address
	Amount     *big.Int
	UnlockedAt *big.Int
}

type LockOwnerChanged struct {
	LockId   *big.Int
	Owner    common.Address
	NewOwner common.Address
}

func NewLockData(lock *Lock) error {
	key := fmt.Sprintf("FinLock-%s-%s", lock.Owner, lock.Id)
	indexKey := fmt.Sprintf("FinLockIndex-%s", lock.Id)

	batch := new(leveldb.Batch)

	js, err := json.Marshal(lock)
	if err != nil {
		return err
	}

	batch.Put([]byte(key), js)
	batch.Put([]byte(indexKey), js)
	return database.Leveldb.Write(batch, nil)
}

func UpdateLockData(newLock *Lock) error {
	key := fmt.Sprintf("FinLock-%s-%s", newLock.Owner, newLock.Id)
	value, err := Get(key)

	var lock Lock
	json.Unmarshal([]byte(value), &lock)

	lock.Amount = newLock.Amount
	lock.UnlockDate = newLock.UnlockDate

	js, err := json.Marshal(lock)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func UnlockData(unlocked *Lock) error {
	key := fmt.Sprintf("FinLock-%s-%s", unlocked.Owner, unlocked.Id)
	value, err := Get(key)

	var lock Lock
	json.Unmarshal([]byte(value), &lock)

	lock.LockStates = 3
	lock.UnlockDate = unlocked.UnlockDate

	js, err := json.Marshal(lock)
	if err != nil {
		return err
	}

	return database.Leveldb.Put([]byte(key), js, nil)
}

func AbandonLock(lockArgs *Lock, newOwner string) error {
	key := fmt.Sprintf("FinLock-%s-%s", lockArgs.Owner, lockArgs.Id)
	value, err := Get(key)

	newKey := fmt.Sprintf("FinLock-%s-%s", newOwner, lockArgs.Id)

	var lock Lock
	json.Unmarshal([]byte(value), &lock)

	lock.Owner = newOwner

	js, err := json.Marshal(lock)
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)
	batch.Delete([]byte(key))
	batch.Put([]byte(newKey), js)

	return database.Leveldb.Write(batch, nil)
}

func ChangeLockStates(owner string, id string, isPending bool) error {
	key := fmt.Sprintf("FinLock-%s-%s", owner, id)
	value, err := Get(key)

	var lock Lock
	json.Unmarshal([]byte(value), &lock)

	if isPending {
		lock.LockStates = 2
	} else {
		lock.LockStates = 1
	}

	js, err := json.Marshal(lock)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func GetLockById(id string) (re string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()

	strKey := fmt.Sprintf("FinLockIndex-%s", id)
	json, err := snapshot.Get([]byte(strKey), nil)
	if err != nil {
		// 处理错误
	}
	re = string(json)
	return
}

func GetLocks(key string) (re string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		// 处理错误
	}
	defer snapshot.Release()

	strKey := fmt.Sprintf("FinLock-%s", key)

	iter := snapshot.NewIterator(util.BytesPrefix([]byte(strKey)), nil)
	defer iter.Release()

	var locks []Lock
	var lock Lock
	for iter.Next() {
		err := json.Unmarshal(iter.Value(), &lock)
		if err != nil {
			return err.Error()
		}
		locks = append(locks, lock)
	}
	js, _ := json.Marshal(locks)
	return string(js)
}
