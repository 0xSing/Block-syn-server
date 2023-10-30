package router

import (
	"github.com/gin-gonic/gin"
	"walletSynV2/handler"
)

// To Java server接口
func initAppRouter(r *gin.RouterGroup) {
	webApiRouter := r.Group("/web")
	{
		webApiRouter.POST("/getSign", handler.GetSign)             //获取签名
		webApiRouter.POST("/getTxs", handler.GetTxs)               //获取交易列表
		webApiRouter.POST("/getWalletSwap", handler.GetWalletSwap) //获取钱包兑换记录
		webApiRouter.POST("/getWalletTx", handler.GetWalletTx)     //获取单独的钱包交易记录
		webApiRouter.POST("/getAddr", handler.GetAddr)             //获取钱包或合约地址
		webApiRouter.POST("/getApprove", handler.GetWalletTxs)     //获取erc20授权信息
		webApiRouter.POST("/getNfts", handler.GetNfts)             //获取钱包拥有的所有nft
		webApiRouter.POST("/addContract", handler.AddContract)     //添加需要扫描的nft

		//FinLock api
		webApiRouter.POST("/getLocks", handler.GetLocks) //获取钱包拥有的锁列表
		webApiRouter.POST("/unLock", handler.Unlocking)  //修改锁状态为pending

		//FinRedPacket api
		webApiRouter.POST("/getRedPacketByHash", handler.GetRPByHash)  //根据红包创建交易hash查询红包信息
		webApiRouter.POST("/getRedPacketById", handler.GetRPById)      //根据红包Id查询红包信息
		webApiRouter.POST("/getShareUri", handler.GetRPShareUri)       //根据hash获取红包分享链接
		webApiRouter.POST("/checkRPTxStates", handler.CheckRPTxStates) //根据hash获取红包Id
		//webApiRouter.POST("/getRPAmount", handler.GetRandomAmount)       //获取红包金额
		webApiRouter.POST("/getRPAmounts", handler.GetRandomAmounts)     //获取指定id红包的所有随机金额
		webApiRouter.POST("/getClaimSign", handler.GetClaimSign)         //获取红包Claim签名
		webApiRouter.POST("/getCreateRPs", handler.GetCreateRPs)         //获取创建红包记录列表
		webApiRouter.POST("/getClaimRPs", handler.GetClaimRPs)           //获取领取红包记录列表
		webApiRouter.POST("/getCreateRPRById", handler.GetCreateRPRById) //根据Id查询红包创建记录
		webApiRouter.POST("/getClaimRPRById", handler.GetClaimRPRById)   //根据Id查询红包领取记录

		//MultiSignWallet api
		webApiRouter.POST("/createWallet", handler.CreateWallet)           //添加多签钱包(同一时间内只能添加一个多签钱包，重复提交则会覆盖)
		webApiRouter.POST("/addMultiOwner", handler.AddOwner)              //给多签钱包添加所有者（拥有投票权）
		webApiRouter.POST("/getWalletInitData", handler.GetInitWalletData) //给多签钱包添加所有者（拥有投票权）

	}
}
