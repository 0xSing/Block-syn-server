package worker

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"io"
	"net/http"
	"time"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	zlog "walletSynV2/utils/zlog_sing"
)

func (w *Worker) AddContract() {
	zlog.Zlog.Info("AddContract start")
	var err error
	url := etc.Conf.ServerUrl.JavaHttpUrl + "/api/v1/dapp/coingecko/public/coin/all"
	c := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := c.Get(url)
	if err != nil {
		zlog.Zlog.Info("AddContract fail, get request error: " + err.Error())
		return
	}

	body, _ := io.ReadAll(resp.Body)
	contracts := Resp{}
	json.Unmarshal(body, &contracts)

	var network int
	if etc.Conf.Server.ChainMode == 1 {
		network = int(node.TronClient.GetChainId()) //波场链id
	} else {
		network = int(node.EthClient.GetChainId())
	}

	var key string
	for _, c := range contracts.Data {
		if c.ChainID == network {
			if etc.Conf.Server.ChainMode == 1 {
				key = fmt.Sprintf("contract-%s", c.Contact)
			} else {
				key = fmt.Sprintf("contract-%s", common.HexToAddress(c.Contact))
			}
			if !models.Has(key) {
				//fmt.Println("add contract", key)
				models.Put(key, c.Symbol)
			}
		}
	}
	zlog.Zlog.Info("AddContract success")
	return
}

type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		ID      string `json:"id"`
		Symbol  string `json:"symbol"`
		ChainID int    `json:"chainId"`
		Contact string `json:"contact"`
	} `json:"data"`
	Failed  bool `json:"failed"`
	Success bool `json:"success"`
}
