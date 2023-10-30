package handler

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
	"walletSynV2/etc"
	"walletSynV2/handler/model"
	"walletSynV2/models"
	"walletSynV2/pkg"
	"walletSynV2/utils"
	zlog "walletSynV2/utils/zlog_sing"
)

var (
	lock utils.DBKeyLock
)

// @Summary	创建多签钱包
// @Tags		multisign
// @Accept		json
// @Produce	json
// @Param		model	body		model.CreateWalletRequest	true	"创建多签钱包参数请求体"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/createWallet [post]
func CreateWallet(c *gin.Context) {
	var ar model.CreateWalletRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	if ar.Creator == "" || ar.Creator == "0x0000000000000000000000000000000000000000" ||
		ar.OwnersCount < ar.Threshold || ar.OwnersCount < 1 || ar.Threshold < 1 ||
		ar.Name == "" {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	creator := ar.Creator
	if etc.Conf.Server.ChainMode != 1 {
		creator = common.HexToAddress(ar.Creator).Hex()
	}

	err := models.CreateMultiSig(&models.MultiSignWallet{
		Addr:        "",
		Creator:     creator,
		Wallets:     []string{creator},
		Times:       []string{strconv.Itoa(int(time.Now().Unix()))},
		OwnersCount: ar.OwnersCount,
		Threshold:   ar.Threshold,
		Name:        ar.Name,
	})
	if err != nil {
		zlog.Zlog.Info("leveldb create multi sign wallet error:" + err.Error())
		resp := pkg.MakeResp(pkg.InternalError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, "success")
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	添加多签钱包拥有者
// @Tags		multisign
// @Accept		json
// @Produce	json
// @Param		model	body		model.AddOwnerRequest	true	"添加多签钱包拥有者参数请求体"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/addMultiOwner [post]
func AddOwner(c *gin.Context) {
	var ar model.AddOwnerRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	creator := ar.Creator
	wallet := ar.Wallet
	if etc.Conf.Server.ChainMode != 1 {
		creator = common.HexToAddress(ar.Creator).Hex()
		wallet = common.HexToAddress(ar.Wallet).Hex()
	}

	key := "addOwner-" + creator
	lock.Lock(key)
	defer lock.Unlock(key)

	mw := models.GetMultiOwner(creator)
	if mw.Creator == "" || !isValid(mw.Wallets, wallet) {
		resp := pkg.MakeResp(pkg.ParamsError, "multi sign wallet not exist")
		c.JSON(resp.HttpCode, resp)
		return
	}

	mw.Wallets = append(mw.Wallets, wallet)
	mw.Times = append(mw.Times, strconv.Itoa(int(time.Now().Unix())))
	err := models.CreateMultiSig(mw) //都一样是添加
	if err != nil {
		zlog.Zlog.Info("leveldb create multi sign wallet error:" + err.Error())
		resp := pkg.MakeResp(pkg.InternalError, "add owner error")
		c.JSON(resp.HttpCode, resp)
		return
	}

	resp := pkg.MakeResp(pkg.Success, "success")
	c.JSON(resp.HttpCode, resp)
	return
}

func GetInitWalletData(c *gin.Context) {
	var ar model.InitWalletRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	mw := models.GetMultiOwner(ar.Creator)
	if len(mw.Wallets) != int(mw.OwnersCount) ||
		mw.Creator == "" {
		resp := pkg.MakeResp(pkg.ParamsError, "wallet not exist or not complete")
		c.JSON(resp.HttpCode, resp)
		return
	}
	//initCode := contract.GetSetup(mw.Wallets, mw.Threshold)

	resp := pkg.MakeResp(pkg.Success, "")
	c.JSON(resp.HttpCode, resp)
	return
}

func isValid(wallets []string, wallet string) (isValid bool) {
	isValid = true
	if wallet == "0x0000000000000000000000000000000000000000" {
		isValid = false
		return
	}
	for _, w := range wallets {
		if w == wallet {
			isValid = false
			return
		}
	}
	return
}
