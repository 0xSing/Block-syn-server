package database

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	Leveldb *leveldb.DB
)

func InitLevelDB(WhichNetwork uint64) error {
	var err error
	Leveldb, err = leveldb.OpenFile(fmt.Sprintf("./chain_%d", WhichNetwork), nil)
	return err
}
