package ton

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	updateSignerKeyBySignatureOpCode = 0x12312324
	signerRecoveryH = 0x04
	updateSignersGas = 200000000
)

func (c *Client) UpdateSigners(ctx context.Context, epochData *db.Epoch, signer *wallet.Wallet) (string, int64, error) {
  ctxt := c.Chain.Client.Client().StickyContext(ctx)

  updateSignerCell, err := c.buildUpdateSignerCell(epochData)
  if err != nil {
    return "", 0, errors.Wrap(err, "failed to build update signer cell")
  }

  block, err := c.Chain.Client.CurrentMasterchainInfo(ctxt)
  if err != nil {
    return "", 0, errors.Wrap(err, "failed to get current master chain info")
  }

  txHashBytes, err := signer.SendManyWaitTxHash(ctxt, []*wallet.Message{
    {
      Mode: 1,
      InternalMessage: &tlb.InternalMessage{
        IHRDisabled: true,
        Bounce:      true,
        DstAddr:     c.Chain.BridgeContractAddress,
        Amount:      tlb.FromNanoTONU(updateSignersGas),
        Body:        updateSignerCell,
      },
    },
  })
  if err != nil {
    return "", 0, errors.Wrap(err, "failed to send update signers transaction")
  }

  return hex.EncodeToString(txHashBytes), int64(block.SeqNo), nil
}

func (c *Client) buildUpdateSignerCell(epochData *db.Epoch) (*cell.Cell, error) {
	x, y, err := chain.HexToCoordinates(epochData.Signer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pubkey")
	}

	rawSig := strings.TrimPrefix(epochData.Signature, "0x")
	signatureBytes, err := hex.DecodeString(rawSig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode signature")
	}

	signatureCell := cell.BeginCell().
		MustStoreSlice(signatureBytes, uint(len(signatureBytes)*8)).
		EndCell()

	body := cell.BeginCell().
		MustStoreUInt(updateSignerKeyBySignatureOpCode, 32).
		MustStoreUInt(signerRecoveryH, 8).
		MustStoreBigUInt(x, 256).
		MustStoreBigUInt(y, 256).
		MustStoreUInt(epochData.StartTime, 32).
		MustStoreUInt(epochData.EndTime, 32).
		MustStoreBoolBit(epochData.SignatureMode).
		MustStoreRef(signatureCell).
		EndCell()

	return body, nil
}
