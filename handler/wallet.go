package handler

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"io"
	"math/big"
	"net/http"
	"strings"
	"walletSynV2/etc"
	"walletSynV2/handler/model"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/pkg"
	"walletSynV2/utils/tron/hexutil"
	zlog "walletSynV2/utils/zlog_sing"
)

const (
	Tx            = 1
	Nft           = 2
	Ftswap        = 3
	approveMethod = "0x095ea7b3"
)

// @Summary	获取交易信息
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetTxsRequest	true	"获取交易信息请求体"
// @Success	200		{object}	model.BasicResp		"success response"
// @Failure	400		{object}	model.BasicResp		"params error"
// @Router		/web/getTxs [post]
func GetTxs(c *gin.Context) {
	var ar model.GetTxsRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	var key string
	token := ar.Token
	wallet := ar.Wallet
	if etc.Conf.Server.ChainMode != 1 {
		wallet = common.HexToAddress(ar.Wallet).Hex()
		if token != "" {
			token = common.HexToAddress(ar.Token).Hex()
		}
	}
	switch ar.Type {
	case Tx:
		key = wallet + "-tx-"
	case Nft:
		key = wallet + "-nft-"
	case Ftswap:
		key = wallet + "-ftswap-"
	default:
		key = wallet + "-"
	}
	txs := models.GetSubset(key, token, ar.Page, ar.Size, ar.Type)
	resp := pkg.MakeResp(pkg.Success, txs)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取钱包兑换记录
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetWalletSwapRequest	true	"获取兑换信息请求体"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/getWalletSwap [post]
func GetWalletSwap(c *gin.Context) {
	var ar model.GetWalletSwapRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	var key string
	wallet := ar.Wallet
	if etc.Conf.Server.ChainMode != 1 {
		wallet = common.HexToAddress(ar.Wallet).Hex()
	}

	key = wallet + "-ftswap-"
	txs := models.GetTx(key, ar.Hash)

	if ar.Pub != "" {
		t := encryptWithPub(ar.Pub, txs)
		if t != "" {
			txs = t
		}
	}

	resp := pkg.MakeResp(pkg.Success, txs)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取单独的钱包交易记录
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetWalletSwapRequest	true	"获取单独的钱包交易记录请求体"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/getWalletTx [post]
func GetWalletTx(c *gin.Context) {
	var ar model.GetWalletSwapRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	var key string
	wallet := ar.Wallet
	if etc.Conf.Server.ChainMode != 1 {
		wallet = common.HexToAddress(ar.Wallet).Hex()
	}

	key = wallet + "-tx-"
	txs := models.GetTx(key, ar.Hash)
	if ar.Pub != "" {
		t := encryptWithPub(ar.Pub, txs)
		if t != "" {
			txs = t
		}
	}

	resp := pkg.MakeResp(pkg.Success, txs)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取钱包或合约地址
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetAddrRequest	true	"获取钱包或合约地址请求体"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getAddr [post]
func GetAddr(c *gin.Context) {
	var ar model.GetAddrRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	var key string
	addr := ar.Addr
	if etc.Conf.Server.ChainMode != 1 {
		addr = common.HexToAddress(ar.Addr).Hex()
	}

	switch ar.Type {
	case 1:
		key = "wallet-" + addr
	case 2:
		key = "contract-" + addr
	default:
		key = ar.Addr
	}
	txs := models.GetSub(key)

	resp := pkg.MakeResp(pkg.Success, txs)
	c.JSON(resp.HttpCode, resp)
	return
}

func encryptWithPub(pubStr string, origData string) string {
	pubKey, err := getPublicKey(pubStr)
	if err != nil {
		return ""
	}

	blockSize := 110 // 每个块大小
	encryptedData := []string{}
	for len(origData) > 0 {
		// 每个块大小不超过RSA密钥的长度
		if len(origData) < blockSize {
			blockSize = len(origData)
		}

		// 对每个块进行加密，并将其追加到加密数据中
		blockData := make([]byte, blockSize)
		copy(blockData, origData[:blockSize])
		encryptedBlockData, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, blockData)
		if err != nil {
			fmt.Println("encrypt error", err)
			return ""
		}
		encryptedData = append(encryptedData, base64.StdEncoding.EncodeToString(encryptedBlockData))
		origData = origData[blockSize:]
	}
	return strings.Join(encryptedData, ",")
}

func getPublicKey(pubStr string) (*rsa.PublicKey, error) {
	pub := []byte("\n-----BEGIN PUBLIC KEY-----\n" + pubStr + "\n-----END PUBLIC KEY-----\n")
	// decode Base64 public key string
	//publicKeyBytes, err := base64.StdEncoding.DecodeString(pubStr)
	//if err != nil {
	//	fmt.Println("DecodeString error:", err)
	//	return
	//}

	// get PEM blocks from raw public key bytes
	var publicKey *rsa.PublicKey
	block, _ := pem.Decode(pub)
	if block == nil {
		fmt.Println("invalid PEM format")
		return publicKey, errors.New("invalid PEM format")
	}

	// PKCS#8 public key format
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println("ParsePKIXPublicKey error:", err)
		return publicKey, err
	}
	var ok bool
	publicKey, ok = pubInterface.(*rsa.PublicKey)
	if !ok {
		fmt.Println("invalid RSA public key")
		return publicKey, err
	}

	return publicKey, nil
}

type Approve struct {
	ERC20    string `json:"erc20"`
	Contract string `json:"contract"`
	Amount   string `json:"amount"`
	Hash     string `json:"hash"`
	Time     string `json:"time"`
}

// @Summary	获取erc20授权信息
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetAddrRequest	true	"获取erc20授权信息请求体"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getApprove [post]
func GetWalletTxs(c *gin.Context) {
	var ar model.GetApproveRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	address := common.HexToAddress(ar.Wallet)
	url := etc.Conf.ServerUrl.GetTxsListUrl + address.Hex()

	response, err := http.Get(url)
	if err != nil {
		zlog.Zlog.Error("Error: failed to fetch transactions.")
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		zlog.Zlog.Error("Error: failed to parse response.")
		return
	}

	var txs model.ScanApiResp
	if err = json.Unmarshal(body, &txs); err != nil {
		zlog.Zlog.Info("get txlist unmarshal error: " + err.Error())
	}

	if etc.Conf.Server.ChainMode == 1 {
		//api key 1c9ac856-850b-4f30-8986-8e93ce1fa7b4 跳转区块浏览器，不做了
		return
	}

	contractABI := `[{"inputs":[],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"owner","type":"address"},{"indexed":true,"internalType":"address","name":"spender","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Approval","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"previousOwner","type":"address"},{"indexed":true,"internalType":"address","name":"newOwner","type":"address"}],"name":"OwnershipTransferred","type":"event"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"uint256","name":"id","type":"uint256"}],"name":"Snapshot","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"internalType":"address","name":"from","type":"address"},{"indexed":true,"internalType":"address","name":"to","type":"address"},{"indexed":false,"internalType":"uint256","name":"value","type":"uint256"}],"name":"Transfer","type":"event"},{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"}],"name":"allowance","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"approve","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"}],"name":"balanceOf","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"snapshotId","type":"uint256"}],"name":"balanceOfAt","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"burn","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"account","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"burnFrom","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"decimals","outputs":[{"internalType":"uint8","name":"","type":"uint8"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"subtractedValue","type":"uint256"}],"name":"decreaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"addedValue","type":"uint256"}],"name":"increaseAllowance","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"mint","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"name","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"owner","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"renounceOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"snapshot","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"symbol","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"uint256","name":"snapshotId","type":"uint256"}],"name":"totalSupplyAt","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transfer","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"from","type":"address"},{"internalType":"address","name":"to","type":"address"},{"internalType":"uint256","name":"amount","type":"uint256"}],"name":"transferFrom","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"nonpayable","type":"function"},{"inputs":[{"internalType":"address","name":"newOwner","type":"address"}],"name":"transferOwnership","outputs":[],"stateMutability":"nonpayable","type":"function"}]`
	abi, err := abi.JSON(strings.NewReader(contractABI)) // contractABI是合约的ABI字节数组
	if err != nil {
		zlog.Zlog.Info("abi gen error: " + err.Error())
	}
	method := "allowance"
	var data map[string]Approve
	data = make(map[string]Approve)

	for _, t := range txs.Result {
		if len(t.Input) < 10 {
			continue
		}
		if t.Input[:10] == approveMethod {
			target := common.HexToAddress(t.Input[10:74])
			input, err := abi.Pack(method, common.HexToAddress(t.From), target)
			if err != nil {
				zlog.Zlog.Info("pack data error: " + err.Error())
			}
			to := common.HexToAddress(t.To)
			callMsg := ethereum.CallMsg{
				From: common.HexToAddress(t.From),
				To:   &to,
				Data: input,
			}

			output, err := node.EthClient.GetClient().CallContract(context.Background(), callMsg, nil)
			if err != nil {
				zlog.Zlog.Info("call contract: " + err.Error())
			}

			result, err := abi.Unpack(method, output)
			if err != nil {
				zlog.Zlog.Info("unpack output error: " + err.Error())
			}
			//fmt.Println("result", result, target)
			data[target.Hex()] = Approve{
				ERC20:    to.Hex(),
				Contract: target.Hex(),
				Amount:   fmt.Sprintf("%s", result[0]),
				Hash:     t.Hash,
				Time:     t.TimeStamp,
			}
		}
	}
	var resl []Approve
	for _, v := range data {
		resl = append(resl, v)
	}
	resp := pkg.MakeResp(pkg.Success, resl)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	获取钱包拥有的所有nft
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.GetNftsRequest	true	"获取钱包拥有的所有nft信息请求体"
// @Success	200		{object}	model.BasicResp			"success response"
// @Failure	400		{object}	model.BasicResp			"params error"
// @Router		/web/getNfts [post]
func GetNfts(c *gin.Context) {
	var ar model.GetNftsRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	nfts := models.GetNfts(common.HexToAddress(ar.Wallet).Hex())
	resp := pkg.MakeResp(pkg.Success, nfts)
	c.JSON(resp.HttpCode, resp)
	return
}

// @Summary	添加需要扫描的nft
// @Tags		Wallet
// @Accept		json
// @Produce	json
// @Param		model	body		model.AddContractRequest	true	"添加需要扫描的nft请求体"
// @Success	200		{object}	model.BasicResp				"success response"
// @Failure	400		{object}	model.BasicResp				"params error"
// @Router		/web/addContract [post]
func AddContract(c *gin.Context) {
	var ar model.AddContractRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	var key string
	contract := ar.Contract
	if etc.Conf.Server.ChainMode != 1 {
		contract = common.HexToAddress(ar.Contract).Hex()
	}
	switch ar.Type {
	case Tx:
		key = "contract-" + contract
	case Nft:
		key = "nft-" + contract
	}
	if key == "" {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}
	if err := models.Put(key, "1"); err != nil {
		zlog.Zlog.Info("add contract db error: " + err.Error())
		return
	}
	zlog.Zlog.Info("add contract to db: " + ar.Contract + " : " + string(ar.Type))
	resp := pkg.MakeResp(pkg.Success, nil)
	c.JSON(resp.HttpCode, resp)
	return
}

func GetSign(c *gin.Context) {
	var ar model.GetSignRequest
	if err := c.ShouldBindJSON(&ar); err != nil {
		resp := pkg.MakeResp(pkg.ParamsError, nil)
		c.JSON(resp.HttpCode, resp)
		return
	}

	num := new(big.Int)
	num, _ = num.SetString(ar.Amount, 10)
	hash := crypto.Keccak256Hash(
		common.HexToAddress(ar.Addr).Bytes(),
		common.HexToAddress(ar.TokenAddr).Bytes(),
		common.LeftPadBytes(big.NewInt(ar.Id).Bytes(), 32),
		common.LeftPadBytes(num.Bytes(), 32),
		common.LeftPadBytes(big.NewInt(ar.Timestamp).Bytes(), 32),
	)
	sign := getSignature("5a9529051d3d402f3237150702db81b46fae1231254c8dc6c00395270afe83fe", hash)
	resp := pkg.MakeResp(pkg.Success, sign)
	c.JSON(resp.HttpCode, resp)
	return
}

func getSignature(pKey string, hash common.Hash) string {
	fmt.Println("sign hash:", hash)
	privateKey, err := crypto.HexToECDSA(pKey)
	if err != nil {
		fmt.Println("get private key error", err)
	}
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		fmt.Println("sign error", err)
	}
	//for solidity verify
	if signature[64] == 0 || signature[64] == 1 {
		signature[64] += 27
	}
	return hexutil.Encode(signature)
}
