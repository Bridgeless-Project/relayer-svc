package ton

import (
	"context"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type Client struct {
	Chain  Chain
	childs []*ChildClient
}

func (c *Client) ChildClients() []chain.ChildClient {
	childs := make([]chain.ChildClient, len(c.childs))
	for i, child := range c.childs {
		childs[i] = child
	}

	return childs
}

func (c *Client) ConfigureChildClients() chain.Client {
	childs := make([]*ChildClient, c.Chain.Workers)
	for i := 0; i < c.Chain.Workers; i++ {
		childs[i] = NewChildClient(c)
	}

	for i, key := range c.Chain.OperatorsPrivateKeys {
		idx := i % c.Chain.Workers
		wallet, err := wallet.FromPrivateKey(c.Chain.Client, key, wallet.V4R2)
		if err != nil {
			panic(errors.Wrap(err, "failed to create wallet"))
		}

		childs[idx].AddSigner(wallet)
	}

	c.childs = childs
	return c
}

func (c *Client) IsProcessed(ctx context.Context, depositData db.Deposit) (bool, error) {
	ctxt := c.Chain.Client.Client().StickyContext(ctx)

	address, err := c.getStoreAddress(ctxt, depositData)
	if err != nil {
		return false, errors.Wrap(err, "error getting store address")
	}

	block, err := c.Chain.Client.CurrentMasterchainInfo(ctxt)
	if err != nil {
		return false, errors.Wrap(err, "error getting current master chain info")
	}

	accountInfo, err := c.Chain.Client.GetAccount(ctxt, block, address)
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

	chain.Client = api

	return &Client{
		Chain: chain,
	}
}

func (c *Client) ChainId() string {
	return c.Chain.Id
}

func (c *Client) Type() chain.Type {
	return chain.TypeTON
}

func (c *Client) WorkersCount() int {
	return c.Chain.Workers
}

func (c *Client) TransactionHashValid(hash string) bool {
	return core.DefaultTransactionHashPattern.MatchString(hash)
}

func (c *Client) Withdraw(ctx context.Context, depositData *db.Deposit, signer *wallet.Wallet) (string, int64, error) {
	ctxt := c.Chain.Client.Client().StickyContext(ctx)

	if depositData.WithdrawalToken == core.DefaultNativeTokenAddress {
		return c.withdrawNative(ctxt, depositData, signer)
	}

	return c.withdrawToken(ctxt, depositData, signer)
}
