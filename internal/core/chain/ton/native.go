package ton

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) withdrawNative(ctx context.Context, depositData *db.Deposit) (string, int64, error) {
	ctxt := c.Chain.Client.Client().StickyContext(ctx)
	withdrawNativeCell, err := c.buildWithdrawNativeCell(depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to build withdraw native cell")
	}

	b, err := c.Chain.Client.GetMasterchainInfo(ctxt)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get master chain info")
	}

	hash, err := c.withdraw(ctxt, withdrawNativeCell)

	return hash, int64(b.SeqNo), errors.Wrapf(err, "failed to withdraw native")
}

func (c *Client) buildWithdrawNativeCell(depositData *db.Deposit) (*cell.Cell, error) {
	hashInt := big.NewInt(0).SetBytes(txHashToBytes32(depositData.TxHash))

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

func (c *Client) getWithdrawalNativeHash(ctx context.Context, deposit db.Deposit) ([]byte, error) {
	master, err := c.Chain.Client.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the master chain info")
	}

	networkCell, err := getNetworkCell(deposit.WithdrawalChainId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the network cell")
	}

	receiverCell, err := getAddressCell(deposit.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get receiver address cell")
	}

	withdrawalAmount, ok := big.NewInt(0).SetString(deposit.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("failed to parse withdrawal amount")
	}

	txHash := big.NewInt(0).SetBytes(txHashToBytes32(deposit.TxHash))
	txNonce := big.NewInt(0).SetUint64(uint64(deposit.TxNonce))

	res, err := c.Chain.Client.WaitForBlock(master.SeqNo).RunGetMethod(
		ctx,
		master,
		c.Chain.BridgeContractAddress,
		withdrawalNativeHashMethod,
		withdrawalAmount,
		receiverCell.BeginParse(),
		txHash,
		txNonce,
		networkCell.BeginParse(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the native hash")
	}

	resBig, err := res.Int(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the withdrawal native hash")
	}

	return resBig.Bytes(), nil
}
