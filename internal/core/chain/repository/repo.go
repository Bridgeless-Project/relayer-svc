package repository

import (
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
)

type clientsRepository struct {
	clients map[string]chain.Client
}

func NewClientsRepository(clients []chain.Client) chain.Repository {
	clientsMap := make(map[string]chain.Client, len(clients))

	for _, cl := range clients {
		// TODO: REMOVE
		if cl.Type() == chain.TypeTON {
			continue
		}
		clientsMap[cl.ChainId()] = cl
	}

	return &clientsRepository{clients: clientsMap}
}

func (p clientsRepository) Client(chainId string) (chain.Client, error) {
	cl, ok := p.clients[chainId]
	if !ok {
		return nil, chain.ErrChainNotSupported
	}

	return cl, nil
}

func (p clientsRepository) SupportsChain(chainId string) bool {
	_, ok := p.clients[chainId]
	return ok
}
