package ton

import (
	"context"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton/wallet"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func (c *Client) UpdateSigners(ctx context.Context, epochData *db.Epoch, signer *wallet.Wallet) (string, int64, error) {
  ctxt := c.Chain.Client.Client().StickyContext(ctx)

  updateSignerCell, err := c.buildUpdateSignerCell(epochData)
  if err != nil {
    return "", 0, errors.Wrap(err, "failed to build update signer cell")
  }

  b, err := c.Chain.Client.GetMasterchainInfo(ctxt)
  if err != nil {
    return "", 0, errors.Wrap(err, "failed to get master chain info")
  }

  txHashBytes, err := signer.SendManyWaitTxHash(ctx, []*wallet.Message{
    {
      Mode: 1,
      InternalMessage: &tlb.InternalMessage{
        IHRDisabled: true,
        Bounce:      true,
        DstAddr:     c.Chain.BridgeContractAddress,
        Amount:      tlb.FromNanoTONU(200000000),
        Body:        updateSignerCell,
      },
    },
  })
  if err != nil {
    return "", 0, errors.Wrap(err, "error sending withdrawal transaction")
  }

  return hex.EncodeToString(txHashBytes), int64(b.SeqNo), nil
}

func (c *Client) buildUpdateSignerCell(epochData *db.Epoch) (*cell.Cell, error) {
	x, y, err := getPubkey(epochData.Signer)
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
		MustStoreUInt(0x12312324, 32).
		MustStoreUInt(4, 8).
		MustStoreBigUInt(x, 256).
		MustStoreBigUInt(y, 256).
		MustStoreUInt(epochData.StartTime, 32).
		MustStoreUInt(epochData.EndTime, 32).
		MustStoreBoolBit(epochData.SignatureMode).
		MustStoreRef(signatureCell).
		EndCell()

	return body, nil
}

func getPubkey(pubkeyHex string) (*big.Int, *big.Int, error) {
	cleanHex := strings.TrimPrefix(pubkeyHex, "0x")
	pubKeyBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid pubkey hex")
	}

	pubTON, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal uncompressed pubkey")
	}

	return pubTON.X, pubTON.Y, nil
}