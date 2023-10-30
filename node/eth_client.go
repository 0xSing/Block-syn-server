package node

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthNodeClient struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	chainId   uint64
	host      string
	isTestnet bool
	Result    interface{}
}

func NewEthClient(host string, isDev bool) (*EthNodeClient, error) {
	rpcDial, err := rpc.Dial(host)
	if err != nil {
		return nil, err
	}

	c := new(EthNodeClient)
	c.client = ethclient.NewClient(rpcDial)
	id, err := c.client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}
	c.chainId = id.Uint64()
	c.host = host
	c.isTestnet = isDev
	c.rpcClient = rpcDial
	return c, nil
}

func (c *EthNodeClient) GetClient() *ethclient.Client {
	return c.client
}

func (c *EthNodeClient) GetRpcClient() *rpc.Client {
	return c.rpcClient
}

func (c *EthNodeClient) GetChainId() uint64 {
	return c.chainId
}

func (c *EthNodeClient) GetHost() string {
	return c.host
}
