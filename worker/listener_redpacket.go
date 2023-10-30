package worker

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/zap"
	"math/big"
	"strconv"
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
	createPacketTopic = "0x52871f59806bc02f686fb16ab6f02e341def69c8c47d20b9e716fa69b3eb6380"
	claimPacketTopic  = "0xaabd1ce8d77185358f8d53064d3f903d11fdc3b10c7a0865c9cda2871981e5a9"
	ownerRefundTopic  = "0x376aba31ca2f718ed24fa28bab1423390ec32e6423afc7d61e74bc5a20a84339"

	zeroAddr        = "0x0000000000000000000000000000000000000000"
	redPacketAbiStr = `[{"inputs":[{"internalType":"address","name":"newPublicKey","type":"address"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"claimer","type":"address"},{"indexed":false,"internalType":"uint256","name":"claimedValue","type":"uint256"},{"indexed":false,"internalType":"address","name":"tokenAddress","type":"address"}],"name":"ClaimSuccess","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"total","type":"uint256"},{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"string","name":"name","type":"string"},{"indexed":false,"internalType":"string","name":"message","type":"string"},{"indexed":false,"internalType":"address","name":"creator","type":"address"},{"indexed":false,"internalType":"uint256","name":"creationTime","type":"uint256"},{"indexed":false,"internalType":"address","name":"tokenAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"number","type":"uint256"},{"indexed":false,"internalType":"bool","name":"ifrandom","type":"bool"},{"indexed":false,"internalType":"uint256","name":"duration","type":"uint256"}],"name":"CreationSuccess","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"},{"indexed":false,"internalType":"address","name":"tokenAddress","type":"address"},{"indexed":false,"internalType":"uint256","name":"remainingBalance","type":"uint256"}],"name":"RefundSuccess","type":"event"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"name":"checkAvailability","outputs":[{"internalType":"address","name":"tokenAddress","type":"address"},{"internalType":"uint256","name":"remainingTokens","type":"uint256"},{"internalType":"uint256","name":"totalNumber","type":"uint256"},{"internalType":"uint256","name":"claimedNumber","type":"uint256"},{"internalType":"uint256","name":"ifrandom","type":"uint256"},{"internalType":"bool","name":"expired","type":"bool"},{"internalType":"uint256","name":"claimedAmount","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"uint256","name":"receivedAmount","type":"uint256"},{"internalType":"bytes","name":"signedMsg","type":"bytes"}],"name":"claim","outputs":[{"internalType":"uint256","name":"claimed","type":"uint256"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"number","type":"uint256"},{"internalType":"bool","name":"ifrandom","type":"bool"},{"internalType":"uint256","name":"duration","type":"uint256"},{"internalType":"string","name":"_message","type":"string"},{"internalType":"string","name":"_name","type":"string"},{"internalType":"uint256","name":"tokenType","type":"uint256"},{"internalType":"address","name":"tokenAddr","type":"address"},{"internalType":"uint256","name":"totalTokens","type":"uint256"}],"name":"createRedPacket","outputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"stateMutability":"payable","type":"function"},{"inputs":[{"internalType":"uint256","name":"startIndex","type":"uint256"},{"internalType":"uint256","name":"endIndex","type":"uint256"}],"name":"getExpiredPackets","outputs":[{"components":[{"internalType":"uint256","name":"id","type":"uint256"},{"internalType":"address","name":"tokenAddr","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"uint256","name":"tokenType","type":"uint256"}],"internalType":"struct FinRedPacket.ExpiredPacket[]","name":"","type":"tuple[]"},{"internalType":"uint256","name":"validIndex","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"getRedPacketNum","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"id","type":"uint256"}],"name":"ownerRefund","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"startIndex","type":"uint256"},{"internalType":"uint256","name":"endIndex","type":"uint256"}],"name":"ownerRefundRange","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"uint256","name":"","type":"uint256"}],"name":"redpacketById","outputs":[{"components":[{"internalType":"uint256","name":"packed1","type":"uint256"},{"internalType":"uint256","name":"packed2","type":"uint256"}],"internalType":"struct FinRedPacket.Packed","name":"packed","type":"tuple"},{"internalType":"address","name":"publicKey","type":"address"},{"internalType":"address","name":"creator","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newPublicKey","type":"address"}],"name":"setPublicKey","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
)

var redPacketABI, _ = abi.JSON(strings.NewReader(redPacketAbiStr))

func ListenFinRedPacketEvent(w *Worker, block *types.Block) {
	defer w.Wg.Done()
	query := ethereum.FilterQuery{
		FromBlock: block.Number(),
		ToBlock:   block.Number(),
		Addresses: []common.Address{
			etc.Conf.EthContract.EthRedPacketAddr,
		},
	}

	logs, err := node.EthClient.GetClient().FilterLogs(context.Background(), query)
	if err != nil {
		zlog.Zlog.Error("logs filter err: ", zap.Error(err))
	}

	for _, eventLog := range logs {
		topic0 := eventLog.Topics[0].Hex()
		tx := block.Transaction(eventLog.TxHash)
		receipt, err := node.EthClient.GetClient().TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			zlog.Zlog.Error("transaction receipt get err: ", zap.Error(err))
		}

		//  使用 switch case
		switch topic0 {
		case createPacketTopic:
			var event models.CreationSuccess
			err := redPacketABI.UnpackIntoInterface(&event, "CreationSuccess", eventLog.Data)
			if err != nil {
				zlog.Zlog.Error("CreationSuccess event explain err: ", zap.Error(err))
			}

			redPacket := &models.RedPacket{
				Id:              event.Id.String(),
				Name:            event.Name,
				Hash:            tx.Hash().Hex(),
				Message:         event.Message,
				Creator:         event.Creator.String(),
				CreationTime:    block.Time(),
				TotalAmount:     event.Total.String(),
				TotalNumber:     event.Number.Uint64(),
				RemainingAmount: event.Total.String(),
				RemainingNumber: event.Number.Uint64(),
				IfRandom:        event.Ifrandom,
				TokenAddress:    event.TokenAddress.String(),
				Duration:        event.Duration.Uint64(),
			}

			var receivers []models.ReceiveRedPacketRecord

			sendRedPacketRecord := &models.SendRedPacketRecord{
				RelationId:   redPacket.Id,
				TokenAddress: redPacket.TokenAddress,
				TotalAmount:  redPacket.TotalAmount,
				TotalNumber:  redPacket.TotalNumber,
				Fee:          new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed))).String(),
				SendTime:     redPacket.CreationTime,
				Sender:       redPacket.Creator,
				Receivers:    receivers,
			}

			redPacketComm := &models.RedPacketCommunication{
				RedPacketNo: redPacket.Id,
				WalletAddr:  redPacket.Creator,
				ChainId:     tx.ChainId().Uint64(),
				Message:     redPacket.Message,
				Hash:        redPacket.Hash,
				Contract:    redPacket.TokenAddress,
				Number:      redPacket.TotalNumber,
				Amount:      redPacket.TotalAmount,
				Fee:         sendRedPacketRecord.Fee,
				StartTime:   redPacket.CreationTime,
				EndTime:     redPacket.CreationTime + redPacket.Duration,
				Type:        1,
			}

			//数据库处理
			err = models.NewRedPacket(redPacket)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			err = models.NewSendRedPacketRecord(sendRedPacketRecord)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			//流水处理
			isEth := strings.EqualFold(zeroAddr, sendRedPacketRecord.TokenAddress)
			err = RecordEthRedPacketTransfer(w, tx, block, receipt, big.NewInt(0), isEth, true)
			if err != nil {
				zlog.Zlog.Error("Record RedPacket Transfer operation err: ", zap.Error(err))
			}

			//data comm
			RedPacketCommunication(redPacketComm)

		case claimPacketTopic:
			var event models.ClaimSuccess
			err := redPacketABI.UnpackIntoInterface(&event, "ClaimSuccess", eventLog.Data)
			if err != nil {
				zlog.Zlog.Error("ClaimSuccess event explain err: ", zap.Error(err))
			}

			receiveRecord := &models.ReceiveRedPacketRecord{
				RelationId:     event.Id.String(),
				Receiver:       event.Claimer.String(),
				TokenAddress:   event.TokenAddress.String(),
				ReceivedAmount: event.ClaimedValue.String(),
				ReceiveTime:    block.Time(),
				Fee:            new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed))).String(),
			}

			//数据库处理
			err = models.ClaimRPDataUpdate(receiveRecord)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			redPacketComm := &models.RedPacketCommunication{
				RedPacketNo:   receiveRecord.RelationId,
				WalletAddr:    receiveRecord.Receiver,
				ChainId:       tx.ChainId().Uint64(),
				Hash:          tx.Hash().String(),
				Contract:      receiveRecord.TokenAddress,
				Number:        1,
				Amount:        receiveRecord.ReceivedAmount,
				Fee:           receiveRecord.Fee,
				OperationTime: receiveRecord.ReceiveTime,
				Type:          2,
			}

			//流水处理
			isEth := strings.EqualFold(zeroAddr, receiveRecord.TokenAddress)
			err = RecordEthRedPacketTransfer(w, tx, block, receipt, event.ClaimedValue, isEth, false)
			if err != nil {
				zlog.Zlog.Error("Record RedPacket Transfer operation err: ", zap.Error(err))
			}

			//data comm
			RedPacketCommunication(redPacketComm)

		case ownerRefundTopic:
			var event models.RefundSuccess
			err := redPacketABI.UnpackIntoInterface(&event, "RefundSuccess", eventLog.Data)
			if err != nil {
				zlog.Zlog.Error("RefundSuccess event explain err: ", zap.Error(err))
			}

			refundRecord := &models.RefundRecord{
				RelationId:     event.Id.String(),
				TokenAddress:   event.TokenAddress.String(),
				ReceivedAmount: event.RemainingBalance.String(),
				ReceiveTime:    block.Time(),
			}

			// 数据库处理
			err = models.OwnerRefundUpdate(refundRecord)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			redPacketComm := &models.RedPacketCommunication{
				RedPacketNo:   refundRecord.RelationId,
				WalletAddr:    contract.RedPacketOwner(),
				ChainId:       tx.ChainId().Uint64(),
				Hash:          tx.Hash().String(),
				Contract:      refundRecord.TokenAddress,
				Number:        1,
				Amount:        refundRecord.ReceivedAmount,
				Fee:           new(big.Int).Mul(tx.GasPrice(), big.NewInt(int64(receipt.GasUsed))).String(),
				OperationTime: refundRecord.ReceiveTime,
				Type:          2,
			}

			//流水处理
			isEth := strings.EqualFold(zeroAddr, refundRecord.TokenAddress)
			err = RecordEthRedPacketTransfer(w, tx, block, receipt, event.RemainingBalance, isEth, false)
			if err != nil {
				zlog.Zlog.Error("Record RedPacket Transfer operation err: ", zap.Error(err))
			}

			//data comm
			RedPacketCommunication(redPacketComm)

		default:

			log.Warn("log topics0 didn't match anything")
		}
	}

}

func ListenTronRedPacketEvent(w *Worker, txId *api.BytesMessage, unObj *core.TriggerSmartContract, transactionId string, currentHeight int64, t *core.Transaction_Contract) {
	defer w.Wg.Done()
	client := node.TronClient.GetClient()
	transinfo, _ := client.GetTransactionInfoById(context.Background(), txId)
	logs := transinfo.Log

	fmt.Println("Scanned FinRedPacket event ...")

	for _, blockLog := range logs {
		topic0 := hexutil.Encode(blockLog.Topics[0])
		fmt.Println(topic0)

		switch topic0 {
		case createPacketTopic:
			var event models.CreationSuccess

			err := redPacketABI.UnpackIntoInterface(&event, "CreationSuccess", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("CreationSuccess event explain err: ", zap.Error(err))
			}

			creator := addrTransTron(&event.Creator)
			var tokenAddr string
			if event.TokenAddress.Hex() != zeroAddr {
				tokenAddr = addrTransTron(&event.TokenAddress)
			} else {
				tokenAddr = zeroAddr
			}

			redPacket := &models.RedPacket{
				Id:              event.Id.String(),
				Hash:            hexutil.Encode(txId.Value),
				Name:            event.Name,
				Message:         event.Message,
				Creator:         creator,
				CreationTime:    uint64(transinfo.BlockTimeStamp / 1000),
				TotalAmount:     event.Total.String(),
				TotalNumber:     event.Number.Uint64(),
				RemainingAmount: event.Total.String(),
				RemainingNumber: event.Number.Uint64(),
				IfRandom:        event.Ifrandom,
				TokenAddress:    tokenAddr,
				Duration:        event.Duration.Uint64(),
			}

			var receivers []models.ReceiveRedPacketRecord

			sendRedPacketRecord := &models.SendRedPacketRecord{
				RelationId:   redPacket.Id,
				TokenAddress: redPacket.TokenAddress,
				TotalAmount:  redPacket.TotalAmount,
				TotalNumber:  redPacket.TotalNumber,
				Fee:          strconv.FormatInt(transinfo.Fee, 10),
				SendTime:     redPacket.CreationTime,
				Sender:       redPacket.Creator,
				Receivers:    receivers,
			}

			redPacketComm := &models.RedPacketCommunication{
				RedPacketNo: redPacket.Id,
				WalletAddr:  redPacket.Creator,
				ChainId:     node.TronClient.GetChainId(),
				Message:     redPacket.Message,
				Hash:        redPacket.Hash,
				Contract:    redPacket.TokenAddress,
				Number:      redPacket.TotalNumber,
				Amount:      redPacket.TotalAmount,
				Fee:         sendRedPacketRecord.Fee,
				StartTime:   redPacket.CreationTime,
				EndTime:     redPacket.CreationTime + redPacket.Duration,
				Type:        1,
			}

			//数据库处理
			err = models.NewRedPacket(redPacket)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			err = models.NewSendRedPacketRecord(sendRedPacketRecord)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			//处理流水
			isTrx := strings.EqualFold(zeroAddr, sendRedPacketRecord.TokenAddress)
			err = RecordTronRedPacketTransfer(w, txId, unObj, transactionId, currentHeight, t, big.NewInt(0), isTrx, true)
			if err != nil {
				zlog.Zlog.Error("Record RedPacket Transfer operation err: ", zap.Error(err))
			}

			//data comm
			RedPacketCommunication(redPacketComm)

		case claimPacketTopic:
			var event models.ClaimSuccess
			err := redPacketABI.UnpackIntoInterface(&event, "ClaimSuccess", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("ClaimSuccess event explain err: ", zap.Error(err))
			}

			creator := addrTransTron(&event.Claimer)
			var tokenAddr string
			if event.TokenAddress.Hex() != zeroAddr {
				tokenAddr = addrTransTron(&event.TokenAddress)
			} else {
				tokenAddr = zeroAddr
			}

			receiveRecord := &models.ReceiveRedPacketRecord{
				RelationId:     event.Id.String(),
				Receiver:       creator,
				TokenAddress:   tokenAddr,
				ReceivedAmount: event.ClaimedValue.String(),
				ReceiveTime:    uint64(transinfo.BlockTimeStamp / 1000),
				Fee:            strconv.FormatInt(transinfo.Fee, 10),
			}

			//数据库处理
			err = models.ClaimRPDataUpdate(receiveRecord)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			redPacketComm := &models.RedPacketCommunication{
				RedPacketNo:   receiveRecord.RelationId,
				WalletAddr:    receiveRecord.Receiver,
				ChainId:       node.TronClient.GetChainId(),
				Hash:          hexutil.Encode(txId.Value),
				Contract:      receiveRecord.TokenAddress,
				Number:        1,
				Amount:        receiveRecord.ReceivedAmount,
				Fee:           receiveRecord.Fee,
				OperationTime: receiveRecord.ReceiveTime,
				Type:          2,
			}

			//处理流水
			isTrx := strings.EqualFold(zeroAddr, receiveRecord.TokenAddress)
			err = RecordTronRedPacketTransfer(w, txId, unObj, transactionId, currentHeight, t, event.ClaimedValue, isTrx, false)
			if err != nil {
				zlog.Zlog.Error("Record RedPacket Transfer operation err: ", zap.Error(err))
			}

			//data comm
			RedPacketCommunication(redPacketComm)

		case ownerRefundTopic:
			var event models.RefundSuccess
			err := redPacketABI.UnpackIntoInterface(&event, "RefundSuccess", blockLog.Data)
			if err != nil {
				zlog.Zlog.Error("RefundSuccess event explain err: ", zap.Error(err))
			}

			var tokenAddr string
			if event.TokenAddress.Hex() != zeroAddr {
				tokenAddr = addrTransTron(&event.TokenAddress)
			} else {
				tokenAddr = zeroAddr
			}

			refundRecord := &models.RefundRecord{
				RelationId:     event.Id.String(),
				TokenAddress:   tokenAddr,
				ReceivedAmount: event.RemainingBalance.String(),
				ReceiveTime:    uint64(transinfo.BlockTimeStamp / 1000),
			}

			// 数据库处理
			err = models.OwnerRefundUpdate(refundRecord)
			if err != nil {
				zlog.Zlog.Error("leveldb operation err: ", zap.Error(err))
			}

			redPacketComm := &models.RedPacketCommunication{
				RedPacketNo:   refundRecord.RelationId,
				WalletAddr:    contract.RedPacketOwner(),
				ChainId:       node.TronClient.GetChainId(),
				Hash:          hexutil.Encode(txId.Value),
				Contract:      refundRecord.TokenAddress,
				Number:        1,
				Amount:        refundRecord.ReceivedAmount,
				Fee:           strconv.FormatInt(transinfo.Fee, 10),
				OperationTime: refundRecord.ReceiveTime,
				Type:          2,
			}

			//处理流水
			isTrx := strings.EqualFold(zeroAddr, refundRecord.TokenAddress)
			err = RecordTronRedPacketTransfer(w, txId, unObj, transactionId, currentHeight, t, event.RemainingBalance, isTrx, false)
			if err != nil {
				zlog.Zlog.Error("Record RedPacket Transfer operation err: ", zap.Error(err))
			}

			//data comm
			RedPacketCommunication(redPacketComm)

		default:
			log.Warn("Log topics0 didn't match anything")
		}
	}
}
