package ton

import (
	"context"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

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

	var h uint64 = 0x04
	x, y, err := getPubkey(epochData.Id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pubkey")
	}

	cell0 := cell.BeginCell().
		MustStoreUInt(h, 8).
		MustStoreBigUInt(x, 256).
		MustStoreBigUInt(y, 256)

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

	var h uint64 = 0x04
	x, y, err := getPubkey(epoch.Id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pubkey")
	}

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

type getPubkeyResponse struct {
	Pubkey string `json:"pub_key"`
}

func getPubkey(epochId uint32) (*big.Int, *big.Int, error) {
	// TODO: replace hardcoded url with API url from config
	url := fmt.Sprintf("http://localhost:1317​cosmos​/bridge​/epoch​/​%d/pubkey", epochId)
	res, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	var resStruct getPubkeyResponse
	if err := json.NewDecoder(res.Body).Decode(&resStruct); err != nil {
		return nil, nil, err
	}

	privKeyBytes := sha256.Sum256([]byte(resStruct.Pubkey))
	privKey, err := ecdh.P256().NewPrivateKey(privKeyBytes[:])
	pubKey := privKey.PublicKey()
	pubKeyBytes := pubKey.Bytes()
	x := new(big.Int).SetBytes(pubKeyBytes[1:33])
	y := new(big.Int).SetBytes(pubKeyBytes[33:65])
	return x, y, nil
}
