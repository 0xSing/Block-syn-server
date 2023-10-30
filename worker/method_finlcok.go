package worker

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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
	unLockMethod       = "6198e339"
	renounceLockMethod = "a57e3141"
	abiS               = `[{"inputs":[{"internalType":"uint256","name":"_fee","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unlockDate","type":"uint256"}],"name":"LockAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"lockId","type":"uint256"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"newOwner","type":"address"}],"name":"LockOwnerChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unlockedAt","type":"uint256"}],"name":"LockRemoved","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"newAmount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"newUnlockDate","type":"uint256"}],"name":"LockUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"inputs":[],"name":"allLocks","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"allLpTokenLockedCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"allNormalTokenLockedCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"cumulativeLockInfo","outputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"lockId","type":"uint256"},{"internalType":"uint256","name":"newAmount","type":"uint256"},{"internalType":"uint256","name":"newUnlockDate","type":"uint256"}],"name":"editLock","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"fee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"start","type":"uint256"},{"internalType":"uint256","name":"end","type":"uint256"}],"name":"getCumulativeLpTokenLockInfo","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getCumulativeLpTokenLockInfoAt","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"start","type":"uint256"},{"internalType":"uint256","name":"end","type":"uint256"}],"name":"getCumulativeNormalTokenLockInfo","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getCumulativeNormalTokenLockInfoAt","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getLockAt","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"start","type":"uint256"},{"internalType":"uint256","name":"end","type":"uint256"}],"name":"getLocksForToken","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getTotalLockCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"token","type":"address"},{"internalType":"bool","name":"isLpToken","type":"bool"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"name":"lock","outputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"lpLockCountForUser","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"lpLockForUserAtIndex","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"lpLocksForUser","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"normalLockCountForUser","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"normalLockForUserAtIndex","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"normalLocksForUser","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"lockId","type":"uint256"}],"name":"renounceLockOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"newFee","type":"uint256"}],"name":"setFee","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"totalLockCountForToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"totalLockCountForUser","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalTokenLockedCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"lockId","type":"uint256"}],"name":"unlock","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"withdrawFee","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
)

var labi, _ = abi.JSON(strings.NewReader(abiS))

// FinLock Tx failed
func EthFinLockTxFailed(w *Worker, tx *types.Transaction) {
	defer w.Wg.Done()
	methodID := common.Bytes2Hex(tx.Data()[0:4])
	if methodID != unLockMethod && methodID != renounceLockMethod {
		return
	}

	receipts, err := node.EthClient.GetClient().TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		zlog.Zlog.Info("Transactionreceipts request fail: "+tx.Hash().String(), zap.Error(err))
		return
	}

	method := ""
	if methodID == unLockMethod {
		method = "unlock"
	} else {
		method = "renounceLockOwnership"
	}

	input, err := labi.Methods[method].Inputs.Unpack(tx.Data()[4:])
	id := input[0].(*big.Int)

	if receipts.Status == 0 {
		var lock models.Lock
		lockStr := models.GetLockById(id.String())
		err := json.Unmarshal([]byte(lockStr), &lock)
		if err != nil {
			zlog.Zlog.Info("Unmarshal data err--------", zap.Error(err))
		}

		err = models.ChangeLockStates(lock.Owner, lock.Id, false)
		if err != nil {
			zlog.Zlog.Info("ChangeLockStates Failed data err--------", zap.Error(err))
		}
	}
}

func TronFinLockTxFailed(w *Worker, unObj *core.TriggerSmartContract, txStatus uint64) {
	defer w.Wg.Done()
	data := unObj.GetData()
	if len(data) < 4 {
		return
	}
	methodID := hexutil.Encode(data[:4])
	if methodID != unLockMethod && methodID != renounceLockMethod {
		return
	}

	method := ""
	if methodID == unLockMethod {
		method = "unlock"
	} else {
		method = "renounceLockOwnership"
	}

	input, err := labi.Methods[method].Inputs.Unpack(data[4:])
	id := input[0].(*big.Int)

	if err != nil {
		zlog.Zlog.Info("unpack data err--------", zap.Error(err))
	}

	if txStatus == 0 {
		var lock models.Lock
		lockStr := models.GetLockById(id.String())
		err := json.Unmarshal([]byte(lockStr), &lock)
		if err != nil {
			zlog.Zlog.Info("Unmarshal data err--------", zap.Error(err))
		}

		err = models.ChangeLockStates(lock.Owner, lock.Id, false)
		if err != nil {
			zlog.Zlog.Info("ChangeLockStates Failed data err--------", zap.Error(err))
		}
	}
}

func RecordEthFinLockTransfer(tx *types.Transaction, block *types.Block, receipts *types.Receipt, isCreate bool) error {
	zlog.Zlog.Info("start eth fin lock record....")
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		zlog.Zlog.Info("get from address fail"+tx.Hash().String(), zap.Error(err))
	}
	// from to 判断
	var (
		from_ string
	)

	if isCreate {
		from_ = from.Hex()
	} else {
		from_ = tx.To().Hex()
	}

	zlog.Zlog.Info("Get erc20 fin lock record....")
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
		}
	}
	return err
}

func RecordTronFinLockTransfer(txid *api.BytesMessage, unObj *core.TriggerSmartContract, transactionId string, currentHeight int64, isCreate bool) error {
	zlog.Zlog.Info("start tron fin lock record....")
	var (
		from_ string
		to_   string
		err   error
	)

	zlog.Zlog.Info("Get trc20 fin lock record....")
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
			}
		}
	}
	return err
}

func RecordEthFinLockFee(tx *types.Transaction, block *types.Block, receipts *types.Receipt) error {
	zlog.Zlog.Info("start eth fin lock fee record....")
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		zlog.Zlog.Info("get from address fail"+tx.Hash().String(), zap.Error(err))
	}

	from_ := from.Hex()
	to_ := tx.To().Hex()

	zlog.Zlog.Info("Get eth fin lock fee record....")
	//主网币转账
	tr := &models.Tx{
		From:        from_,
		To:          to_,
		Contract:    "",
		Amount:      *tx.Value(),
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
	return err
}

func RecordTronFinLockFee(txid *api.BytesMessage, transactionId string, currentHeight int64, t *core.Transaction_Contract) error {
	zlog.Zlog.Info("start tron fin lock fee record....")
	var err error

	zlog.Zlog.Info("Get trx fin lock fee record....")
	//主网币转账
	unObj := &core.TransferContract{}
	err = proto.Unmarshal(t.Parameter.GetValue(), unObj)
	if err != nil {
		zlog.Zlog.Error("parse Contract %v err: %v", zap.Error(err))
		return err
	}

	from := hdwallet.EncodeCheck(unObj.GetOwnerAddress())
	to := hdwallet.EncodeCheck(unObj.GetToAddress())

	if models.HasWallet(from) || models.HasWallet(to) {
		transinfo, _ := node.TronClient.GetClient().GetTransactionInfoById(context.Background(), txid)
		//成功和失败交易都写进
		tr := &models.Tx{
			From:        from,
			To:          to,
			Contract:    "",
			Amount:      *big.NewInt(unObj.GetAmount()),
			Status:      1,
			Hash:        transactionId,
			ChainId:     *big.NewInt(int64(node.TronClient.GetChainId())),
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
	}
	return err
}
