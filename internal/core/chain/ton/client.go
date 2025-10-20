package ton

import (
	"context"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"

	"github.com/pkg/errors"
)

type Client struct {
	Chain          Chain
	OperatorWallet *wallet.Wallet
}

func (c *Client) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	address, err := c.getStoreAddress(ctx, depositData)
	if err != nil {
		return false, errors.Wrap(err, "error getting store address")
	}

	block, err := c.Chain.Client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return false, errors.Wrap(err, "error getting current master chain info")
	}

	accountInfo, err := c.Chain.Client.WaitForBlock(block.SeqNo).GetAccount(ctx, block, address)
	if err != nil {
		return false, errors.Wrap(err, "error getting account info")
	}

	return accountInfo.IsActive, nil
}

// NewBridgeClient creates a new bridge Client for the given chain.
func NewBridgeClient(chain Chain) *Client {
	liteClt := liteclient.NewConnectionPool()
	err := liteClt.AddConnectionsFromConfigUrl(context.Background(), chain.RPC.GlobalConfigUrl)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to global config"))
	}
	globalConfig, err := liteclient.GetConfigFromUrl(context.Background(), chain.RPC.GlobalConfigUrl)
	if err != nil {
		panic(errors.Wrap(err, "failed to get global config"))
	}

	api := ton.NewAPIClient(liteClt, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(globalConfig)
	api.WithTimeout(chain.RPC.Timeout * time.Second)
	operatorWallet, err := wallet.FromPrivateKey(api, chain.OperatorPrivateKey, wallet.V4R2)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to operator"))
	}

	chain.Client = api

	return &Client{
		Chain:          chain,
		OperatorWallet: operatorWallet,
	}
}

func (c *Client) ChainId() string {
	return c.Chain.Id
}

func (c *Client) Type() chain.Type {
	return chain.TypeTON
}

func (c *Client) TransactionHashValid(hash string) bool {
	return core.DefaultTransactionHashPattern.MatchString(hash)
}
