package ton

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) updateSigner(ctx context.Context, epochData *db.Epoch, signer *wallet.Wallet) (string, int64, error) {
	ctxt := c.Chain.Client.Client().StickyContext(ctx)
	updateSignerCell, err := c.buildUpdateSignerCell(epochData)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to build update signer cell")
	}

	b, err := c.Chain.Client.GetMasterchainInfo(ctxt)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get master chain info")
	}

	bytes, err := signer.SendManyWaitTxHash(ctx, []*wallet.Message{
		{
			Mode: 1,
			InternalMessage: &tlb.InternalMessage{
				IHRDisabled: true,
				Bounce:      false,
				DstAddr:     c.Chain.BridgeContractAddress,
				Amount:      tlb.FromNanoTONU(1500000000),
				Body:        updateSignerCell,
			},
		},
	})
	if err != nil {
		return "", 0, errors.Wrap(err, "error sending withdrawal transaction")
	}

	return hex.EncodeToString(bytes), int64(b.SeqNo), errors.Wrapf(err, "failed to withdraw native")
}

func (c *Client) buildUpdateSignerCell(epochData *db.Epoch) (*cell.Cell, error) {
	signatureCell, err := getSignatureCell(epochData.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "error getting signature cell")
	}

	// TODO: Replace the key with actual values
	cell0 := cell.BeginCell().
		MustStoreUInt(1, 8).
		MustStoreBigUInt(big.NewInt(2), 256).
		MustStoreBigUInt(big.NewInt(3), 256)

	cell1 := cell.BeginCell().
		MustStoreUInt(epochData.StartTime, 64).
		MustStoreUInt(epochData.EndTime, 64).
		MustStoreBoolBit(epochData.SignatureMode).
		MustStoreRef(signatureCell).
		EndCell()

	body := cell.BeginCell().
		MustStoreBuilder(cell0.MustStoreRef(cell1)).
		EndCell()

	return body, nil
}

func (c *Client) getUpdateSignerHash(ctx context.Context, epoch *db.Epoch) ([]byte, error) {
	master, err := c.Chain.Client.GetMasterchainInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the master chain info")
	}

	// TODO: Replace the key with actual values
	h := big.NewInt(1)
	x := big.NewInt(2)
	y := big.NewInt(3)

	res, err := c.Chain.Client.WaitForBlock(master.SeqNo).RunGetMethod(
		ctx,
		master,
		c.Chain.BridgeContractAddress,
		updateSignerHashMethod,
		h,
		x,
		y,
		epoch.StartTime,
		epoch.EndTime,
		epoch.SignatureMode,
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
