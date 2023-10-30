package node

import (
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"walletSynV2/etc"
	zlog "walletSynV2/utils/zlog_sing"
)

var (
	EthClient  *EthNodeClient
	TronClient *TronNodeClient
)

func InitEthClient() (err error) {
	isDev := strings.EqualFold(etc.Conf.Server.NodeMode, "dev")

	EthClient, err = NewEthClient(etc.Conf.Server.NodeUrl, isDev)
	if err != nil {
		zlog.Zlog.Error("create ether client fail", zap.Error(err))
		return
	}
	InitEthContract()
	return
}

func InitTronClient() (err error) {
	isDev := strings.EqualFold(etc.Conf.Server.NodeMode, "dev")

	TronClient, err = NewTronClient(etc.Conf.Server.NodeUrl, isDev)
	if err != nil {
		zlog.Zlog.Error("create tron client fail", zap.Error(err))
		return
	}
	return
}

func InitEthContract() {
	contract := etc.Conf.Contract
	sourceValue := reflect.ValueOf(contract).Elem()
	targetValue := reflect.New(reflect.TypeOf(etc.EthContract{})).Elem()

	// 遍历源结构体的字段
	for i := 0; i < sourceValue.NumField(); i++ {
		// 获取源结构体字段的值
		paramValue := sourceValue.Field(i).Interface().(string)
		// 将字符串参数转换为 common.Address
		addressParam := common.HexToAddress(paramValue)

		// 将 common.Address 参数设置到目标结构体中的对应字段
		targetField := targetValue.Field(i)
		if targetField.CanSet() {
			targetField.Set(reflect.ValueOf(addressParam))
		}
	}
	ethContract := targetValue.Interface().(etc.EthContract)
	etc.Conf.EthContract = &ethContract
}
