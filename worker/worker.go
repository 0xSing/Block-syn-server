package worker

import (
	"github.com/robfig/cron"
	"strconv"
	"time"
	"walletSynV2/etc"
	"walletSynV2/utils"
	zlog "walletSynV2/utils/zlog_sing"
)

type Worker struct {
	Closed chan struct{}
	Wg     *utils.FinWaitGroup
}

func NewWorker() *Worker {
	return &Worker{
		Closed: make(chan struct{}),
		Wg:     utils.NewFinWaitGroup(),
	}
}

// 定时任务
func (w *Worker) timer() {
	loc, _ := time.LoadLocation("Local")
	c := cron.NewWithLocation(loc)
	err := c.AddFunc("0 0/10 * * * ?  ", w.AddContract) // 10分钟更新一次，监听20合约
	if err != nil {
		zlog.Zlog.Error("start add contract task failed")
	}
	c.Start()
}

func (w *Worker) Stop() {
	close(w.Closed)
	for w.Wg.GetCounter() != 0 {
		time.Sleep(time.Millisecond * 200)
		counter := w.Wg.GetCounter()
		str := strconv.Itoa(int(counter))
		println("has " + str + " goroutines is running...")
	}
	w.Wg.Wait()
}

func (w *Worker) Run() {
	// These two goroutines don't need to wait
	go w.timer()
	go w.UpdateWallet()

	w.Wg.Add(2)
	if etc.Conf.Server.ChainMode == 0 {
		go w.GetEthBlock()
		go w.EthScanNft()
	} else {
		go w.GetTronBlock()
		go w.TronScanNft()
	}
}
