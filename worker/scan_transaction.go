package worker

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"strings"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	"walletSynV2/utils/tron/hexutil"
	zlog "walletSynV2/utils/zlog_sing"
)

func EthTransactions(w *Worker, block *types.Block) {
	defer w.Wg.Done()
	for _, tx := range block.Transactions() {
		if tx.To() == nil { //不是合约调用
			continue
		}

		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			zlog.Zlog.Info("get from address fail"+tx.Hash().String(), zap.Error(err))
			continue
		}

		if strings.EqualFold(tx.To().Hex(), etc.Conf.EthContract.EthFTSwapAddr.Hex()) &&
			tx.To().Hex() != "0x0000000000000000000000000000000000000000" {
			w.Wg.Add(1)
			go EthFlSwapScan(w, &from, block.Time(), block.Number(), tx)
			continue
		}

		if strings.EqualFold(tx.To().Hex(), etc.Conf.EthContract.EthFinLockAddr.Hex()) &&
			tx.To().Hex() != "0x0000000000000000000000000000000000000000" {
			// failed tx scan
			w.Wg.Add(1)
			go EthFinLockTxFailed(w, tx)
			continue
		}

		if strings.EqualFold(tx.To().Hex(), etc.Conf.EthContract.EthBatchTransferAddr.Hex()) &&
			tx.To().Hex() != "0x0000000000000000000000000000000000000000" {
			w.Wg.Add(1)
			go EthBatchTransfer(w, tx, from, block)
			continue
		}

		if tx.Value().Sign() > 0 &&
			len(tx.Data()) < 4 &&
			(models.HasWallet(tx.To().Hex()) || models.HasWallet(from.Hex())) {
			w.Wg.Add(1)
			go EthTransferRecord(w, tx, from, block)
		}

		if models.HasContract(tx.To().Hex()) {
			w.Wg.Add(1)
			go ERC20TransferRecord(w, tx, from, block)
		}
	}
}

func TronTransactions(w *Worker, block *api.BlockExtention, currentHeight int64) {
	defer w.Wg.Done()
	for _, tx := range block.Transactions {
		transactionId := hexutil.Encode(tx.Txid)
		txid := new(api.BytesMessage)
		decode, err := hexutil.Decode(transactionId)
		if err != nil {
			zlog.Zlog.Info("get tx hash fail", zap.Error(err))
			continue
		}

		txid.Value = decode

		rets := tx.Transaction.Ret
		//应该只有一个
		for _, t := range tx.Transaction.RawData.Contract {
			txStatus := uint64(1)
			if len(rets) < 1 || rets[0].ContractRet != core.Transaction_Result_SUCCESS {
				//失败交易
				txStatus = 0
				//continue
			}

			if t.Type == core.Transaction_Contract_TransferContract {
				w.Wg.Add(1)
				go TronTransferRecord(w, txid, txStatus, transactionId, currentHeight, t)

			} else if t.Type == core.Transaction_Contract_TriggerSmartContract { //调用智能合约
				unObj := &core.TriggerSmartContract{}
				err := proto.Unmarshal(t.Parameter.GetValue(), unObj)
				if err != nil {
					zlog.Zlog.Error("parse Contract %v err: "+t.Type.String(), zap.Error(err))
					continue
				}
				contract := hdwallet.EncodeCheck(unObj.GetContractAddress())

				// swap scan
				if strings.EqualFold(contract, etc.Conf.Contract.FTSwapAddr) {
					w.Wg.Add(1)
					go TronFlSwapScan(w, unObj, txid, txStatus, contract, transactionId, currentHeight)
					continue
				}

				// listener finlock logs or finlock tx fail
				if strings.EqualFold(contract, etc.Conf.Contract.FinLockAddr) {
					w.Wg.Add(2)
					go ListenTronFinLockEvent(w, txid, unObj, transactionId, currentHeight, t)
					go TronFinLockTxFailed(w, unObj, txStatus)
					continue
				}

				// listener redpacket logs
				if strings.EqualFold(contract, etc.Conf.Contract.RedPacketAddr) {
					w.Wg.Add(1)
					go ListenTronRedPacketEvent(w, txid, unObj, transactionId, currentHeight, t)
					continue
				}

				// scan batch transfer
				if strings.EqualFold(contract, etc.Conf.Contract.BatchTransferAddr) {
					w.Wg.Add(1)
					go TronScanBatchTransfer(w, txid, txStatus, unObj)
					continue
				}

				// scan trc20 token transfer
				if !models.HasContract(contract) {
					continue
				}
				w.Wg.Add(1)
				go TRC20TransferRecord(w, unObj, txid, txStatus, contract, transactionId, currentHeight)

			}

		}

	}
}

func WorkerPushTo(w *Worker, js []byte) {
	PushTo(w, js)
}
