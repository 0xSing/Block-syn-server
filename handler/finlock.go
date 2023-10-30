package handler

import (
	"github.com/gin-gonic/gin"
	"walletSynV2/handler/model"
	"walletSynV2/models"
	"walletSynV2/pkg"
)

// @Summary	获取当前钱包拥有的锁列表
// @Tags		FinLock
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetOwnerLocks	true	"当前钱包信息"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/getLocks [post]
func GetLocks(c *gin.Context) {
	var ar model.GetOwnerLocks
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	locks := models.GetLocks(ar.Owner)
	resp := pkg.MakeResp(pkg.Success, locks)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	修改锁状态为pending
// @Tags		FinLock
// @Accept		json
// @Produce	json
// @Param		model	body		model.ChangeLockStates	true	"解锁信息"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/unLock [post]
func Unlocking(c *gin.Context) {
	var ar model.ChangeLockStates
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	_ = models.ChangeLockStates(ar.Owner, ar.Id, true)
	resp := pkg.MakeResp(pkg.Success, "")
	c.JSON(resp.HttpCode, resp)
	return
}
