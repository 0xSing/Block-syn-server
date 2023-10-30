package worker

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"math/big"
	"time"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/utils/tron/api"
	zlog "walletSynV2/utils/zlog_sing"
)

func (w *Worker) GetEthBlock() {
	defer w.Wg.Done()
	currentHeight, err := models.ReadChainHeight("chainHeight")
	if err != nil {
		zlog.Zlog.Error("read chain height err", zap.Error(err))
	}
	if currentHeight < etc.Conf.Server.StartHeight {
		currentHeight = etc.Conf.Server.StartHeight
	}

	for {
		select {
		case <-w.Closed:
			fmt.Println("Eth scan block got close signal.....")
			return

		default:
			zlog.Zlog.Info("synchronizing block currentHeight: ", zap.Int64("currentHeight", currentHeight))
			time.Sleep(1 * time.Millisecond)

			blockNumber := big.NewInt(currentHeight)
			block, err := node.EthClient.GetClient().BlockByNumber(context.Background(), blockNumber)

			if err != nil {
				zlog.Zlog.Info("The maximum block has been reached...")
				//zlog.Zlog.Info("err :", zap.Error(err))
				time.Sleep(3000 * time.Millisecond)
				continue
			}

			// filter FinLock Event Logs
			w.Wg.Add(1)
			go ListenFinLockEvent(w, block)

			// filter RedPacket Event Logs
			w.Wg.Add(1)
			go ListenFinRedPacketEvent(w, block)

			// scan transactions
			w.Wg.Add(1)
			go EthTransactions(w, block)

			currentHeight++
			if err := models.UpdateChainHeight("chainHeight", currentHeight); err != nil {
				zlog.Zlog.Warn("update chain height err height", zap.Int64("currentHeight", currentHeight),
					zap.Error(err))
			}
		}
	}
}

func (w *Worker) GetTronBlock() {
	defer w.Wg.Done()
	client := node.TronClient.GetClient()
	currentHeight, err := models.ReadChainHeight("chainHeight")
	if err != nil {
		zlog.Zlog.Error("read chain height err", zap.Error(err))
	}
	if currentHeight < etc.Conf.Server.StartHeight {
		currentHeight = etc.Conf.Server.StartHeight
	}

	for {
		select {
		case <-w.Closed:
			fmt.Println("Tron scan block got close signal.....")
			return

		default:
			zlog.Zlog.Info("synchronizing block currentHeight: ", zap.Int64("currentHeight", currentHeight))
			block, err := client.GetBlockByNum2(context.Background(), &api.NumberMessage{Num: currentHeight})
			if err != nil {
				zlog.Zlog.Error("Get Tron block err: " + err.Error())
				return
			}

			if block.String() == "" {
				zlog.Zlog.Info("The maximum block has been reached...")
				time.Sleep(3000 * time.Millisecond)
				continue
			}

			// scan tron block transaction
			w.Wg.Add(1)
			go TronTransactions(w, block, currentHeight)

			currentHeight++
			if err := models.UpdateChainHeight("chainHeight", currentHeight); err != nil {
				zlog.Zlog.Warn("update chain height err", zap.Int64("currentHeight", currentHeight),
					zap.Error(err))
			}
		}
	}
}
