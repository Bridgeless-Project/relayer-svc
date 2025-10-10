package ton

import (
	"context"
	"time"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	chain "github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster/chains"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"

	"github.com/pkg/errors"
)

type Client struct {
	Chain
}

// NewBridgeClient creates a new bridge Client for the given chain.
func NewBridgeClient(chain Chain) *Client {
	liteClt := liteclient.NewConnectionPool()
	err := liteClt.AddConnectionsFromConfigUrl(context.Background(), chain.RPC.GlobalConfigUrl)
	if err != nil {
		panic(errors.Wrap(err, "failed to connect to global config"))
	}
	globalConfig, err := liteclient.GetConfigFromUrl(context.Background(), chain.RPC.GlobalConfigUrl)

	api := ton.NewAPIClient(liteClt, ton.ProofCheckPolicyFast).WithRetry()
	api.SetTrustedBlockFromConfig(globalConfig)
	api.WithTimeout(chain.RPC.Timeout * time.Second)

	chain.Client = api

	return &Client{
		chain,
	}
}

func (c *Client) ChainId() string {
	return c.Id
}

func (c *Client) Type() chain.Type {
	return chain.TypeTON
}

func (c *Client) AddressValid(addr string) bool {
	_, err := address.ParseAddr(addr)
	return err == nil
}

func (c *Client) TransactionHashValid(hash string) bool {
	return broadcaster.DefaultTransactionHashPattern.MatchString(hash)
}
