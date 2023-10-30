package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/smirkcat/hdwallet"
	"math"
	"math/big"
	"net/url"
	"strings"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/service/contract"
	"walletSynV2/utils"
	zlog "walletSynV2/utils/zlog_sing"
)

func SpliceShareLink(hash string) (string, error) {
	cryptoId, err := utils.RedPacketCrypto(hash)
	if err != nil {
		return "", err
	}
	escape := url.QueryEscape(cryptoId)
	uri := fmt.Sprintf("%s%s", etc.Conf.ServerUrl.RedPacketShareUrl, escape)
	return uri, nil
}

func GetRPAmount(id string) (string, error) {
	var packet models.RedPacket
	packetStr := models.GetRedPacket(id)
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		zlog.Zlog.Error("packet explain err")
		return "", err
	}
	var (
		tokenAddr       string
		remainingAmount *big.Int
		totalNumber     int64
		claimedNumber   int64
		ifrandom        bool
		expired         bool
	)

	if etc.Conf.Server.ChainMode == 1 {
		tokenAddr, remainingAmount, totalNumber, claimedNumber, ifrandom, expired, _ = contract.TronCheckAvailability(id)
	} else {
		tokenAddr, remainingAmount, totalNumber, claimedNumber, ifrandom, expired, _ = contract.CheckAvailability(id)
	}
	if expired {
		zlog.Zlog.Error("Red Packet was expired:!")
		return "", errors.New("Red Packet was expired:!")
	}

	envelope := generateLuckyRedEnvelope(remainingAmount, int(totalNumber-claimedNumber), tokenAddr, ifrandom)
	return envelope.String(), nil
}

func GetRPAmounts(id string) ([]string, error) {
	var packet models.RedPacket
	packetStr := models.GetRedPacket(id)
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		zlog.Zlog.Error("packet explain err")
		return nil, err
	}
	var (
		tokenAddr       string
		remainingAmount *big.Int
		totalNumber     int64
		claimedNumber   int64
		ifrandom        bool
		expired         bool
	)

	if etc.Conf.Server.ChainMode == 1 {
		tokenAddr, remainingAmount, totalNumber, claimedNumber, ifrandom, expired, _ = contract.TronCheckAvailability(id)
	} else {
		tokenAddr, remainingAmount, totalNumber, claimedNumber, ifrandom, expired, _ = contract.CheckAvailability(id)
	}
	if expired {
		zlog.Zlog.Error("Red Packet was expired!")
		return nil, errors.New("Red Packet was expired!")
	}

	envelopes := getAllLuckyRedEnvelope(remainingAmount, int(totalNumber-claimedNumber), tokenAddr, ifrandom)
	return envelopes, nil
}

func GenerateClaimSign(idStr string, amountStr string, receiver string) (string, error) {
	var id big.Int
	id.SetString(idStr, 10)

	var amount big.Int
	amount.SetString(amountStr, 10)

	if etc.Conf.Server.ChainMode == 1 {
		tronVar, _ := hdwallet.DecodeCheck(receiver)
		receiver = common.BytesToAddress(tronVar[1:]).Hex()
	}

	prefix := "\x19FinToken RedPacket Signed Message:\n32"

	hash := crypto.Keccak256Hash(
		[]byte(prefix),
		common.LeftPadBytes(id.Bytes(), 32),
		common.LeftPadBytes(amount.Bytes(), 32),
		common.HexToAddress(receiver).Bytes(),
	)
	sign, err := getSignature("54f77eee4bc271bceb83726fc33f1d43d35fa6afa2934358d027d81e983d5a8d", hash)
	if err != nil {
		return "", err
	}

	return sign, nil
}

func getSignature(pKey string, hash common.Hash) (string, error) {
	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		fmt.Println("get private key error", err)
		return "", err
	}
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		fmt.Println("sign error", err)
		return "", err
	}
	//for solidity verify
	if signature[64] == 0 || signature[64] == 1 {
		signature[64] += 27
	}
	return hexutil.Encode(signature), nil
}

func generateLuckyRedEnvelope(remainAmount *big.Int, remainNum int, tokenAddr string, ifRandom bool) *big.Int {
	if remainNum == 1 {
		return remainAmount
	}

	//合约查询精度决定红包最小值
	baseDecimals := 6
	var tokenDecimals int
	if strings.EqualFold(tokenAddr, "0x0000000000000000000000000000000000000000") {
		if etc.Conf.Server.ChainMode == 1 {
			tokenDecimals = 6
		} else {
			tokenDecimals = 18
		}
	} else if etc.Conf.Server.ChainMode == 1 {
		address := common.HexToAddress(tokenAddr)
		bytes := address.Hash().Bytes()
		bytes[11] = 0x41
		tronAddr := hdwallet.EncodeCheck(bytes[11:])
		tokenDecimals = int(contract.TronDecimals(tronAddr))
	} else {
		tokenDecimals = int(contract.Decimals(tokenAddr))
	}

	min := new(big.Int).Mul(big.NewInt(1), big.NewInt(int64(math.Pow(10, float64(tokenDecimals-baseDecimals)))))
	if remainNum == 0 {
		return big.NewInt(0)
	}
	averageAmount := new(big.Int).Div(remainAmount, big.NewInt(int64(remainNum)))
	max := new(big.Int).Mul(averageAmount, big.NewInt(2))

	// 如果为平均红包
	if !ifRandom {
		return averageAmount
	}

	// 计算最大可用金额，确保不会出现空包红包
	randAmount, _ := rand.Int(rand.Reader, max)
	if randAmount.Cmp(min) < 0 {
		remainAmount = min
	}

	return randAmount
}

func getAllLuckyRedEnvelope(totalAmount *big.Int, remainNum int, tokenAddr string, ifRandom bool) []string {
	var luckyNum []string

	varAmount := big.NewInt(0)
	num := remainNum
	for i := 0; i < num; i++ {
		redEnvelopeAmount := generateLuckyRedEnvelope(totalAmount, remainNum, tokenAddr, ifRandom)
		fmt.Println("本次开包金额：", redEnvelopeAmount)
		totalAmount = new(big.Int).Sub(totalAmount, redEnvelopeAmount)
		remainNum--
		varAmount = varAmount.Add(varAmount, redEnvelopeAmount)
		luckyNum = append(luckyNum, redEnvelopeAmount.String())
	}
	fmt.Println("红包总金额：", varAmount)

	return luckyNum
}
