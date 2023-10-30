package worker

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"math/big"
	"strings"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	"walletSynV2/utils/tron/hexutil"
	zlog "walletSynV2/utils/zlog_sing"
)

const (
	batchTransferTopic     = "0x05010e30e029c91a494389a51200ece68eaa2bc42a7b648536c43ac1f660037a"
	tronBatchTransferTopic = "05010e30e029c91a494389a51200ece68eaa2bc42a7b648536c43ac1f660037a"
	batchAbi               = `[{"inputs":[],"name":"claim","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address[]","name":"from","type":"address[]"},{"indexed":false,"internalType":"uint256[]","name":"value","type":"uint256[]"},{"indexed":false,"internalType":"address","name":"token","type":"address"}],"name":"EtherBatchTransfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"EtherClaim","type":"event"},{"inputs":[{"internalType":"address[]","name":"recipients","type":"address[]"},{"internalType":"uint256[]","name":"values","type":"uint256[]"},{"internalType":"address","name":"token","type":"address"}],"name":"send","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"_p","type":"uint256"}],"name":"setP","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"_owner","type":"address"}],"name":"transferOwner","outputs":[],"stateMutability":"nonpayable","type":"function"},{"stateMutability":"payable","type":"receive"}]`
)

func EthBatchTransfer(w *Worker, tx *types.Transaction, from common.Address, block *types.Block) {
	defer w.Wg.Done()
	client := node.EthClient
	receipts, err := client.GetClient().TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		zlog.Zlog.Info("get BT receipt error: " + err.Error())
		return
	}
	btabi, err := abi.JSON(strings.NewReader(batchAbi))
	if err != nil {
		zlog.Zlog.Info("bt abi json error" + err.Error())
		return
	}
	//dataInput, err := btabi.Unpack("send", tx.Data()[4:])
	dataInput, err := btabi.Methods["send"].Inputs.Unpack(tx.Data()[4:])
	if err != nil {
		zlog.Zlog.Info("decode input data error:" + err.Error())
		return
	}
	ct := dataInput[2].(common.Address)
	contract := ct.Hex()
	if ct.Hex() == "0x0000000000000000000000000000000000000000" {
		contract = ""
	}
	if receipts.Status == 0 {
		//失败交易
		txe := &models.Tx{
			From:        from.Hex(),
			To:          "",
			Contract:    contract,
			Status:      receipts.Status,
			Amount:      *big.NewInt(0),
			Hash:        tx.Hash().Hex(),
			ChainId:     *big.NewInt(int64(client.GetChainId())),
			Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
			Time:        block.Time(),
			TxType:      TxMainErc,
			BlockNumber: *block.Number(),
			GasPrice:    *tx.GasPrice(),
		}
		if models.HasWallet(from.Hex()) {
			if err := models.NewTx(from.Hex(), txe); err != nil {
				zlog.Zlog.Info("new bt tx to error: "+tx.Hash().String(), zap.Error(err))
			}
		}
		return
	}

	for _, l := range receipts.Logs {
		if len(l.Topics) == 0 {
			zlog.Zlog.Error("this log didn't have topics： " + l.TxHash.Hex())
			continue
		}

		if !strings.EqualFold(l.Topics[0].Hex(), batchTransferTopic) {
			continue
		}

		data, err := btabi.Events["EtherBatchTransfer"].Inputs.Unpack(l.Data)
		if err != nil {
			zlog.Zlog.Info("unpack bt abi error" + err.Error())
			continue
		}
		tos := data[0].([]common.Address)
		amounts := data[1].([]*big.Int)

		for i, to := range tos {
			txe := &models.Tx{
				From:        from.Hex(),
				To:          to.Hex(),
				Contract:    contract,
				Status:      receipts.Status,
				Amount:      *amounts[i],
				Hash:        tx.Hash().Hex(),
				ChainId:     *big.NewInt(int64(client.GetChainId())),
				Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
				Time:        block.Time(),
				TxType:      TxMainErc,
				BlockNumber: *block.Number(),
				GasPrice:    *tx.GasPrice(),
			}
			if models.HasWallet(to.Hex()) {
				if err := models.NewTx(to.Hex(), txe); err != nil {
					zlog.Zlog.Info("new bt tx to error: "+tx.Hash().String(), zap.Error(err))
				}
			}
			if models.HasWallet(from.Hex()) {
				if err := models.NewTx(from.Hex(), txe); err != nil {
					zlog.Zlog.Info("new bt tx to error: "+tx.Hash().String(), zap.Error(err))
				}
			}
			//无论在不在数据库，都推送
			jsB, _ := json.Marshal(txe)
			w.Wg.Add(1)
			go PushTo(w, jsB)
		}
	}
}

func TronScanBatchTransfer(w *Worker, txId *api.BytesMessage, txStatus uint64, unObj *core.TriggerSmartContract) {
	defer w.Wg.Done()
	client := node.TronClient
	transinfo, _ := client.GetClient().GetTransactionInfoById(context.Background(), txId)
	logs := transinfo.Log
	from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())

	txData := unObj.GetData()
	btabi, err := abi.JSON(strings.NewReader(batchAbi))
	if err != nil {
		zlog.Zlog.Info("bt abi json error", zap.Error(err))
		return
	}
	dataInput, err := btabi.Methods["send"].Inputs.Unpack(txData[4:])
	ct := dataInput[2].(common.Address)
	contract := hdwallet.EncodeCheck(append([]byte{0x41}, ct.Bytes()...))
	if ct.Hex() == "0x0000000000000000000000000000000000000000" {
		contract = ""
	}

	if txStatus == 0 {
		//失败交易
		txe := &models.Tx{
			From:        from,
			To:          "",
			Contract:    contract,
			Status:      txStatus,
			Amount:      *big.NewInt(0),
			Hash:        hexutil.Encode(txId.Value),
			ChainId:     *big.NewInt(int64(client.GetChainId())),
			Gas:         *big.NewInt(transinfo.GetFee()),
			Time:        uint64(transinfo.BlockTimeStamp) / 1000,
			TxType:      TxMainErc,
			BlockNumber: *big.NewInt(transinfo.BlockNumber),
			GasPrice:    *big.NewInt(0),
		}
		if models.HasWallet(from) {
			if err := models.NewTx(from, txe); err != nil {
				zlog.Zlog.Info("new bt tx to error: "+hexutil.Encode(txId.Value), zap.Error(err))
			}
		}
	}

	for _, l := range logs {
		if !strings.EqualFold(hexutil.Encode(l.Topics[0]), tronBatchTransferTopic) {
			continue
		}

		data, err := btabi.Events["EtherBatchTransfer"].Inputs.Unpack(l.Data)
		if err != nil {
			zlog.Zlog.Info("unpack bt abi error", zap.Error(err))
			continue
		}
		tos := data[0].([]common.Address)
		amounts := data[1].([]*big.Int)
		var total big.Int
		for i, to := range tos {
			tTron := hdwallet.EncodeCheck(append([]byte{0x41}, to.Bytes()...))
			total.Add(&total, amounts[i])

			txe := &models.Tx{
				From:        from,
				To:          tTron,
				Contract:    contract,
				Status:      txStatus,
				Amount:      *amounts[i],
				Hash:        hexutil.Encode(txId.Value),
				ChainId:     *big.NewInt(int64(client.GetChainId())),
				Gas:         *big.NewInt(transinfo.GetFee()),
				Time:        uint64(transinfo.BlockTimeStamp) / 1000,
				TxType:      TxMainErc,
				BlockNumber: *big.NewInt(transinfo.BlockNumber),
				GasPrice:    *big.NewInt(0),
			}
			if models.HasWallet(tTron) {
				if err := models.NewTx(tTron, txe); err != nil {
					zlog.Zlog.Info("new bt tx to error: "+hexutil.Encode(txId.Value), zap.Error(err))
				}
			}
			if models.HasWallet(from) {
				if err := models.NewTx(from, txe); err != nil {
					zlog.Zlog.Info("new bt tx to error: "+hexutil.Encode(txId.Value), zap.Error(err))
				}
			}
			//推送
			jsB, _ := json.Marshal(txe)
			w.Wg.Add(1)
			go PushTo(w, jsB)
		}

	}
}
