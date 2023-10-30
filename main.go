package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"walletSynV2/database"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/router"
	. "walletSynV2/utils/zlog_sing"
	"walletSynV2/worker"
)

//	@title			FinToken合约端程序
//	@version		1.0
//	@description	FinToken合约端API文档

//	@contact.name	API Support
//	@contact.author
//	@contact.email

// @host		localhost:8080
// @BasePath	/api
func main() {
	err := etc.InitConfig("./conf/config.yaml")
	if err != nil {
		Zlog.Error("init conf failed", zap.Error(err))
		os.Exit(1)
	}

	// 初始化节点
	if etc.Conf.Server.ChainMode == 0 {
		err := node.InitEthClient()
		if err != nil {
			Zlog.Error("init eth client failed", zap.Error(err))
			os.Exit(1)
		}
		etc.Conf.Level.ChainId = node.EthClient.GetChainId()

	} else {
		err := node.InitTronClient()
		if err != nil {
			Zlog.Error("init tron client failed", zap.Error(err))
			os.Exit(1)
		}
		etc.Conf.Level.ChainId = node.TronClient.GetChainId()
	}

	// 初始化signal监听 channel
	stopSignal := make(chan os.Signal)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	// 初始化Mysql对象模型
	//models.RegisterMySQLModel()

	// 初始化Mysql
	//isDebug := strings.EqualFold(etc.Conf.Server.RunMode, "debug")
	//err = database.InitMysql(etc.Conf.Mysql, isDebug)
	//if err != nil {
	//	Zlog.Error("init MySQL failed", zap.Error(err))
	//	os.Exit(1)
	//}

	// 初始化LevelDB
	err = database.InitLevelDB(etc.Conf.Level.ChainId)
	if err != nil {
		Zlog.Error("init LevelDB failed", zap.Error(err))
		os.Exit(1)
	}

	// 初始化gin router
	go router.Init(etc.Conf.Server.HttpPort, etc.Conf.Server.RunMode)

	// 初始化全局channal
	models.NftStop = make(chan struct{})
	models.NftStart = make(chan struct{})

	// 开始扫块
	w := worker.NewWorker()
	w.Run()

	select {

	case sig := <-stopSignal:
		Zlog.Info(fmt.Sprintf("Got %s signal. Aborting...\n", sig))
		w.Stop()
	}
}
