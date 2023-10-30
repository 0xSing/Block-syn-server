package handler

import (
	"github.com/gin-gonic/gin"
	"walletSynV2/handler/model"
	"walletSynV2/models"
	"walletSynV2/pkg"
)

func SwitchNFTScan(c *gin.Context) {
	var ar model.SwitchMode
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	if ar.Mode == 1 {
		models.NftStart <- struct{}{}
	} else if ar.Mode == 0 {
		models.NftStop <- struct{}{}
	}

	resp := pkg.MakeResp(pkg.Success, nil)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	更新快照扫下来的nft拥有者的信息
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.UpdateNftRequest	true	"更新快照扫下来的nft拥有者的信息请求体"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/updateNft [post]
func UpdateNft(c *gin.Context) {
	var ar model.UpdateNftRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	if len(ar.Keys) != len(ar.Values) {
		resp := pkg.MakeResp(pkg.ParamsErrorI18n, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	if err := models.UpdateNfts(ar.Keys, ar.Values, ar.Type); err != nil {
		resp := pkg.MakeResp(pkg.NONEWST, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, "")
	c.JSON(resp.HttpCode, resp)
	return
}
