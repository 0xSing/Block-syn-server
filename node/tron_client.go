package node

import (
	"google.golang.org/grpc"
	"walletSynV2/utils/tron/api"
)

type TronNodeClient struct {
	client    api.WalletClient
	rpcClient *grpc.ClientConn
	chainId   uint64
	host      string
	isTestnet bool
	Result    interface{}
}

func NewTronClient(host string, isDev bool) (*TronNodeClient, error) {
	rpcClient, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := api.NewWalletClient(rpcClient)

	c := new(TronNodeClient)
	c.client = client
	c.rpcClient = rpcClient
	c.host = host
	c.isTestnet = isDev
	if isDev {
		c.chainId = 2494104990
	} else {
		c.chainId = 728126428
	}
	return c, nil
}

func (c *TronNodeClient) GetClient() api.WalletClient {
	return c.client
}

func (c *TronNodeClient) GetRpcClient() *grpc.ClientConn {
	return c.rpcClient
}

func (c *TronNodeClient) GetChainId() uint64 {
	return c.chainId
}

func (c *TronNodeClient) GetHost() string {
	return c.host
}
