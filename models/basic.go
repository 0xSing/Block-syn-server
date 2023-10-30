package models

import (
	"time"
)

type (
	BasicModel struct {
		Status  int32     `json:"status"`
		Deleted bool      `json:"-"`
		Created time.Time `orm:"auto_now_add; type(datetime)" json:"-"`
		Updated time.Time `orm:"auto_now; type(datetime)" json:"-"`
	}
)

// 错误定义
// example:
// var BalanceErr = errors.New("余额不足")
var (
	NftStop  chan struct{}
	NftStart chan struct{}
)
