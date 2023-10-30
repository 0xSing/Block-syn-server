package contract

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"strings"
	"walletSynV2/node"
)

const erc20ABI = `[{"inputs":[{"internalType":"string","name":"name_","type":"string"},{"internalType":"string","name":"symbol_","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"}]`

var tokenAbi, _ = abi.JSON(strings.NewReader(erc20ABI))

func Decimals(tokenAddr string) uint8 {
	method := "decimals"
	input, _ := tokenAbi.Pack(method)

	tokenAddress := common.HexToAddress(tokenAddr)
	callMsg := ethereum.CallMsg{
		From: common.HexToAddress("0x16178b55b663Fa065f8054391C15cDa30B700Add"),
		To:   &tokenAddress,
		Data: input,
	}

	output, _ := node.EthClient.GetClient().CallContract(context.Background(), callMsg, nil)
	result, _ := tokenAbi.Unpack(method, output)

	decimals := result[0].(uint8)
	return decimals
}

func Name(tokenAddr string) string {
	method := "name"
	input, _ := tokenAbi.Pack(method)

	tokenAddress := common.HexToAddress(tokenAddr)
	callMsg := ethereum.CallMsg{
		From: common.HexToAddress("0x16178b55b663Fa065f8054391C15cDa30B700Add"),
		To:   &tokenAddress,
		Data: input,
	}

	output, _ := node.EthClient.GetClient().CallContract(context.Background(), callMsg, nil)
	result, _ := tokenAbi.Unpack(method, output)

	name := result[0].(string)
	return name
}

func Symbol(tokenAddr string) string {
	method := "symbol"
	input, _ := tokenAbi.Pack(method)

	tokenAddress := common.HexToAddress(tokenAddr)
	callMsg := ethereum.CallMsg{
		From: common.HexToAddress("0x16178b55b663Fa065f8054391C15cDa30B700Add"),
		To:   &tokenAddress,
		Data: input,
	}

	output, _ := node.EthClient.GetClient().CallContract(context.Background(), callMsg, nil)
	result, _ := tokenAbi.Unpack(method, output)

	symbol := result[0].(string)
	return symbol
}
