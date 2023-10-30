package models

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
	"walletSynV2/database"
	"walletSynV2/utils"
	zlog "walletSynV2/utils/zlog_sing"
)

var (
	redPacketLock = &utils.DBKeyLock{}
)

type RedPacket struct {
	Id              string `json:"id"`
	Hash            string `json:"hash"`
	Name            string `json:"name"`
	Message         string `json:"message"`
	Creator         string `json:"creator"`
	CreationTime    uint64 `json:"creationTime"`
	TotalAmount     string `json:"totalAmount"`
	TotalNumber     uint64 `json:"totalNumber"`
	RemainingAmount string `json:"remainingAmount"`
	RemainingNumber uint64 `json:"remainingNumber"`
	IfRandom        bool   `json:"ifRandom"`
	TokenAddress    string `json:"tokenAddress"`
	Duration        uint64 `json:"duration"`
}

type RedPacketCommunication struct {
	RedPacketNo   string `json:"redPacketNo"`
	WalletAddr    string `json:"walletAddr"`
	ChainId       uint64 `json:"chainId"`
	Hash          string `json:"hash"`
	Message       string `json:"message"`
	Contract      string `json:"contract"` //代币地址
	Number        uint64 `json:"number"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
	StartTime     uint64 `json:"startTime"`
	EndTime       uint64 `json:"endTime"`
	OperationTime uint64 `json:"operationTime"`
	Type          uint8  `json:"type"` //1-发红包 2-收红包"
}

type SendRedPacketRecord struct {
	RelationId   string                   `json:"relationId"`
	TokenAddress string                   `json:"tokenAddress"`
	TotalAmount  string                   `json:"totalAmount"`
	TotalNumber  uint64                   `json:"totalNumber"`
	Fee          string                   `json:"fee"`
	SendTime     uint64                   `json:"SendTime"`
	Sender       string                   `json:"sender"`
	Receivers    []ReceiveRedPacketRecord `json:"receivers"`
}

type ReceiveRedPacketRecord struct {
	RelationId     string `json:"relationId"`
	Receiver       string `json:"receiver"`
	TokenAddress   string `json:"tokenAddress"`
	ReceivedAmount string `json:"receivedAmount"`
	ReceiveTime    uint64 `json:"receiveTime"`
	Fee            string `json:"fee"`
}

type RefundRecord struct {
	RelationId     string `json:"relationId"`
	TokenAddress   string `json:"tokenAddress"`
	ReceivedAmount string `json:"receivedAmount"`
	ReceiveTime    uint64 `json:"receiveTime"`
}

type CreationSuccess struct {
	Id           *big.Int
	Total        *big.Int
	Name         string
	Message      string
	Creator      common.Address
	CreationTime *big.Int
	TokenAddress common.Address
	Number       *big.Int
	Ifrandom     bool
	Duration     *big.Int
}

type ClaimSuccess struct {
	Id           *big.Int
	Claimer      common.Address
	TokenAddress common.Address
	ClaimedValue *big.Int
}

type RefundSuccess struct {
	Id               *big.Int
	TokenAddress     common.Address
	RemainingBalance *big.Int
}

func NewRedPacket(redPacket *RedPacket) error {
	key := fmt.Sprintf("FinRedPacket-%s", redPacket.Id)
	keyIndex := fmt.Sprintf("%s-FinRedPacketIndex", redPacket.Hash)
	js, err := json.Marshal(redPacket)
	if err != nil {
		return err
	}

	batch := new(leveldb.Batch)
	batch.Put([]byte(key), js)
	batch.Put([]byte(keyIndex), []byte(redPacket.Id))

	return database.Leveldb.Write(batch, nil)
}

func NewSendRedPacketRecord(sendRecord *SendRedPacketRecord) error {
	key := fmt.Sprintf("SendRPRecod-%s-%s", sendRecord.Sender, sendRecord.RelationId)
	fmt.Println(key)
	js, err := json.Marshal(sendRecord)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func NewReceiveRedPacketRecord(receiveRecord *ReceiveRedPacketRecord) error {
	key := fmt.Sprintf("ReceiveRPRecod-%s-%s", receiveRecord.Receiver, receiveRecord.RelationId)
	fmt.Println(key)
	js, err := json.Marshal(receiveRecord)
	if err != nil {
		return err
	}
	return database.Leveldb.Put([]byte(key), js, nil)
}

func GetRedPacket(id string) (rp string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		return err.Error()
	}
	defer snapshot.Release()

	key := fmt.Sprintf("FinRedPacket-%s", id)
	rpByte, err := snapshot.Get([]byte(key), nil)
	if err != nil {
		return err.Error()
	}
	rp = string(rpByte)
	return
}

func GetRedPacketByHash(hash string) (rp string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		return err.Error()
	}
	defer snapshot.Release()

	key := fmt.Sprintf("%s-FinRedPacketIndex", hash)
	rpByte, err := snapshot.Get([]byte(key), nil)
	if err != nil {
		return err.Error()
	}

	key = fmt.Sprintf("FinRedPacket-%s", string(rpByte))
	rpByte2, err := snapshot.Get([]byte(key), nil)
	if err != nil {
		return err.Error()
	}
	rp = string(rpByte2)
	return
}

func GetSendRPRecord(id string, creator string) (sendrpr string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		return err.Error()
	}
	defer snapshot.Release()

	key := fmt.Sprintf("SendRPRecod-%s-%s", creator, id)
	rprByte, err := snapshot.Get([]byte(key), nil)
	if err != nil {
		return err.Error()
	}
	sendrpr = string(rprByte)
	return
}

func GetReceiveRPRecord(id string, receiver string) (receiverpr string) {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		return err.Error()
	}
	defer snapshot.Release()

	key := fmt.Sprintf("ReceiveRPRecod-%s-%s", receiver, id)
	rprByte, err := snapshot.Get([]byte(key), nil)
	if err != nil {
		return err.Error()
	}
	receiverpr = string(rprByte)
	return
}

func GetSendRPRecords(creator string) string {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		return err.Error()
	}
	defer snapshot.Release()

	preKey := fmt.Sprintf("SendRPRecod-%s-", creator)
	iter := snapshot.NewIterator(util.BytesPrefix([]byte(preKey)), nil)
	defer iter.Release()

	var sendRPRecords []SendRedPacketRecord
	for iter.Next() {
		var sendRPRs SendRedPacketRecord
		err := json.Unmarshal(iter.Value(), &sendRPRs)
		if err != nil {
			return err.Error()
		}
		sendRPRecords = append(sendRPRecords, sendRPRs)
	}

	re, _ := json.Marshal(sendRPRecords)
	return string(re)
}

func GetReceiveRPRecords(receiver string) string {
	snapshot, err := database.Leveldb.GetSnapshot() //创建快照进行读
	if err != nil {
		return err.Error()
	}
	defer snapshot.Release()

	preKey := fmt.Sprintf("ReceiveRPRecod-%s-", receiver)
	iter := snapshot.NewIterator(util.BytesPrefix([]byte(preKey)), nil)
	defer iter.Release()

	var receiveRPRecords []ReceiveRedPacketRecord
	for iter.Next() {
		var receiveRPRs ReceiveRedPacketRecord
		err := json.Unmarshal(iter.Value(), &receiveRPRs)
		if err != nil {
			return err.Error()
		}
		receiveRPRecords = append(receiveRPRecords, receiveRPRs)
	}

	re, _ := json.Marshal(receiveRPRecords)
	return string(re)
}

func ClaimRPDataUpdate(receiveRecord *ReceiveRedPacketRecord) error {
	lockKey := fmt.Sprintf("ClaimRP-%s", receiveRecord.RelationId)
	redPacketLock.Lock(lockKey)
	defer redPacketLock.Unlock(lockKey)

	batch := new(leveldb.Batch)

	// 1 捞red packet对象，修改参数，入库
	var packet RedPacket
	packetStr := GetRedPacket(receiveRecord.RelationId)
	println(receiveRecord.RelationId)
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		zlog.Zlog.Error(err.Error())
		return err
	}

	var subAmount = new(big.Int)
	var amountPer = new(big.Int)
	subAmount.SetString(receiveRecord.ReceivedAmount, 10)
	amountPer.SetString(packet.RemainingAmount, 10)

	packet.RemainingAmount = amountPer.Sub(amountPer, subAmount).String()
	packet.RemainingNumber = packet.RemainingNumber - 1

	packetKey := fmt.Sprintf("FinRedPacket-%s", receiveRecord.RelationId)
	newPacket, err := json.Marshal(packet)
	if err != nil {
		zlog.Zlog.Error(err.Error())
		return err
	}

	batch.Put([]byte(packetKey), newPacket)
	// 2 封装receive record对象，入库
	err = NewReceiveRedPacketRecord(receiveRecord)
	if err != nil {
		zlog.Zlog.Error(err.Error())
		return err
	}

	// 3 捞send record 对象，加入receive记录，入库
	var sendRecord SendRedPacketRecord
	sendRecordStr := GetSendRPRecord(packet.Id, packet.Creator)
	err = json.Unmarshal([]byte(sendRecordStr), &sendRecord)
	if err != nil {
		zlog.Zlog.Error(err.Error())
		return err
	}

	sendRecord.Receivers = append(sendRecord.Receivers, *receiveRecord)

	sendKey := fmt.Sprintf("SendRPRecod-%s-%s", sendRecord.Sender, sendRecord.RelationId)
	newSendRecord, err := json.Marshal(sendRecord)
	if err != nil {
		zlog.Zlog.Error(err.Error())
		return err
	}

	batch.Put([]byte(sendKey), newSendRecord)
	return database.Leveldb.Write(batch, nil)
}

func OwnerRefundUpdate(refundRecord *RefundRecord) error {
	lockKey := fmt.Sprintf("RefundRP-%s", refundRecord.RelationId)
	redPacketLock.Lock(lockKey)
	defer redPacketLock.Unlock(lockKey)

	batch := new(leveldb.Batch)

	// 1 捞red packet对象，修改参数，入库
	var packet RedPacket
	packetStr := GetRedPacket(refundRecord.RelationId)
	err := json.Unmarshal([]byte(packetStr), &packet)
	if err != nil {
		return err
	}

	var subAmount = new(big.Int)
	var amountPer = new(big.Int)
	subAmount.SetString(refundRecord.ReceivedAmount, 10)
	amountPer.SetString(packet.RemainingAmount, 10)

	packet.RemainingAmount = amountPer.Sub(amountPer, subAmount).String()
	packet.RemainingNumber = packet.RemainingNumber - 1

	packetKey := fmt.Sprintf("FinRedPacket-%s", refundRecord.RelationId)
	newPacket, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	batch.Put([]byte(packetKey), newPacket)
	// 2 封装refund record对象，入库
	refundKey := fmt.Sprintf("RefundRPRecod-%s", refundRecord.RelationId)

	refundStr, err := json.Marshal(refundRecord)
	if err != nil {
		return err
	}
	batch.Put([]byte(refundKey), refundStr)

	return database.Leveldb.Write(batch, nil)
}
