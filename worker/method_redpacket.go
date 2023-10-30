package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/utils"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	"walletSynV2/utils/tron/hexutil"
	zlog "walletSynV2/utils/zlog_sing"
)

func RedPacketCommunication(redPacketComm *models.RedPacketCommunication) {
	redPacketCommStr, err := json.Marshal(redPacketComm)
	if err != nil {
		zlog.Zlog.Error("comm explain fail", zap.Error(err))
	}
	zlog.Zlog.Info("send red packet communication :" + string(redPacketCommStr))

	cryptoComm, err := utils.RedPacketCrypto(string(redPacketCommStr))
	if err != nil {
		zlog.Zlog.Error("comm fail", zap.Error(err))
	}
	zlog.Zlog.Info("cryptoComm message:" + cryptoComm)

	reqStr := fmt.Sprintf(`{"encryptData": "%s"}`, cryptoComm)
	body := []byte(reqStr)

	url := etc.Conf.ServerUrl.JavaHttpUrl + "/api/v1/trade/redPacket/public/addRecord"
	c := &http.Client{
		Timeout: 15 * time.Second,
	}

	var respBody redPacketCommResp
	resp, err := c.Post(url, "application/json", bytes.NewBuffer(body))
	defer resp.Body.Close()
	for i := 0; i < 15; i++ {
		respBodyStr, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			zlog.Zlog.Error("respBodyStr explain err", zap.Error(err))
		}
		err = json.Unmarshal(respBodyStr, &respBody)
		if err != nil {
			zlog.Zlog.Error("respBody explain err", zap.Error(err))
		}
		zlog.Zlog.Info("red packet push to Java Server，respon：" + string(respBodyStr))
		if err == nil && resp.StatusCode == 200 && (respBody.Code == 200 || respBody.Code == 10058) {
			return
		}
		time.Sleep(time.Second * 5)
		resp, err = c.Post(url, "application/json", bytes.NewBuffer(body))
	}
}

type redPacketCommResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func RecordEthRedPacketTransfer(w *Worker, tx *types.Transaction, block *types.Block, receipts *types.Receipt, amount *big.Int, isEth bool, isCreate bool) error {
	zlog.Zlog.Info("start eth red packet record....")
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		zlog.Zlog.Info("get from address fail"+tx.Hash().String(), zap.Error(err))
	}
	// from to 判断
	var (
		from_ string
		to_   string
	)

	if isCreate {
		from_ = from.Hex()
		to_ = tx.To().Hex()
	} else {
		from_ = tx.To().Hex()
		to_ = from.Hex()
	}

	if isEth {
		zlog.Zlog.Info("Get eth red packet record....")
		var amount_ big.Int
		if isCreate {
			amount_ = *tx.Value()
		} else {
			amount_ = *amount
		}

		//主网币转账
		tr := &models.Tx{
			From:        from_,
			To:          to_,
			Contract:    "",
			Amount:      amount_,
			Status:      receipts.Status,
			Hash:        tx.Hash().Hex(),
			ChainId:     *big.NewInt(int64(node.EthClient.GetChainId())),
			Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
			Time:        block.Time(),
			TxType:      TxMainErc,
			BlockNumber: *block.Number(),
			GasPrice:    *tx.GasPrice(),
		}
		if models.HasWallet(from_) {
			if err := models.NewTx(from_, tr); err != nil {
				zlog.Zlog.Info("new mainet error: "+tx.Hash().String(), zap.Error(err))
			}
		}
		if models.HasWallet(to_) {
			if err := models.NewTx(to_, tr); err != nil {
				zlog.Zlog.Info("new mainet error: "+tx.Hash().String(), zap.Error(err))
			}
		}

		if models.HasWallet(from_) || models.HasWallet(to_) {
			//TODO 推送服务
			jsB, _ := json.Marshal(tr)
			w.Wg.Add(1)
			go PushTo(w, jsB)
		}

	} else {
		zlog.Zlog.Info("Get erc20 red packet record....")
		for i := 0; i < len(receipts.Logs); i++ {
			if !(strings.EqualFold(receipts.Logs[i].Topics[0].String(), Transfer)) {
				continue
			}
			if len(receipts.Logs[i].Topics) == 3 {
				//erc20 tx
				to := common.HexToAddress(receipts.Logs[i].Topics[2].Hex())
				amount := new(big.Int)
				amount.SetBytes(receipts.Logs[i].Data)

				txe := &models.Tx{
					From:        from_,
					To:          to.Hex(),
					Contract:    receipts.Logs[i].Address.Hex(),
					Status:      receipts.Status,
					Amount:      *amount,
					Hash:        tx.Hash().Hex(),
					ChainId:     *big.NewInt(int64(node.EthClient.GetChainId())),
					Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
					Time:        block.Time(),
					TxType:      TxMainErc,
					BlockNumber: *block.Number(),
					GasPrice:    *tx.GasPrice(),
				}
				println(from_)
				if models.HasWallet(from_) {
					if err := models.NewTx(from_, txe); err != nil {
						zlog.Zlog.Info("new tx error: "+tx.Hash().String(), zap.Error(err))
					}
				}

				if models.HasWallet(to.Hex()) {
					if err := models.NewTx(to.Hex(), txe); err != nil {
						zlog.Zlog.Info("new tx error: "+tx.Hash().String(), zap.Error(err))
						continue
					}
				}

				if models.HasWallet(from_) || models.HasWallet(to.Hex()) {
					//TODO 推送服务
					jsB, _ := json.Marshal(txe)
					w.Wg.Add(1)
					go PushTo(w, jsB)
				}
			}
		}
	}
	return err
}

func RecordTronRedPacketTransfer(w *Worker, txid *api.BytesMessage, unObj *core.TriggerSmartContract, transactionId string, currentHeight int64, t *core.Transaction_Contract, amount *big.Int, isTrx bool, isCreate bool) error {
	zlog.Zlog.Info("start tron red packet record....")
	var (
		from_ string
		to_   string
		err   error
	)

	if isTrx {
		zlog.Zlog.Info("Get trx red packet record....")
		//主网币转账
		unObj := &core.TransferContract{}
		err := proto.Unmarshal(t.Parameter.GetValue(), unObj)
		if err != nil {
			zlog.Zlog.Error("parse Contract %v err: %v", zap.Error(err))
			return err
		}

		from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
		to := hdwallet.EncodeCheck(unObj.GetToAddress())
		var amount_ big.Int

		if isCreate {
			from_ = from
			to_ = to
			amount_ = *big.NewInt(unObj.GetAmount())
		} else {
			from_ = to
			to_ = from
			amount_ = *amount
		}

		if models.HasWallet(from) || models.HasWallet(to) {
			transinfo, _ := node.TronClient.GetClient().GetTransactionInfoById(context.Background(), txid)
			//成功和失败交易都写进
			tr := &models.Tx{
				From:        from_,
				To:          to_,
				Contract:    "",
				Amount:      amount_,
				Status:      1,
				Hash:        transactionId,
				ChainId:     *big.NewInt(int64(node.TronClient.GetChainId())),
				Gas:         *big.NewInt(transinfo.GetFee()),
				Time:        uint64(transinfo.BlockTimeStamp) / 1000, //换成秒
				TxType:      TxMainErc,
				BlockNumber: *big.NewInt(currentHeight),
				GasPrice:    *big.NewInt(0),
			}
			if models.HasWallet(from_) {
				if err := models.NewTx(from_, tr); err != nil {
					zlog.Zlog.Info("new ftswap error: "+transactionId, zap.Error(err))
				}
			}
			if models.HasWallet(to_) {
				if err := models.NewTx(to_, tr); err != nil {
					zlog.Zlog.Info("new ftswap error: "+transactionId, zap.Error(err))
				}
			}

			if models.HasWallet(from_) || models.HasWallet(to_) {
				//TODO 推送服务
				jsB, _ := json.Marshal(tr)
				w.Wg.Add(1)
				go PushTo(w, jsB)
			}
		}
	} else {
		zlog.Zlog.Info("Get trc20 red packet record....")
		client := node.TronClient
		transinfo, _ := client.GetClient().GetTransactionInfoById(context.Background(), txid)
		from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())

		for _, evenlog := range transinfo.Log {
			if hexutil.Encode(evenlog.Topics[0]) == Transferid {
				if len(evenlog.Topics) == 3 {
					//trc 20 transfer
					amount := new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Data))
					if len(evenlog.Topics[1]) != 32 || len(evenlog.Topics[2]) != 32 {
						continue
					}
					evenlog.Topics[1][11] = 0x41
					evenlog.Topics[2][11] = 0x41
					//from := hdwallet.EncodeCheck(evenlog.Topics[1][11:])
					to := hdwallet.EncodeCheck(evenlog.Topics[2][11:])

					if isCreate {
						from_ = from
						to_ = to
					} else {
						from_ = to
						to_ = from
					}
					token := hdwallet.EncodeCheck(append([]byte{0x41}, evenlog.Address...))

					txe := &models.Tx{
						From:        from_,
						To:          to_,
						Contract:    token,
						Amount:      *amount,
						Status:      1,
						Hash:        transactionId,
						ChainId:     *big.NewInt(int64(client.GetChainId())),
						Gas:         *big.NewInt(transinfo.GetFee()),
						Time:        uint64(transinfo.BlockTimeStamp) / 1000,
						TxType:      TxMainErc,
						BlockNumber: *big.NewInt(currentHeight),
						GasPrice:    *big.NewInt(0),
					}
					if models.HasWallet(from_) {
						if err := models.NewTx(from_, txe); err != nil {
							zlog.Zlog.Info("new tx error: "+transactionId, zap.Error(err))
						}
					}
					if models.HasWallet(to_) {
						if err := models.NewTx(to_, txe); err != nil {
							zlog.Zlog.Info("new tx error: "+transactionId, zap.Error(err))
						}
					}

					if models.HasWallet(from_) || models.HasWallet(to_) {
						//TODO 推送服务
						jsB, _ := json.Marshal(txe)
						w.Wg.Add(1)
						go PushTo(w, jsB)
					}
				}
			}
		}
	}
	return err
}
