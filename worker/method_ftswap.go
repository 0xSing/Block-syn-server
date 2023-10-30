package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	"walletSynV2/utils/tron/hexutil"
	zlog "walletSynV2/utils/zlog_sing"
)

const (
	ethSwapMethod    = "a22f16b4"
	ethSwapInMethod  = "f060cd5f"
	swapMethod       = "17725244"
	swapInMethod     = "102792f3"
	ftswapTopic      = "0xa9e4f578695aa65d64b1bf1cae8d6e014fa176457bda6a288250840f400dbbde"
	transferidHex    = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	ftswapid         = "a9e4f578695aa65d64b1bf1cae8d6e014fa176457bda6a288250840f400dbbde"
	trxPurchaseTopic = "dad9ec5c9b9c82bf6927bf0b64293dcdd1f82c92793aef3c5f26d7b93a4a5306"
	Ftswap           = 3
)

func EthFlSwapScan(w *Worker, from *common.Address, time uint64, blockNumber *big.Int, tx *types.Transaction) {
	defer w.Wg.Done()
	methodID := common.Bytes2Hex(tx.Data()[0:4])
	if methodID != ethSwapMethod && methodID != ethSwapInMethod {
		return
	}

	receipts, err := node.EthClient.GetClient().TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		zlog.Zlog.Info("Transactionreceipts request fail: "+tx.Hash().String(), zap.Error(err))
		return
	}

	method := ""
	if methodID == ethSwapMethod {
		method = "swap"
	} else if methodID == ethSwapInMethod {
		method = "swapIn"
	}

	abiS := `[{"inputs":[{"internalType":"address","name":"owner","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"target","type":"address"},{"indexed":false,"internalType":"address","name":"fromToken","type":"address"},{"indexed":false,"internalType":"address","name":"toToken","type":"address"},{"indexed":false,"internalType":"uint256","name":"amountIn","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amountOut","type":"uint256"}],"name":"FTSwap_V2","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address payable","name":"_feeTo","type":"address"},{"internalType":"address payable","name":"_altcoinsFeeTo","type":"address"},{"internalType":"uint256","name":"_feeRate","type":"uint256"}],"name":"setFee","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"},{"internalType":"bool","name":"isValid","type":"bool"}],"name":"setTarget","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"},{"internalType":"uint256","name":"_amountIn","type":"uint256"},{"internalType":"uint256","name":"_amountOutMin","type":"uint256"},{"internalType":"address","name":"_fromToken","type":"address"},{"internalType":"address","name":"_toToken","type":"address"},{"internalType":"address[]","name":"_path","type":"address[]"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"enum FTSwap.SwapType","name":"_swapType","type":"uint8"}],"name":"swap","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"},{"internalType":"uint256","name":"_amountInMax","type":"uint256"},{"internalType":"uint256","name":"_amountOut","type":"uint256"},{"internalType":"address","name":"_fromToken","type":"address"},{"internalType":"address","name":"_toToken","type":"address"},{"internalType":"address[]","name":"_path","type":"address[]"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"enum FTSwap.SwapType","name":"_swapType","type":"uint8"}],"name":"swapIn","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"stateMutability":"payable","type":"receive"},{"inputs":[],"name":"getFee","outputs":[{"internalType":"address payable","name":"_feeTo","type":"address"},{"internalType":"address payable","name":"_altcoinsFeeTo","type":"address"},{"internalType":"uint256","name":"_feeRate","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"}],"name":"getTarget","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`
	cabi, err := abi.JSON(strings.NewReader(abiS))
	input, err := cabi.Methods[method].Inputs.Unpack(tx.Data()[4:])
	if err != nil {
		zlog.Zlog.Info("unpack data err--------", zap.Error(err))
	}
	toAddr := fmt.Sprintf("%v", input[6])

	if receipts.Status == 0 {
		if models.HasWallet(from.Hex()) {
			ft := &models.Ftswap{
				From:        from.Hex(),
				ToAddr:      toAddr,
				FromToken:   fmt.Sprintf("%v", input[3]),
				ToToken:     fmt.Sprintf("%v", input[4]),
				Status:      receipts.Status,
				AmountIn:    *big.NewInt(0),
				AmountOut:   *big.NewInt(0),
				Contract:    tx.To().Hex(),
				Hash:        tx.Hash().Hex(),
				ChainId:     *big.NewInt(int64(node.EthClient.GetChainId())),
				Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
				Time:        time,
				TxType:      Ftswap,
				BlockNumber: *blockNumber,
				GasPrice:    *tx.GasPrice(),
			}

			if err := models.NewFtswap(from.Hex(), ft); err != nil {
				zlog.Zlog.Info("new ftswap error "+tx.Hash().String(), zap.Error(err))
				return
			}
			jsB, _ := json.Marshal(ft)
			w.Wg.Add(1)
			go PushTo(w, jsB)
		}
		return
	}

	for i := 0; i < len(receipts.Logs); i++ {
		if !strings.EqualFold(receipts.Logs[i].Topics[0].String(), ftswapTopic) {
			continue
		}

		tokenData := receipts.Logs[i].Data //[0:32] [32:64] [64:96] [96:128] [128:160]
		//fmt.Println("token data:", tokenData)

		fromT := common.BytesToAddress(tokenData[32:64])
		fromTs := fromT.Hex()
		if len(input) == 8 && input[7] == uint8(0) {
			fromTs = ""
		}

		toT := common.BytesToAddress(tokenData[64:96])
		toTs := toT.Hex()
		if len(input) == 8 && input[7] == uint8(1) {
			toTs = ""
		}
		//fmt.Println("toContractAddress :", toT.Hex())

		amountIn := new(big.Int)
		amountIn.SetBytes(tokenData[96:128])
		//fmt.Println("chainId :", amountIn.Int64())

		amountOut := new(big.Int)
		amountOut.SetBytes(tokenData[128:160])
		if i > 0 && strings.EqualFold(receipts.Logs[i-1].Topics[0].String(), transferidHex) &&
			strings.EqualFold(common.HexToAddress(receipts.Logs[i-1].Topics[2].Hex()).Hex(), fmt.Sprintf("%v", input[6])) {
			amountOut.SetBytes(receipts.Logs[i-1].Data)
		}

		ft := &models.Ftswap{
			From:        from.Hex(),
			ToAddr:      toAddr,
			FromToken:   fromTs,
			ToToken:     toTs,
			Status:      receipts.Status,
			AmountIn:    *amountIn,
			AmountOut:   *amountOut,
			Contract:    tx.To().Hex(),
			Hash:        tx.Hash().Hex(),
			ChainId:     *big.NewInt(int64(node.EthClient.GetChainId())),
			Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
			Time:        time,
			TxType:      Ftswap,
			BlockNumber: *blockNumber,
			GasPrice:    *tx.GasPrice(),
		}

		if err := models.NewFtswap(from.Hex(), ft); err != nil {
			zlog.Zlog.Info("new ftswap error "+tx.Hash().String(), zap.Error(err))
			return
		}
		if !strings.EqualFold(from.Hex(), toAddr) {
			if err := models.NewFtswap(toAddr, ft); err != nil {
				zlog.Zlog.Info("new ftswap toAddr error "+tx.Hash().String(), zap.Error(err))
				return
			}
		}
		//push
		jsB, _ := json.Marshal(ft)
		w.Wg.Add(1)
		go PushTo(w, jsB)
	}
}

func TronFlSwapScan(w *Worker, unObj *core.TriggerSmartContract, txid *api.BytesMessage, txStatus uint64, contract string, transactionId string, currentHeight int64) {
	defer w.Wg.Done()
	client := node.TronClient
	//  ftswap 合约
	data := unObj.GetData()
	if len(data) < 4 {
		return
	}
	if hexutil.Encode(data[:4]) != swapMethod && hexutil.Encode(data[:4]) != swapInMethod {
		return
	}
	from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
	transinfo, _ := client.GetClient().GetTransactionInfoById(context.Background(), txid)

	abiS := `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"target","type":"address"},{"indexed":false,"internalType":"address","name":"fromToken","type":"address"},{"indexed":false,"internalType":"address","name":"toToken","type":"address"},{"indexed":false,"internalType":"uint256","name":"amountIn","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"amountOut","type":"uint256"}],"name":"FTSwap_V2","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Paused","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"account","type":"address"}],"name":"Unpaused","type":"event"},{"inputs":[],"name":"getFee","outputs":[{"internalType":"address payable","name":"_feeTo","type":"address"},{"internalType":"address payable","name":"_altcoinsFeeTo","type":"address"},{"internalType":"uint256","name":"_feeRate","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"pause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"paused","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address payable","name":"_feeTo","type":"address"},{"internalType":"address payable","name":"_altcoinsFeeTo","type":"address"},{"internalType":"uint256","name":"_feeRate","type":"uint256"}],"name":"setFee","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"},{"internalType":"uint256","name":"_amountIn","type":"uint256"},{"internalType":"uint256","name":"_amountOutMin","type":"uint256"},{"internalType":"address","name":"_fromToken","type":"address"},{"internalType":"address","name":"_toToken","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"enum FTSwap.SwapType","name":"_swapType","type":"uint8"}],"name":"swap","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"_target","type":"address"},{"internalType":"uint256","name":"_amountInMax","type":"uint256"},{"internalType":"uint256","name":"_amountOut","type":"uint256"},{"internalType":"address","name":"_fromToken","type":"address"},{"internalType":"address","name":"_toToken","type":"address"},{"internalType":"address","name":"_to","type":"address"},{"internalType":"enum FTSwap.SwapType","name":"_swapType","type":"uint8"}],"name":"swapIn","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"unpause","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
	cabi, err := abi.JSON(strings.NewReader(abiS))
	input, err := cabi.Methods["swap"].Inputs.Unpack(data[4:])
	if err != nil {
		zlog.Zlog.Info("unpack data err--------", zap.Error(err))
		return
	}
	to := common.HexToHash(fmt.Sprintf("%v", input[5]))
	to[11] = 0x41
	toAddr := hdwallet.EncodeCheck(to[11:32])

	if txStatus == 0 {
		f := common.HexToHash(fmt.Sprintf("%v", input[3]))
		t := common.HexToHash(fmt.Sprintf("%v", input[4]))
		f[11] = 0x41
		t[11] = 0x41
		fromT := hdwallet.EncodeCheck(f[11:32])
		toT := hdwallet.EncodeCheck(t[11:32])
		//失败交易
		if models.HasWallet(from) {
			ft := &models.Ftswap{
				From:        from,
				ToAddr:      toAddr,
				FromToken:   fromT,
				ToToken:     toT,
				Status:      txStatus,
				AmountIn:    *big.NewInt(0),
				AmountOut:   *big.NewInt(0),
				Contract:    contract,
				Hash:        transactionId,
				ChainId:     *big.NewInt(int64(client.GetChainId())),
				Gas:         *big.NewInt(transinfo.GetFee()),
				Time:        uint64(transinfo.BlockTimeStamp) / 1000,
				TxType:      Ftswap,
				BlockNumber: *big.NewInt(currentHeight),
				GasPrice:    *big.NewInt(0),
			}
			if err := models.NewFtswap(from, ft); err != nil {
				zlog.Zlog.Info("new ftswap error: "+transactionId, zap.Error(err))
				return
			}
			jsB, _ := json.Marshal(ft)
			w.Wg.Add(1)
			go PushTo(w, jsB)
		}
		return
	}

	var amountOut *big.Int
	for _, evenlog := range transinfo.Log {
		//special handler of swap amountOut
		if hexutil.Encode(evenlog.Topics[0]) == trxPurchaseTopic {
			amountOut = new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Topics[3]))
			amountOut = amountOut.Mul(amountOut, big.NewInt(995))
			amountOut = amountOut.Div(amountOut, big.NewInt(1000))
		}
		if hexutil.Encode(evenlog.Topics[0]) != ftswapid {

			continue
		}
		evenlog.Data[43] = 0x41
		fromT := hdwallet.EncodeCheck(evenlog.Data[43:64])
		if len(input) == 8 && input[7] == uint8(0) {
			fromT = ""
		}
		evenlog.Data[75] = 0x41
		toT := hdwallet.EncodeCheck(evenlog.Data[75:96])
		if len(input) == 8 && input[7] == uint8(1) {
			toT = ""
		}

		amountIn := new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Data[100:128]))
		//amountOut := new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Data[128:160]))

		ft := &models.Ftswap{
			From:        from,
			ToAddr:      toAddr,
			FromToken:   fromT,
			ToToken:     toT,
			Status:      txStatus,
			AmountIn:    *amountIn,
			AmountOut:   *amountOut,
			Contract:    contract,
			Hash:        transactionId,
			ChainId:     *big.NewInt(int64(client.GetChainId())),
			Gas:         *big.NewInt(transinfo.GetFee()),
			Time:        uint64(transinfo.BlockTimeStamp) / 1000,
			TxType:      Ftswap,
			BlockNumber: *big.NewInt(currentHeight),
			GasPrice:    *big.NewInt(0),
		}
		if err := models.NewFtswap(from, ft); err != nil {
			zlog.Zlog.Info("new ftswap error: "+transactionId, zap.Error(err))
			return
		}
		if !strings.EqualFold(from, toAddr) {
			if err := models.NewFtswap(toAddr, ft); err != nil {
				zlog.Zlog.Info("new ftswap toAddr error: "+transactionId, zap.Error(err))
				return
			}
		}
		jsB, _ := json.Marshal(ft)
		w.Wg.Add(1)
		go PushTo(w, jsB)
	}
}

type PushResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Failed  bool   `json:"failed"`
}

func PushTo(w *Worker, js []byte) {
	defer w.Wg.Done()
	for i := 0; i < 15; i++ {
		//fmt.Println("json", json)
		url := etc.Conf.ServerUrl.JavaHttpUrl + "/api/v1/tokenmarket/tradeRecord/public/syncToApp"
		client := &http.Client{}
		req, _ := http.NewRequest("POST", url, bytes.NewReader(js))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			zlog.Zlog.Error("response error in pushTo: " + err.Error())
			return
		}
		body, _ := io.ReadAll(resp.Body)
		r := PushResp{}
		if err := json.Unmarshal(body, &r); err != nil {
			zlog.Zlog.Error("umarshal json error in pushTo: " + err.Error())
			return
		}
		if r.Success {
			zlog.Zlog.Info("push res" + string(js))
			return
		}
		zlog.Zlog.Info("push msg error" + string(body))
		time.Sleep(time.Second * 3)
	}
}
