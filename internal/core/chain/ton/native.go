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

func (c *Client) WithdrawNative(ctx context.Context, depositData db.Deposit) (txHash string, err error) {
	withdrawNativeCell, err := c.buildWithdrawNativeCell(depositData)
	if err != nil {
		return "", errors.Wrap(err, "failed to build withdraw native cell")
	}

	hash, err := c.withdraw(ctx, withdrawNativeCell)
	if err != nil {
		return "", errors.Wrap(err, "failed to process withdrawal")
	}

	return hash, nil
}

func (c *Client) buildWithdrawNativeCell(depositData db.Deposit) (*cell.Cell, error) {
	hashBytes, err := hexutil.Decode(depositData.TxHash)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding txHash")
	}
	hashInt := big.NewInt(0).SetBytes(hashBytes)

	networkCell, err := getNetworkCell(depositData.WithdrawalChainId)
	if err != nil {
		return nil, errors.Wrap(err, "error getting network cell")
	}

	signatureCell, err := getSignatureCell(depositData.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "error getting signature cell")
	}

	receiverAddress, err := address.ParseAddr(depositData.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing receiver address")
	}

	amount, ok := big.NewInt(0).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("error parsing withdrawal amount")
	}

	cell0 := cell.BeginCell().
		MustStoreUInt(withdrawNativeCode, opCodeBitSize).
		MustStoreBigInt(amount, amountBitSize).
		MustStoreAddr(receiverAddress).
		MustStoreBigInt(hashInt, hashBitSize)

	cell1 := cell.BeginCell().
		MustStoreBigInt(big.NewInt(depositData.TxNonce), nonceBitSize).
		MustStoreRef(signatureCell).
		MustStoreRef(networkCell)

	body := cell.BeginCell().MustStoreBuilder(cell0.MustStoreRef(cell1.EndCell())).EndCell()

	return body, nil
}
