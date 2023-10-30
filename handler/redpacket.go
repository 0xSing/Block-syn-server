package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strings"
	"walletSynV2/handler/model"
	"walletSynV2/models"
	"walletSynV2/pkg"
	"walletSynV2/service"
)

// @Summary	根据红包创建交易hash查询红包信息
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetRPByHashReq	true	"生成红包交易hash"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getRedPacketByHash [post]
func GetRPByHash(c *gin.Context) {
	var ar model.GetRPByHashReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	packet := models.GetRedPacketByHash(ar.Hash)
	resp := pkg.MakeResp(pkg.Success, packet)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	根据红包Id查询红包信息
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetRPByIdReq	true	"红包Id"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/getRedPacketById [post]
func GetRPById(c *gin.Context) {
	var ar model.GetRPByIdReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	packet := models.GetLockById(ar.Id)
	resp := pkg.MakeResp(pkg.Success, packet)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	根据hash获取红包分享链接
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetRPShareUri	true	"红包交易hash"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/getShareUri [post]
func GetRPShareUri(c *gin.Context) {
	var ar model.GetRPShareUri
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	shareLink, err := service.SpliceShareLink(ar.Hash)
	if err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, "tx hash pending, please wait a mount")
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, shareLink)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	根据hash确认链上红包交易状态，获取红包Id
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.CheckRPStates	true	"红包交易hash"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/checkRPTxStates [post]
func CheckRPTxStates(c *gin.Context) {
	var ar model.CheckRPStates
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	var rp models.RedPacket
	rpStr := models.GetRedPacketByHash(ar.Hash)
	err := json.Unmarshal([]byte(rpStr), &rp)
	if err != nil {
		resp := pkg.MakeResp(pkg.RedPacketTxPending, "")
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, rp.Id)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取红包金额
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetRandomAmountReq	true	"红包Id"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/getRPAmount [post]
func GetRandomAmount(c *gin.Context) {
	var ar model.GetRandomAmountReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	amount, err := service.GetRPAmount(ar.Id)
	if err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	if strings.EqualFold(amount, "0") {
		resp := pkg.MakeResp(pkg.RedPacketOutOfSock, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, amount)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取指定id红包的所有随机金额
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetRandomAmountReq	true	"红包Id"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/getRPAmounts [post]
func GetRandomAmounts(c *gin.Context) {
	var ar model.GetRandomAmountReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	amount, err := service.GetRPAmounts(ar.Id)
	if err != nil {
		resp := pkg.MakeResp(pkg.RedPacketWasExpired, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, amount)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取红包Claim签名
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetClaimSignReq	true	"获取红包签名信息"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getClaimSign [post]
func GetClaimSign(c *gin.Context) {
	var ar model.GetClaimSignReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	claimSign, err := service.GenerateClaimSign(ar.Id, ar.Amount, ar.Receiver)
	if err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, claimSign)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取创建红包记录列表
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetCreateRPsReq	true	"查询记录钱包地址"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getCreateRPs [post]
func GetCreateRPs(c *gin.Context) {
	var ar model.GetCreateRPsReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	records := models.GetSendRPRecords(ar.Owner)
	resp := pkg.MakeResp(pkg.Success, records)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取领取红包记录列表
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetClaimRPsReq	true	"查询记录钱包地址"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getClaimRPs [post]
func GetClaimRPs(c *gin.Context) {
	var ar model.GetClaimRPsReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	records := models.GetReceiveRPRecords(ar.Owner)
	resp := pkg.MakeResp(pkg.Success, records)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	根据Id查询红包创建记录
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetCreateRPRByIdReq	true	"查询红包创建信息"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/getCreateRPRById [post]
func GetCreateRPRById(c *gin.Context) {
	var ar model.GetCreateRPRByIdReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	records := models.GetSendRPRecord(ar.Id, ar.Owner)
	resp := pkg.MakeResp(pkg.Success, records)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	根据Id查询红包领取记录
// @Tags		RedPacket
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetClaimRPRByIdReq	true	"查询红包领取信息"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/getClaimRPRById [post]
func GetClaimRPRById(c *gin.Context) {
	var ar model.GetClaimRPRByIdReq
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	records := models.GetReceiveRPRecord(ar.Id, ar.Owner)
	resp := pkg.MakeResp(pkg.Success, records)
	c.JSON(resp.HttpCode, resp)
	return
}
