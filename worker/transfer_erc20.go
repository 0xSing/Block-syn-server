package worker

import (
	"context"
	"encoding/json"
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
	TransferMethod = "a9059cbb"
	Transfer       = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	Transferid     = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

func ERC20TransferRecord(w *Worker, tx *types.Transaction, from common.Address, block *types.Block) {
	defer w.Wg.Done()
	client := node.EthClient
	receipts, err := client.GetClient().TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		zlog.Zlog.Info("Transactionreceipts request fail: "+tx.Hash().String(), zap.Error(err))
		return
	}

	var amount big.Int
	var to string
	if receipts.Status == 0 {
		if len(tx.Data()) > 4 {
			if common.Bytes2Hex(tx.Data()[:4]) != TransferMethod {
				return
			}
			to = common.BytesToAddress(tx.Data()[4:36]).Hex()
			amount = *new(big.Int).SetBytes(common.TrimLeftZeroes(tx.Data()[36:]))
		} else {
			amount = *big.NewInt(0)
			to = ""
		}

		//失败转账交易记录
		if models.HasWallet(from.Hex()) {
			txe := &models.Tx{
				From:        from.Hex(),
				To:          to,
				Contract:    tx.To().Hex(),
				Status:      receipts.Status,
				Amount:      amount,
				Hash:        tx.Hash().Hex(),
				ChainId:     *big.NewInt(int64(client.GetChainId())),
				Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
				Time:        block.Time(),
				TxType:      TxMainErc,
				BlockNumber: *block.Number(),
				GasPrice:    *tx.GasPrice(),
			}
			if err := models.NewTx(from.Hex(), txe); err != nil {
				zlog.Zlog.Info("new tx error: "+tx.Hash().String(), zap.Error(err))
			}
		}
		return
	}
	for i := 0; i < len(receipts.Logs); i++ {
		if !(strings.EqualFold(receipts.Logs[i].Topics[0].String(), Transfer) &&
			strings.EqualFold(receipts.Logs[i].Address.String(), tx.To().String())) {
			continue
		}
		if len(receipts.Logs[i].Topics) == 3 {
			//erc20 tx
			to := common.HexToAddress(receipts.Logs[i].Topics[2].Hex())
			amount := new(big.Int)
			amount.SetBytes(receipts.Logs[i].Data)

			txe := &models.Tx{
				From:        from.Hex(),
				To:          to.Hex(),
				Contract:    tx.To().Hex(),
				Status:      receipts.Status,
				Amount:      *amount,
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
					zlog.Zlog.Info("new tx error: "+tx.Hash().String(), zap.Error(err))
				}
			}
			if models.HasWallet(to.Hex()) {
				if err := models.NewTx(to.Hex(), txe); err != nil {
					zlog.Zlog.Info("new tx error: "+tx.Hash().String(), zap.Error(err))
					continue
				}
			}
			if models.HasWallet(to.Hex()) || models.HasWallet(from.Hex()) {
				//推送服务
				jsB, _ := json.Marshal(txe)
				w.Wg.Add(1)
				go PushTo(w, jsB)
			}
		}
	}
}

func TRC20TransferRecord(w *Worker, unObj *core.TriggerSmartContract, txid *api.BytesMessage, txStatus uint64, contract string, transactionId string, currentHeight int64) {
	defer w.Wg.Done()
	client := node.TronClient
	transinfo, _ := client.GetClient().GetTransactionInfoById(context.Background(), txid)
	from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
	data := unObj.GetData()

	if txStatus == 0 && len(data) > 40 && hexutil.Encode(data[:4]) == "a9059cbb" {
		data[15] = 0x41
		to := hdwallet.EncodeCheck(data[15:36])
		amount := new(big.Int).SetBytes(common.TrimLeftZeroes(data[40:]))
		//失败转账交易
		if models.HasWallet(from) {
			txe := &models.Tx{
				From:        from,
				To:          to,
				Contract:    contract,
				Status:      txStatus,
				Amount:      *amount,
				Hash:        transactionId,
				ChainId:     *big.NewInt(int64(client.GetChainId())),
				Gas:         *big.NewInt(transinfo.GetFee()),
				Time:        uint64(transinfo.BlockTimeStamp) / 1000,
				TxType:      TxMainErc,
				BlockNumber: *big.NewInt(currentHeight),
				GasPrice:    *big.NewInt(0),
			}
			if err := models.NewTx(from, txe); err != nil {
				zlog.Zlog.Info("new tx error: "+transactionId, zap.Error(err))
			}
		}
		return
	}

	for _, evenlog := range transinfo.Log {
		if hexutil.Encode(evenlog.Topics[0]) == Transferid &&
			contract == hdwallet.EncodeCheck(append([]byte{0x41}, evenlog.Address...)) {
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
				txe := &models.Tx{
					From:        from,
					To:          to,
					Contract:    contract,
					Amount:      *amount,
					Status:      txStatus,
					Hash:        transactionId,
					ChainId:     *big.NewInt(int64(client.GetChainId())),
					Gas:         *big.NewInt(transinfo.GetFee()),
					Time:        uint64(transinfo.BlockTimeStamp) / 1000,
					TxType:      TxMainErc,
					BlockNumber: *big.NewInt(currentHeight),
					GasPrice:    *big.NewInt(0),
				}
				if models.HasWallet(from) {
					if err := models.NewTx(from, txe); err != nil {
						zlog.Zlog.Info("new tx error: "+transactionId, zap.Error(err))
					}
				}
				if models.HasWallet(to) {
					if err := models.NewTx(to, txe); err != nil {
						zlog.Zlog.Info("new tx error: "+transactionId, zap.Error(err))
					}
				}
				if models.HasWallet(from) || models.HasWallet(to) {
					//推送服务
					jsB, _ := json.Marshal(txe)
					w.Wg.Add(1)
					go PushTo(w, jsB)
				}
			} else if len(evenlog.Topics) == 4 {
				//TODO nft transfer

			}
		}

	}

}
