package contract

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"strings"
	"walletSynV2/node"
	"walletSynV2/utils/tron/core"
	zlog "walletSynV2/utils/zlog_sing"
)

const tronErc20ABI = `[{"inputs":[{"internalType":"string","name":"name_","type":"string"},{"internalType":"string","name":"symbol_","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`

var (
	tronTokenAbi, _ = abi.JSON(strings.NewReader(tronErc20ABI))
)

func TronName(tokenAddr string) string {
	method := "name"

	input, err := tronTokenAbi.Pack(method)
	if err != nil {
		fmt.Println("pack data error", err)
	}

	contract := new(core.TriggerSmartContract)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		zlog.Zlog.Info("gen private key", zap.Error(err))
	}
	contract.OwnerAddress = hdwallet.PubkeyToTronAddress(privateKey.PublicKey).Bytes()
	contract.ContractAddress, _ = hdwallet.DecodeCheck(tokenAddr)
	contract.Data = input

	transferTransactionEx, err := node.TronClient.GetClient().TriggerConstantContract(context.Background(), contract)
	if err != nil {
		zlog.Zlog.Info("call contract err", zap.Error(err))
	}

	if transferTransactionEx == nil || len(transferTransactionEx.GetConstantResult()) == 0 {
		zlog.Zlog.Info("empty data")
	}

	output := transferTransactionEx.GetConstantResult()[0]
	result, _ := tronTokenAbi.Unpack(method, output)

	s := result[0].(string)
	return s
}

func TronSymbol(tokenAddr string) string {
	method := "symbol"

	input, err := tronTokenAbi.Pack(method)
	if err != nil {
		fmt.Println("pack data error", err)
	}

	contract := new(core.TriggerSmartContract)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		zlog.Zlog.Info("gen private key", zap.Error(err))
	}
	contract.OwnerAddress = hdwallet.PubkeyToTronAddress(privateKey.PublicKey).Bytes()
	contract.ContractAddress, _ = hdwallet.DecodeCheck(tokenAddr)
	contract.Data = input

	transferTransactionEx, err := node.TronClient.GetClient().TriggerConstantContract(context.Background(), contract)
	if err != nil {
		zlog.Zlog.Info("call contract err", zap.Error(err))
	}

	if transferTransactionEx == nil || len(transferTransactionEx.GetConstantResult()) == 0 {
		zlog.Zlog.Info("empty data")
	}

	output := transferTransactionEx.GetConstantResult()[0]
	result, _ := tronTokenAbi.Unpack(method, output)

	s := result[0].(string)
	return s
}

func TronDecimals(tokenAddr string) uint8 {
	method := "decimals"

	input, err := tronTokenAbi.Pack(method)
	if err != nil {
		fmt.Println("pack data error", err)
	}

	contract := new(core.TriggerSmartContract)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		zlog.Zlog.Info("gen private key", zap.Error(err))
	}
	contract.OwnerAddress = hdwallet.PubkeyToTronAddress(privateKey.PublicKey).Bytes()
	contract.ContractAddress, _ = hdwallet.DecodeCheck(tokenAddr)
	contract.Data = input

	transferTransactionEx, err := node.TronClient.GetClient().TriggerConstantContract(context.Background(), contract)
	if err != nil {
		zlog.Zlog.Info("call contract err", zap.Error(err))
	}

	if transferTransactionEx == nil || len(transferTransactionEx.GetConstantResult()) == 0 {
		zlog.Zlog.Info("empty data")
	}

	output := transferTransactionEx.GetConstantResult()[0]
	result, _ := tronTokenAbi.Unpack(method, output)

	s := result[0].(uint8)
	return s
}
