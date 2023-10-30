package contract

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"math/big"
	"strings"
	"walletSynV2/etc"
	"walletSynV2/node"
	"walletSynV2/utils/tron/core"
	zlog "walletSynV2/utils/zlog_sing"
)

const redPacketABI = `[{"inputs":[{"internalType":"address","name":"newPublicKey","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"claimer","type":"address"},{"indexed":false,"internalType":"uint256","name":"claimedValue","type":"uint256"},{"indexed":false,"internalType":"address","name":"tokenAddress","type":"address"}],"name":"ClaimSuccess","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"total","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"string","name":"name","type":"string"},{"indexed":false,"internalType":"string","name":"message","type":"string"},{"indexed":false,"internalType":"address","name":"creator","type":"address"},{"indexed":false,"internalType":"uint256","name":"creationTime","type":"uint256"},{"indexed":false,"internalType":"address","name":"tokenAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"number","type":"uint256"},{"indexed":false,"internalType":"bool","name":"ifrandom","type":"bool"},{"indexed":false,"internalType":"uint256","name":"duration","type":"uint256"}],"name":"CreationSuccess","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"tokenAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"remainingBalance","type":"uint256"}],"name":"RefundSuccess","type":"event"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"name":"checkAvailability","outputs":[{"internalType":"address","name":"tokenAddress","type":"address"},{"internalType":"uint256","name":"remainingTokens","type":"uint256"},{"internalType":"uint256","name":"totalNumber","type":"uint256"},{"internalType":"uint256","name":"claimedNumber","type":"uint256"},{"internalType":"uint256","name":"ifrandom","type":"uint256"},{"internalType":"bool","name":"expired","type":"bool"},{"internalType":"uint256","name":"claimedAmount","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"uint256","name":"receivedAmount","type":"uint256"},{"internalType":"bytes","name":"signedMsg","type":"bytes"}],"name":"claim","outputs":[{"internalType":"uint256","name":"claimed","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"number","type":"uint256"},{"internalType":"bool","name":"ifrandom","type":"bool"},{"internalType":"uint256","name":"duration","type":"uint256"},{"internalType":"string","name":"_message","type":"string"},{"internalType":"string","name":"_name","type":"string"},{"internalType":"uint256","name":"tokenType","type":"uint256"},{"internalType":"address","name":"tokenAddr","type":"address"},{"internalType":"uint256","name":"totalTokens","type":"uint256"}],"name":"createRedPacket","outputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"startIndex","type":"uint256"},{"internalType":"uint256","name":"endIndex","type":"uint256"}],"name":"getExpiredPackets","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"tokenAddr","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"tokenType","type":"uint256"}],"internalType":"struct FinRedPacket.ExpiredPacket[]","name":"","type":"tuple[]"},{"internalType":"uint256","name":"validIndex","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getRedPacketNum","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"name":"ownerRefund","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"startIndex","type":"uint256"},{"internalType":"uint256","name":"endIndex","type":"uint256"}],"name":"ownerRefundRange","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"redpacketById","outputs":[{"components":[{"internalType":"uint256","name":"packed1","type":"uint256"},{"internalType":"uint256","name":"packed2","type":"uint256"}],"internalType":"struct FinRedPacket.Packed","name":"packed","type":"tuple"},{"internalType":"address","name":"publicKey","type":"address"},{"internalType":"address","name":"creator","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newPublicKey","type":"address"}],"name":"setPublicKey","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"}]`

var redPackAbi, _ = abi.JSON(strings.NewReader(redPacketABI))

func RedPacketOwner() string {
	method := "owner"
	input, _ := redPackAbi.Pack(method)

	callMsg := ethereum.CallMsg{
		From: common.HexToAddress("0x16178b55b663Fa065f8054391C15cDa30B700Add"),
		To:   &etc.Conf.EthContract.EthRedPacketAddr,
		Data: input,
	}
	output, _ := node.EthClient.GetClient().CallContract(context.Background(), callMsg, nil)
	result, _ := redPackAbi.Unpack(method, output)

	ownerAddr := result[0].(common.Address)
	return ownerAddr.Hex()
}

func CheckAvailability(idStr string) (string, *big.Int, int64, int64, bool, bool, *big.Int) {
	method := "checkAvailability"
	var id big.Int
	id.SetString(idStr, 10)

	input, _ := redPackAbi.Pack(method, &id)

	callMsg := ethereum.CallMsg{
		From: common.HexToAddress("0x16178b55b663Fa065f8054391C15cDa30B700Add"),
		To:   &etc.Conf.EthContract.EthRedPacketAddr,
		Data: input,
	}

	output, _ := node.EthClient.GetClient().CallContract(context.Background(), callMsg, nil)

	result, _ := redPackAbi.Unpack(method, output)

	tokenAddress := result[0].(common.Address).Hex()
	remainingTokens := result[1].(*big.Int)

	totalNum := result[2].(*big.Int)
	claimedNum := result[3].(*big.Int)

	ifRandom := result[4].(*big.Int)
	var isRandom bool
	if ifRandom.Cmp(big.NewInt(1)) == 0 {
		isRandom = true
	} else {
		isRandom = false
	}

	expired := result[5].(bool)
	claimedAmount := result[6].(*big.Int)
	return tokenAddress, remainingTokens, totalNum.Int64(), claimedNum.Int64(), isRandom, expired, claimedAmount

}

func TronCheckAvailability(idStr string) (string, *big.Int, int64, int64, bool, bool, *big.Int) {
	method := "checkAvailability"

	var id big.Int
	id.SetString(idStr, 10)

	input, err := redPackAbi.Pack(method, &id)
	if err != nil {
		fmt.Println("pack data error", err)
	}

	contract := new(core.TriggerSmartContract)
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		zlog.Zlog.Info("gen private key", zap.Error(err))
	}
	contract.OwnerAddress = hdwallet.PubkeyToTronAddress(privateKey.PublicKey).Bytes()
	contract.ContractAddress, _ = hdwallet.DecodeCheck(etc.Conf.Contract.RedPacketAddr)
	contract.Data = input

	transferTransactionEx, err := node.TronClient.GetClient().TriggerConstantContract(context.Background(), contract)
	if err != nil {
		zlog.Zlog.Info("call contract err", zap.Error(err))
	}

	if transferTransactionEx == nil || len(transferTransactionEx.GetConstantResult()) == 0 {
		zlog.Zlog.Info("empty data")
	}

	output := transferTransactionEx.GetConstantResult()[0]
	result, _ := redPackAbi.Unpack(method, output)

	tokenAddress := result[0].(common.Address).Hex()
	remainingTokens := result[1].(*big.Int)

	totalNum := result[2].(*big.Int)
	claimedNum := result[3].(*big.Int)

	ifRandom := result[4].(*big.Int)
	var isRandom bool
	if ifRandom.Cmp(big.NewInt(1)) == 0 {
		isRandom = true
	} else {
		isRandom = false
	}

	expired := result[5].(bool)
	claimedAmount := result[6].(*big.Int)
	return tokenAddress, remainingTokens, totalNum.Int64(), claimedNum.Int64(), isRandom, expired, claimedAmount
}
