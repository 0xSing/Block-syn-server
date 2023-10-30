package model

type (
	GetSignRequest struct {
		Addr      string `json:"addr"`
		TokenAddr string `json:"token"`
		Id        int64  `json:"id"`
		Amount    string `json:"amount"`
		Timestamp int64  `json:"timestamp"`
	}

	GetTxsRequest struct {
		Wallet string `json:"wallet"`
		Type   uint   `json:"type"`
		Page   int    `json:"page"`
		Size   int    `json:"size"`
		Token  string `json:"token"`
	}

	GetWalletSwapRequest struct {
		Wallet string `json:"wallet"`
		Hash   string `json:"hash"`
		Pub    string `json:"pub"`
	}

	GetAddrRequest struct {
		Addr string `json:"addr"`
		Type int    `json:"type"`
	}

	GetApproveRequest struct {
		Wallet string `json:"wallet"`
	}

	GetNftsRequest struct {
		Wallet string `json:"wallet"`
	}

	AddContractRequest struct {
		Type     int    `json:"type"`
		Contract string `json:"contract"`
	}

	ScanApiResp struct {
		Status  string `json:"-"`
		Message string `json:"message"`
		Result  []struct {
			BlockNumber       string `json:"blockNumber"`
			TimeStamp         string `json:"timeStamp"`
			Hash              string `json:"hash"`
			Nonce             string `json:"nonce"`
			BlockHash         string `json:"blockHash"`
			TransactionIndex  string `json:"transactionIndex"`
			From              string `json:"from"`
			To                string `json:"to"`
			Value             string `json:"value"`
			Gas               string `json:"gas"`
			GasPrice          string `json:"gasPrice"`
			IsError           string `json:"isError"`
			TxreceiptStatus   string `json:"txreceipt_status"`
			Input             string `json:"input"`
			ContractAddress   string `json:"contractAddress"`
			CumulativeGasUsed string `json:"cumulativeGasUsed"`
			GasUsed           string `json:"gasUsed"`
			Confirmations     string `json:"confirmations"`
			MethodID          string `json:"methodId,omitempty"`
			FunctionName      string `json:"functionName,omitempty"`
		} `json:"result"`
	}
)
