package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"math/big"
	"strings"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/service/contract"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	zlog "walletSynV2/utils/zlog_sing"
)

const (
	addLockLogTopic   = "0x694af1cc8727cdd0afbdd53d9b87b69248bd490224e9dd090e788546506e076f"
	editLockLogTopic  = "0xa8b26360df8d5e154ffa5a8a7e894e85f781acfbbef0b744fb9551d8fd0fd36c"
	unLockLogTopic    = "0xc6532367992b32e42a62dd89fc3574876d97d081a263aa6e030f34b79b7e6e8b"
	renounceLockTopic = "0x9075ad040756c0d8743a1fed927066a92c4755071615bf61e04b17583d961caf"

	finLockAbiStr = `[{"inputs":[{"internalType":"uint256","name":"_fee","type":"uint256"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unlockDate","type":"uint256"}],"name":"LockAdded","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"lockId","type":"uint256"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"address","name":"newOwner","type":"address"}],"name":"LockOwnerChanged","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"unlockedAt","type":"uint256"}],"name":"LockRemoved","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"token","type":"address"},{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"uint256","name":"newAmount","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"newUnlockDate","type":"uint256"}],"name":"LockUpdated","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"inputs":[],"name":"allLocks","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"allLpTokenLockedCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"allNormalTokenLockedCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"","type":"address"}],"name":"cumulativeLockInfo","outputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"lockId","type":"uint256"},{"internalType":"uint256","name":"newAmount","type":"uint256"},{"internalType":"uint256","name":"newUnlockDate","type":"uint256"}],"name":"editLock","outputs":[],"stateMutability":"payable","type":"function"},{"inputs":[],"name":"fee","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"start","type":"uint256"},{"internalType":"uint256","name":"end","type":"uint256"}],"name":"getCumulativeLpTokenLockInfo","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getCumulativeLpTokenLockInfoAt","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"start","type":"uint256"},{"internalType":"uint256","name":"end","type":"uint256"}],"name":"getCumulativeNormalTokenLockInfo","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getCumulativeNormalTokenLockInfoAt","outputs":[{"components":[{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"factory","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"internalType":"struct FinLock.CumulativeLockInfo","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"index","type":"uint256"}],"name":"getLockAt","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"},{"internalType":"uint256","name":"start","type":"uint256"},{"internalType":"uint256","name":"end","type":"uint256"}],"name":"getLocksForToken","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getTotalLockCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"token","type":"address"},{"internalType":"bool","name":"isLpToken","type":"bool"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"name":"lock","outputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"lpLockCountForUser","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"lpLockForUserAtIndex","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"lpLocksForUser","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"normalLockCountForUser","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"},{"internalType":"uint256","name":"index","type":"uint256"}],"name":"normalLockForUserAtIndex","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock","name":"","type":"tuple"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"normalLocksForUser","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"token","type":"address"},{"internalType":"address","name":"owner","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"lockDate","type":"uint256"},{"internalType":"uint256","name":"unlockDate","type":"uint256"}],"internalType":"struct FinLock.Lock[]","name":"","type":"tuple[]"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"lockId","type":"uint256"}],"name":"renounceLockOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"newFee","type":"uint256"}],"name":"setFee","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"token","type":"address"}],"name":"totalLockCountForToken","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"user","type":"address"}],"name":"totalLockCountForUser","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalTokenLockedCount","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"lockId","type":"uint256"}],"name":"unlock","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"withdrawFee","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
)

var FinLockABI, _ = abi.JSON(strings.NewReader(finLockAbiStr))

func ListenFinLockEvent(w *Worker, block *types.Block) {
	defer w.Wg.Done()
	//create event logs filter
	client := node.EthClient.GetClient()
	query := ethereum.FilterQuery{
		FromBlock: block.Number(),
		ToBlock:   block.Number(),
		Addresses: []common.Address{
			etc.Conf.EthContract.EthFinLockAddr,
		},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		zlog.Zlog.Error("logs filter err: ", zap.Error(err))
	}
	if len(logs) == 0 {
		return
	}

	fmt.Println("Scanned FinLock event ...")

	// Deal with logs
	for _, blockLog := range logs {
		topic0 := blockLog.Topics[0].Hex()
		tx := block.Transaction(blockLog.TxHash)
		receipt, err := node.EthClient.GetClient().TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			zlog.Zlog.Error("transaction receipt get err: ", zap.Error(err))
		}

		switch topic0 {
		case addLockLogTopic:
			var event models.AddLocked

			err := FinLockABI.UnpackIntoInterface(&event, "LockAdded", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockAdded event explain err: ", zap.Error(err))
			}

			event.Id = new(big.Int).SetBytes(blockLog.Topics[1].Bytes())

			tokenAddr := event.Token.Hex()

			name := contract.Name(tokenAddr)
			decimals := contract.Decimals(tokenAddr)
			symbol := contract.Symbol(tokenAddr)

			lock := &models.Lock{
				Id:         event.Id.String(),
				Token:      tokenAddr,
				Name:       name,
				Symbol:     symbol,
				Decimals:   decimals,
				Owner:      event.Owner.Hex(),
				Amount:     event.Amount.String(),
				LockDate:   block.Time(),
				UnlockDate: event.UnlockDate.Uint64(),
				LockStates: 1,
			}

			err = models.NewLockData(lock)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			// 手续费处理
			err = RecordEthFinLockFee(tx, block, receipt)
			if err != nil {
				zlog.Zlog.Error("Record FinLock Fee operation err: ", zap.Error(err))
			}

			//流水处理
			err = RecordEthFinLockTransfer(tx, block, receipt, true)
			if err != nil {
				zlog.Zlog.Error("Record FinLock Transfer operation err: ", zap.Error(err))
			}

		case editLockLogTopic:
			var event models.LockUpdated

			err := FinLockABI.UnpackIntoInterface(&event, "LockUpdated", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockUpdated event explain err: ", zap.Error(err))
			}
			event.Id = new(big.Int).SetBytes(blockLog.Topics[1].Bytes())

			lock := &models.Lock{
				Id:         event.Id.String(),
				Owner:      event.Owner.Hex(),
				Amount:     event.NewAmount.String(),
				UnlockDate: event.NewUnlockDate.Uint64(),
			}

			err = models.UpdateLockData(lock)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

		case unLockLogTopic:
			var event models.LockRemoved

			err := FinLockABI.UnpackIntoInterface(&event, "LockRemoved", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockRemoved event explain err: ", zap.Error(err))
			}

			event.Id = new(big.Int).SetBytes(blockLog.Topics[1].Bytes())

			lock := &models.Lock{
				Id:         event.Id.String(),
				Owner:      event.Owner.Hex(),
				UnlockDate: block.Time(),
			}

			lockJson, _ := json.Marshal(lock)
			println(string(lockJson))

			err = models.UnlockData(lock)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			//流水处理
			err = RecordEthFinLockTransfer(tx, block, receipt, false)
			if err != nil {
				zlog.Zlog.Error("Record FinLock Transfer operation err: ", zap.Error(err))
			}

		case renounceLockTopic:
			var event models.LockOwnerChanged

			err := FinLockABI.UnpackIntoInterface(&event, "LockOwnerChanged", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockOwnerChanged event explain err: ", zap.Error(err))
			}

			lock := &models.Lock{
				Id:    event.LockId.String(),
				Owner: event.Owner.Hex(),
			}

			err = models.AbandonLock(lock, event.NewOwner.Hex())
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

		default:

			log.Warn("log topics0 didn't match anything")
		}
	}
}

func ListenTronFinLockEvent(w *Worker, txId *api.BytesMessage, unObj *core.TriggerSmartContract, transactionId string, currentHeight int64, t *core.Transaction_Contract) {
	defer w.Wg.Done()
	client := node.TronClient.GetClient()
	transinfo, _ := client.GetTransactionInfoById(context.Background(), txId)
	logs := transinfo.Log

	fmt.Println("Scanned FinLock event ...")

	for _, blockLog := range logs {
		topic0 := hexutil.Encode(blockLog.Topics[0])
		fmt.Println(topic0)

		switch topic0 {
		case addLockLogTopic:
			var event models.AddLocked

			err := FinLockABI.UnpackIntoInterface(&event, "LockAdded", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockAdded event explain err: ", zap.Error(err))
			}

			event.Id = new(big.Int).SetBytes(blockLog.Topics[1])

			owner := addrTransTron(&event.Owner)
			token := addrTransTron(&event.Token)

			name := contract.TronName(token)
			symbol := contract.TronSymbol(token)
			decimals := contract.TronDecimals(token)

			lock := &models.Lock{
				Id:         event.Id.String(),
				Token:      token,
				Name:       name,
				Symbol:     symbol,
				Decimals:   decimals,
				Owner:      owner,
				Amount:     event.Amount.String(),
				LockDate:   uint64(transinfo.BlockTimeStamp / 1000),
				UnlockDate: event.UnlockDate.Uint64(),
				LockStates: 1,
			}

			marshal, _ := json.Marshal(lock)
			fmt.Println(string(marshal))

			err = models.NewLockData(lock)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			//手续费处理
			err = RecordTronFinLockFee(txId, transactionId, currentHeight, t)
			if err != nil {
				zlog.Zlog.Error("Record FinLock Fee operation err: ", zap.Error(err))
			}

			//处理流水
			err = RecordTronFinLockTransfer(txId, unObj, transactionId, currentHeight, true)
			if err != nil {
				zlog.Zlog.Error("Record FinLock Transfer operation err: ", zap.Error(err))
			}

		case editLockLogTopic:
			var event models.LockUpdated

			err := FinLockABI.UnpackIntoInterface(&event, "LockUpdated", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockUpdated event explain err: ", zap.Error(err))
			}
			event.Id = new(big.Int).SetBytes(blockLog.Topics[1])

			owner := addrTransTron(&event.Owner)

			lock := &models.Lock{
				Id:         event.Id.String(),
				Owner:      owner,
				Amount:     event.NewAmount.String(),
				UnlockDate: event.NewUnlockDate.Uint64(),
			}

			err = models.UpdateLockData(lock)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

		case unLockLogTopic:
			var event models.LockRemoved

			err := FinLockABI.UnpackIntoInterface(&event, "LockRemoved", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockRemoved event explain err: ", zap.Error(err))
			}

			event.Id = new(big.Int).SetBytes(blockLog.Topics[1])

			owner := addrTransTron(&event.Owner)

			lock := &models.Lock{
				Id:         event.Id.String(),
				Owner:      owner,
				UnlockDate: uint64(transinfo.BlockTimeStamp / 1000),
			}

			lockJson, _ := json.Marshal(lock)
			println(string(lockJson))

			err = models.UnlockData(lock)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			//处理流水
			err = RecordTronFinLockTransfer(txId, unObj, transactionId, currentHeight, false)
			if err != nil {
				zlog.Zlog.Error("Record FinLock Transfer operation err: ", zap.Error(err))
			}

		case renounceLockTopic:
			var event models.LockOwnerChanged

			err := FinLockABI.UnpackIntoInterface(&event, "LockOwnerChanged", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("LockOwnerChanged event explain err: ", zap.Error(err))
			}

			owner := addrTransTron(&event.Owner)
			newOwner := addrTransTron(&event.NewOwner)

			lock := &models.Lock{
				Id:    event.LockId.String(),
				Owner: owner,
			}

			err = models.AbandonLock(lock, newOwner)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

		default:
			log.Warn("Log topics0 didn't match anything")
		}
	}
}

func addrTransTron(addr *common.Address) string {
	bytes := addr.Hash().Bytes()
	bytes[11] = 0x41
	tronAddr := hdwallet.EncodeCheck(bytes[11:])
	return tronAddr
}
