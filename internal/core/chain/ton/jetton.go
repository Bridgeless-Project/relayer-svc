package ton

import (
	"context"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) withdrawToken(ctx context.Context, depositData db.Deposit) (string, int64, error) {
	body, err := c.buildWithdrawJettonCell(ctx, depositData)
	if err != nil {
		return "", 0, errors.Wrap(err, "error building withdraw jetton cell")
	}

	b, err := c.Chain.Client.GetMasterchainInfo(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "error getting master chain info")
	}

	txHash, err := c.withdraw(ctx, body)

	return txHash, int64(b.SeqNo), errors.Wrapf(err, "failed to withdraw jetton")
}

func (c *Client) buildWithdrawJettonCell(ctx context.Context, depositData db.Deposit) (*cell.Cell, error) {
	hashInt := big.NewInt(0).SetBytes(txHashToBytes32(depositData.TxHash))
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

func (c *Client) getWithdrawalJettonHash(ctx context.Context, deposit db.Deposit) ([]byte, error) {
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
		return nil, errors.Wrap(err, "failed to get receiver cell")
	}

	var wrappedBit int64
	if deposit.IsWrappedToken {
		wrappedBit = trueBit
	}

	withdrawalTokenCell, err := getAddressCell(deposit.WithdrawalToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get withdrawal token address cell")
	}

	withdrawalAmount, ok := big.NewInt(0).SetString(deposit.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("failed to parse withdrawal amount")
	}

	txNonce := big.NewInt(0).SetUint64(uint64(deposit.TxNonce))

	txHash := big.NewInt(0).SetBytes(txHashToBytes32(deposit.TxHash))

	res, err := c.Chain.Client.WaitForBlock(master.SeqNo).RunGetMethod(
		ctx,
		master,
		c.Chain.BridgeContractAddress,
		withdrawalJettonHashMethod,
		withdrawalAmount,
		receiverCell.BeginParse(),
		txHash,
		txNonce,
		networkCell.BeginParse(),
		wrappedBit,
		withdrawalTokenCell.BeginParse(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the jetton hash")
	}

	resBig, err := res.Int(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the jetton hash")
	}

	return resBig.Bytes(), nil
}

func (c *Client) deriveJettonAddress(ctx context.Context, ownerAddress, jettonAddress *address.Address) (*address.Address, error) {
	block, err := c.Chain.Client.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error getting current masterchain info")
	}

	queryCell := cell.BeginCell()
	err = queryCell.StoreAddr(ownerAddress)
	if err != nil {
		return nil, errors.Wrap(err, "error storing owner address")
	}

	res, err := c.Chain.Client.WaitForBlock(block.SeqNo).RunGetMethod(ctx, block, jettonAddress, getJettonWalletMethod, queryCell.EndCell().BeginParse())
	if err != nil {
		return nil, errors.Wrap(err, "error getting jetton address")
	}

	resSlice, err := res.Slice(0)
	if err != nil {
		return nil, errors.Wrap(err, "error getting result slice")
	}
	val, err := resSlice.LoadAddr()
	if err != nil {
		return nil, errors.Wrap(err, "error loading jetton address")
	}

	return val, nil
}
