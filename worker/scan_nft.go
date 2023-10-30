package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/smirkcat/hdwallet"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"math/big"
	"strings"
	"time"
	"walletSynV2/etc"
	"walletSynV2/models"
	"walletSynV2/node"
	"walletSynV2/utils/tron/api"
	"walletSynV2/utils/tron/core"
	"walletSynV2/utils/tron/hexutil"
	zlog "walletSynV2/utils/zlog_sing"
)

const (
	transfer          = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	transferid        = "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	tranferMethod     = "23b872dd"
	safeTranferMethod = "b88d4fde"
	TxNft             = 2
	NftPre            = "-ownNft-"
)

func (w *Worker) EthScanNft() {
	defer w.Wg.Done()
	client := node.EthClient
	start, err := models.ReadChainHeight("nftHeight")
	if err != nil {
		zlog.Zlog.Fatal("read chain height err" + err.Error())
	}
	if start < etc.Conf.Server.NftHeight {
		start = etc.Conf.Server.NftHeight
	}

	for {
		select {
		case <-w.Closed:
			fmt.Println("Eth nft scan block got close signal.....")
			return

		case <-models.NftStop:
			fmt.Println("Eth nft scan got stop signal.....")
			select {
			case <-models.NftStart:
				fmt.Println("Eth nft scan got start signal.....")
			}

		default:
			zlog.Zlog.Info("synchronizing nft currentHeight: ", zap.Int64("currentHeight", start))
			//get to be nft in snap, and start to scan it
			c, n := models.GetToBeNft()
			if c != "" && n < start {
				if err = models.Put("nft-"+c, "1"); err != nil {
					zlog.Zlog.Info("put nft error" + err.Error())
					return
				}
				if err = models.Del("tobe-nft-" + c); err != nil {
					zlog.Zlog.Info("del to be nft error" + err.Error())
					return
				}
			}
			block, err := client.GetClient().BlockByNumber(context.Background(), big.NewInt(start))
			if err != nil {
				zlog.Zlog.Info("The maximum nft block has been reached...")
				//zlog.Zlog.Info("err: " + err.Error())
				time.Sleep(3000 * time.Millisecond)
				continue
			}

			EthScanFailNft(block)
			// 创建过滤器，筛选出与该智能合约有关的交易记录
			//bhash := block.Hash()
			//fmt.Println("block hash", bhash.Hex(), start)
			query := ethereum.FilterQuery{
				//Addresses: []common.Address{common.HexToAddress(nft.Contract)},
				FromBlock: big.NewInt(start),
				ToBlock:   big.NewInt(start),
				//BlockHash: &bhash,
				Topics: [][]common.Hash{
					{
						common.HexToHash(transfer),
					},
				},
			}

			logs, err := client.GetClient().FilterLogs(context.Background(), query)
			if err != nil {
				zlog.Zlog.Info("get eth nodeInit error" + err.Error())
				return
			}

			// 遍历所有交易记录，并输出相关信息
			for _, vLog := range logs {
				if !models.HasNft(vLog.Address.Hex()) {
					continue
				}
				tx, isPending, err := client.GetClient().TransactionByHash(context.Background(), vLog.TxHash)
				if err != nil {
					fmt.Println("get tx error" + err.Error())
					continue
				}
				if isPending {
					continue
				}

				from := common.HexToAddress(vLog.Topics[1].Hex())
				to := common.HexToAddress(vLog.Topics[2].Hex())
				tokenId := vLog.Topics[3].Big()
				receipts, err := node.EthClient.GetClient().TransactionReceipt(context.Background(), tx.Hash())

				txN := &models.NftTx{
					From:     from.Hex(),
					To:       to.Hex(),
					Contract: tx.To().Hex(),
					TokenId:  tokenId.String(),
					Status:   1,
					Hash:     tx.Hash().Hex(),
					ChainId:  *big.NewInt(int64(client.GetChainId())),
					Gas:      *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
					//Time:        uint64(time.Now().Unix()),
					Time:        block.Time(),
					TxType:      TxNft,
					BlockNumber: *big.NewInt(int64(vLog.BlockNumber)),
					GasPrice:    *tx.GasPrice(),
				}
				if models.HasWallet(from.Hex()) {
					if err := models.NewNftTx(from.Hex(), txN); err != nil {
						zlog.Zlog.Info("new nft tx error" + tx.Hash().String() + err.Error())
					}
				}
				if models.HasWallet(to.Hex()) {
					if err := models.NewNftTx(to.Hex(), txN); err != nil {
						zlog.Zlog.Info("new nft tx error" + tx.Hash().String() + err.Error())
					}
					jsB, _ := json.Marshal(txN)
					w.Wg.Add(1)
					go WorkerPushTo(w, jsB)
				}

				UpdateBalance(from.Hex(), to.Hex(), tokenId, tx.To().Hex())

			}
			start++
			if err := models.UpdateChainHeight("nftHeight", start); err != nil {
				zlog.Zlog.Warn("update chain height err height", zap.Int("start", int(start)),
					zap.Error(err))
			}
		}
	}

}

func (w *Worker) TronScanNft() {
	defer w.Wg.Done()
	// 创建Tron客户端
	client := node.TronClient

	start, err := models.ReadChainHeight("nftHeight")
	if err != nil {
		zlog.Zlog.Fatal("read nft height err" + err.Error())
	}
	if start < etc.Conf.Server.NftHeight {
		start = etc.Conf.Server.NftHeight
	}

	for {
		select {
		case <-w.Closed:
			fmt.Println("Tron nft scan block got close signal.....")
			return

		case <-models.NftStop:
			fmt.Println("Eth nft scan got stop signal.....")
			select {
			case <-models.NftStart:
				fmt.Println("Eth nft scan got start signal.....")
			}

		default:
			zlog.Zlog.Info("synchronizing nft currentHeight: ", zap.Int64("currentHeight", start))
			//get to be nft in snap, and start to scan it
			c, n := models.GetToBeNft()
			if c != "" && n < start {
				if err = models.Put("nft-"+c, "1"); err != nil {
					zlog.Zlog.Info("put nft error" + err.Error())
					return
				}
				if err = models.Del("tobe-nft-" + c); err != nil {
					zlog.Zlog.Info("del to be nft error" + err.Error())
					return
				}
			}

			block, err := client.GetClient().GetBlockByNum2(context.Background(), &api.NumberMessage{Num: start})
			if err != nil {
				zlog.Zlog.Error("Get Tron block err: " + err.Error())
				return
			}

			if block.String() == "" {
				zlog.Zlog.Info("The maximum nft block has been reached...")
				time.Sleep(3000 * time.Millisecond)
				continue
			}

			for _, tx := range block.Transactions {
				transactionId := hexutil.Encode(tx.Txid)
				txid := new(api.BytesMessage)
				txid.Value, err = hexutil.Decode(transactionId)

				rets := tx.Transaction.Ret
				//应该只有一个
				for _, t := range tx.Transaction.RawData.Contract {
					txStatus := uint64(1)
					if len(rets) < 1 || rets[0].ContractRet != core.Transaction_Result_SUCCESS {
						//失败交易
						txStatus = 0
						continue
					}

					if t.Type == core.Transaction_Contract_TriggerSmartContract { //调用智能合约
						unObj := &core.TriggerSmartContract{}
						err := proto.Unmarshal(t.Parameter.GetValue(), unObj)
						if err != nil {
							zlog.Zlog.Error("parse Contract %v err: " + err.Error())
							continue
						}
						contract := hdwallet.EncodeCheck(unObj.GetContractAddress())
						if strings.EqualFold(contract, "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t") {
							//筛选掉大量的USDT的交易
							continue
						}
						transinfo, _ := client.GetClient().GetTransactionInfoById(context.Background(), txid)
						data := unObj.GetData()
						if txStatus == 0 && len(data) > 72 && hexutil.Encode(data[:4]) == "23b872dd" {
							data[15] = 0x41
							from := hdwallet.EncodeCheck(data[15:36])
							data[51] = 0x41
							to := hdwallet.EncodeCheck(data[51:68])
							tokenId := new(big.Int).SetBytes(common.TrimLeftZeroes(data[72:100]))
							//失败nft转账交易
							if models.HasWallet(from) {
								txn := &models.NftTx{
									From:        from,
									To:          to,
									Contract:    contract,
									Status:      txStatus,
									TokenId:     tokenId.String(),
									Hash:        transactionId,
									ChainId:     *big.NewInt(int64(client.GetChainId())),
									Gas:         *big.NewInt(transinfo.GetFee()),
									Time:        uint64(transinfo.BlockTimeStamp) / 1000,
									TxType:      TxNft,
									BlockNumber: *big.NewInt(start),
									GasPrice:    *big.NewInt(0),
								}
								if err := models.NewNftTx(from, txn); err != nil {
									zlog.Zlog.Info("new tx error: "+transactionId, zap.Error(err))
								}
							}
							continue
						}

						for _, evenlog := range transinfo.Log {
							la := common.BytesToHash(evenlog.Address)
							la[11] = 0x41
							addr := hdwallet.EncodeCheck(la[11:])
							if !models.HasNft(addr) {
								continue
							}
							if hexutil.Encode(evenlog.Topics[0]) == transferid &&
								len(evenlog.Topics) == 4 {
								//nft transfer
								if len(evenlog.Topics[1]) != 32 || len(evenlog.Topics[2]) != 32 {
									continue
								}
								evenlog.Topics[1][11] = 0x41
								evenlog.Topics[2][11] = 0x41
								from := hdwallet.EncodeCheck(evenlog.Topics[1][11:])
								to := hdwallet.EncodeCheck(evenlog.Topics[2][11:])
								tokenId := new(big.Int).SetBytes(common.TrimLeftZeroes(evenlog.Topics[3]))
								txn := &models.NftTx{
									From:        from,
									To:          to,
									Contract:    contract,
									TokenId:     tokenId.String(),
									Status:      txStatus,
									Hash:        transactionId,
									ChainId:     *big.NewInt(int64(client.GetChainId())),
									Gas:         *big.NewInt(transinfo.GetFee()),
									Time:        uint64(transinfo.BlockTimeStamp) / 1000,
									TxType:      TxNft,
									BlockNumber: *big.NewInt(start),
									GasPrice:    *big.NewInt(0),
								}
								if models.HasWallet(from) {
									if err := models.NewNftTx(from, txn); err != nil {
										zlog.Zlog.Info("new tx error" + transactionId + err.Error())
									}
								}
								if models.HasWallet(to) {
									if err := models.NewNftTx(to, txn); err != nil {
										zlog.Zlog.Info("new tx error" + transactionId + err.Error())
									}
									//推送服务
									jsB, _ := json.Marshal(txn)
									w.Wg.Add(1)
									go WorkerPushTo(w, jsB)
								}
								UpdateBalance(from, to, tokenId, contract)
							}
						}
					}
				}
			}
			start++
			if err := models.UpdateChainHeight("nftHeight", start); err != nil {
				zlog.Zlog.Warn("update chain height err height", zap.Any("start", start),
					zap.Error(err))
			}
		}
	}
}

func EthScanFailNft(block *types.Block) {
	for _, tx := range block.Transactions() {
		if tx.To() == nil { //不是合约调用
			continue
		}
		if !models.HasNft(tx.To().Hex()) {
			continue
		}
		client := node.EthClient
		receipts, err := client.GetClient().TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			zlog.Zlog.Info("get nft receipts error", zap.Any("err:", err))
			continue
		}
		if receipts.Status != 0 {
			continue
		}
		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			zlog.Zlog.Info("get from address fail"+tx.Hash().String(), zap.Error(err))
			continue
		}
		if len(tx.Data()) > 4 &&
			(common.Bytes2Hex(tx.Data()[:4]) == safeTranferMethod || common.Bytes2Hex(tx.Data()[:4]) == tranferMethod) {

			to := common.BytesToAddress(tx.Data()[36:68]).Hex()
			tokenId := *new(big.Int).SetBytes(common.TrimLeftZeroes(tx.Data()[68:100]))
			txN := &models.NftTx{
				From:        from.Hex(),
				To:          to,
				Contract:    tx.To().Hex(),
				TokenId:     tokenId.String(),
				Status:      0,
				Hash:        tx.Hash().Hex(),
				ChainId:     *big.NewInt(int64(client.GetChainId())),
				Gas:         *tx.GasTipCap().Mul(tx.GasTipCap(), big.NewInt(int64(receipts.GasUsed))),
				Time:        block.Time(),
				TxType:      TxNft,
				BlockNumber: *block.Number(),
				GasPrice:    *tx.GasPrice(),
			}
			if models.HasWallet(from.Hex()) {
				if err := models.NewNftTx(from.Hex(), txN); err != nil {
					zlog.Zlog.Info("new nft tx error" + tx.Hash().String() + err.Error())
				}
			}
		}

	}
}

func UpdateBalance(from string, to string, tokenId *big.Int, contract string) {
	fO, _ := models.Get(from + NftPre + contract)
	fOwnA := strings.Split(fO, ",")
	fOwnA = DelSlice(fOwnA, tokenId.String())
	if (len(fOwnA) == 1 && fOwnA[0] == "") || len(fOwnA) == 0 {
		if err := models.Del(from + NftPre + contract); err != nil {
			fmt.Println("del f own key error" + err.Error())
		}
	} else {
		if err := models.Put(from+NftPre+contract, strings.Join(fOwnA, ",")); err != nil {
			zlog.Zlog.Info("put own nft fail, fOwn:", zap.Any("fOwnA", fOwnA))
		}
	}

	tO, _ := models.Get(to + NftPre + contract)
	tOwnA := strings.Split(tO, ",")
	if len(tOwnA) == 1 && tOwnA[0] == "" {
		tOwnA = []string{tokenId.String()}
	} else {
		tOwnA = append(tOwnA, tokenId.String())
	}
	if err := models.Put(to+NftPre+contract, strings.Join(tOwnA, ",")); err != nil {
		zlog.Zlog.Info("put own nft fail, tOwn:", zap.Any("tOwnA", tOwnA))
	}
}

func DelSlice(slice []string, val string) []string {
	j := 0
	if len(slice) > 0 {
		for _, item := range slice {
			if item != val {
				slice[j] = item
				j++
			}
		}
	}
	return slice[:j]
}
