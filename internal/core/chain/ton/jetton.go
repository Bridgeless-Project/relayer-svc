package ton

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) WithdrawToken(ctx context.Context, depositData db.Deposit) (string, error) {
	body, err := c.buildWithdrawJettonCell(ctx, depositData)
	if err != nil {
		return "", errors.Wrap(err, "error building withdraw jetton cell")
	}

	txHash, err := c.withdraw(ctx, body)
	if err != nil {
		return "", errors.Wrap(err, "error withdrawing jetton cell")
	}

	return txHash, nil
}

func (c *Client) buildWithdrawJettonCell(ctx context.Context, depositData db.Deposit) (*cell.Cell, error) {
	hashBytes, err := hexutil.Decode(depositData.TxHash)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding txHash")
	}
	hashInt := big.NewInt(0).SetBytes(hashBytes)
	signCell, err := getSignatureCell(depositData.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "error getting signature")
	}

	networkCell, err := getNetworkCell(depositData.WithdrawalChainId)
	if err != nil {
		return nil, errors.Wrap(err, "error converting network")
	}

	jettonAddress, err := address.ParseAddr(depositData.WithdrawalToken)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing jetton address")
	}

	bridgeJettonAddress, err := c.deriveJettonAddress(ctx, c.Chain.BridgeContractAddress, jettonAddress)

	receiverAddress, err := address.ParseAddr(depositData.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing receiver address")
	}

	amount, ok := big.NewInt(0).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.Wrap(err, "error parsing withdrawal amount")
	}

	cell2 := cell.BeginCell().
		MustStoreBigInt(big.NewInt(forwardTonAmount), amountBitSize).
		MustStoreBigInt(big.NewInt(totalTonAmount), amountBitSize).EndCell()

	cell1 := cell.BeginCell().
		MustStoreBigInt(big.NewInt(depositData.TxNonce), nonceBitSize).
		MustStoreRef(networkCell).
		MustStoreBoolBit(depositData.IsWrappedToken).
		MustStoreRef(signCell).
		MustStoreAddr(jettonAddress).
		MustStoreAddr(bridgeJettonAddress).
		MustStoreRef(cell2).EndCell()

	cell0 := cell.BeginCell().
		MustStoreUInt(withdrawJettonCode, opCodeBitSize).
		MustStoreAddr(receiverAddress).
		MustStoreBigInt(amount, amountBitSize).
		MustStoreBigInt(hashInt, amountBitSize).
		MustStoreRef(cell1)

	finalBody := cell.BeginCell().MustStoreBuilder(cell0).EndCell()

	return finalBody, nil
}
