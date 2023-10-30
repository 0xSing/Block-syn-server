package router

import (
	"github.com/gin-gonic/gin"
	"walletSynV2/handler"
	"walletSynV2/middleware"
)

// internal api
func initInternalRouter(r *gin.RouterGroup) {
	internalApiRouter := r.Group("/internal", middleware.Auth())
	{
		internalApiRouter.POST("/nftRunSwitch", handler.SwitchNFTScan)
		internalApiRouter.POST("/updateNft", handler.UpdateNft) //更新快照扫下来的nft拥有者的信息
	}
}
