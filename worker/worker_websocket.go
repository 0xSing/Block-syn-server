package worker

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"time"
	"walletSynV2/etc"
	"walletSynV2/models"
	zlog "walletSynV2/utils/zlog_sing"
)

type WalletResp struct {
	Type      string `json:"type"`
	SocketDto struct {
		Type string `json:"type"`
		Data string `json:"data"`
	} `json:"socketDto"`
	Subject string   `json:"subject"`
	List    []string `json:"list"`
}

func (w *Worker) UpdateWallet() {
	fmt.Println("start update wallet")
	time.Sleep(time.Second * 10)
	wsUrl := etc.Conf.ServerUrl.JavaWebSocketUrl
	// 创建连接
	connect, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if nil != err {
		zlog.Zlog.Error("connect ws error", zap.Any("err info", err))
		//重连
		go w.UpdateWallet()
		return
	}
	//defer connect.Close()

	//定时向客户端发送数据
	go tickWriter(connect)

	//启动数据读取循环，读取客户端发送来的数据
	for {
		messageType, messageData, err := connect.ReadMessage()
		if nil != err {
			zlog.Zlog.Error("read ws msg error", zap.Any("err info", err))
			connect.Close()
			//重新连接
			go w.UpdateWallet()
			return
		}
		switch messageType {
		case websocket.TextMessage: //文本数据
			getWallet(string(messageData))
		case websocket.BinaryMessage: //二进制数据
			fmt.Println(messageData)
		case websocket.CloseMessage: //关闭
			go w.UpdateWallet()
			return
		case websocket.PingMessage: //Ping
		case websocket.PongMessage: //Pong
		default:
		}
	}
}

func getWallet(jsonS string) {
	if jsonS == "ping" {
		return
	}
	var js WalletResp
	json.Unmarshal([]byte(jsonS), &js)
	for _, l := range js.List {
		if etc.Conf.Server.ChainMode == 1 && len(l) == 34 { //trx链钱包地址
			walletKey := "wallet-" + l
			if !models.Has(walletKey) {
				fmt.Println("write wallet", l, len(l))
				if err := models.Put(walletKey, "1"); err != nil {
					log.Error("wallet insert error", zap.Any(l, err))
				}
			}
		} else if len(l) == 42 && etc.Conf.Server.ChainMode == 0 {
			walletKey := fmt.Sprintf("wallet-%s", common.HexToAddress(l))
			if !models.Has(walletKey) {
				fmt.Println("write wallet", l)
				if err := models.Put(walletKey, "1"); err != nil {
					log.Error("wallet insert error", zap.Any(l, err))
				}
			}
		}
	}
}

// 心跳数据
func tickWriter(connect *websocket.Conn) {
	json := "{\"type\":\"WalletAddrSubscribe\",\"data\":\"{\\\"subject\\\":\\\"WalletAddr\\\",\\\"key\\\":\\\"WalletAddr\\\"}\"}"
	err := connect.WriteMessage(websocket.TextMessage, []byte(json))
	if nil != err {
		log.Error("write ws msg error:WalletAddrSubscribe", zap.Any("WalletAddrSubscribe", err))
		return
	}
	for {
		//心跳
		err := connect.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"pong\", \"data\":\"\"}"))
		if nil != err {
			log.Error("write ws msg error:pong", zap.Any("pong", err))
			break
		}
		time.Sleep(time.Second * 30)
	}
}
