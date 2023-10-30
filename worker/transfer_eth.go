package worker

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math/big"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	zlog "walletSynV2/utils/zlog_sing"
)

const TxMainErc = 1

func EthTransferRecord(w *Worker, tx *types.Transaction, from common.Address, block *types.Block) {
	defer w.Wg.Done()
	client := node.EthClient
	receipts, err := client.GetClient().TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		zlog.Zlog.Info("Transactionreceipts request fail: "+tx.Hash().String(), zap.Error(err))
		return
	}

	//主网币转账
	tr := &models.Tx{
		From:        from.Hex(),
		To:          tx.To().Hex(),
		Contract:    "",
		Amount:      *tx.Value(),
		Status:      receipts.Status,
		Hash:        tx.Hash().Hex(),
		ChainId:     *big.NewInt(int64(client.GetChainId())),
		Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
		Time:        block.Time(),
		TxType:      TxMainErc,
		BlockNumber: *block.Number(),
		GasPrice:    *tx.GasPrice(),
	}

	if models.HasWallet(from.Hex()) {
		if err := models.NewTx(from.Hex(), tr); err != nil {
			zlog.Zlog.Info("new mainet error: "+tx.Hash().String(), zap.Error(err))
		}
	}
	if models.HasWallet(tx.To().Hex()) {
		if err := models.NewTx(tx.To().Hex(), tr); err != nil {
			zlog.Zlog.Info("new mainet error: "+tx.Hash().String(), zap.Error(err))
		}

	}
	if models.HasWallet(tx.To().Hex()) || models.HasWallet(from.Hex()) {
		jsB, _ := json.Marshal(tr)
		w.Wg.Add(1)
		go PushTo(w, jsB)
	}
}

func TronTransferRecord(w *Worker, txid *api.BytesMessage, txStatus uint64, transactionId string, currentHeight int64, t *core.Transaction_Contract) {
	defer w.Wg.Done()
	client := node.TronClient
	//主网币转账
	unObj := &core.TransferContract{}
	err := proto.Unmarshal(t.Parameter.GetValue(), unObj)
	if err != nil {
		zlog.Zlog.Error("parse Contract %v err: %v", zap.Error(err))
		return
	}

	from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
	to := hdwallet.EncodeCheck(unObj.GetToAddress())
	if models.HasWallet(from) || models.HasWallet(to) {
		transinfo, _ := client.GetClient().GetTransactionInfoById(context.Background(), txid)
		//成功和失败交易都写进
		tr := &models.Tx{
			From:        from,
			To:          to,
			Contract:    "",
			Amount:      *big.NewInt(unObj.GetAmount()),
			Status:      txStatus,
			Hash:        transactionId,
			ChainId:     *big.NewInt(int64(client.GetChainId())),
			Gas:         *big.NewInt(transinfo.GetFee()),
			Time:        uint64(transinfo.BlockTimeStamp) / 1000, //换成秒
			TxType:      TxMainErc,
			BlockNumber: *big.NewInt(currentHeight),
			GasPrice:    *big.NewInt(0),
		}
		if models.HasWallet(from) {
			if err := models.NewTx(from, tr); err != nil {
				zlog.Zlog.Info("new ftswap error: "+transactionId, zap.Error(err))
			}
		}
		if models.HasWallet(to) {
			if err := models.NewTx(to, tr); err != nil {
				zlog.Zlog.Info("new ftswap error: "+transactionId, zap.Error(err))
			}
		}
		if models.HasWallet(from) || models.HasWallet(to) {
			//TODO 推送服务
			jsB, _ := json.Marshal(tr)
			w.Wg.Add(1)
			go PushTo(w, jsB)
		}
	}
}
