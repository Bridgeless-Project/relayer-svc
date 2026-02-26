package solana

import (
	"context"
	"strconv"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/pkg/errors"
)

func (c *Client) UpdateSigners(ctx context.Context, epochData *db.Epoch, signer *solana.Wallet) (string, int64, error) {
	sigArray, recoveryId, signerBytes, err := parseSignatureAndSigner(epochData.Signature, epochData.Signer)
	if err != nil {
		return "", 0, err
	}

	var newSigner [33]byte
	copy(newSigner[:], signerBytes)

	nonce, err := strconv.ParseUint(epochData.Nonce, 10, 64)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to parse nonce")
	}

	var reqType contract.UpdateSignerRequestType
	if epochData.SignatureMode {
		reqType = contract.UpdateSignerRequestTypeAdd
	} else {
		reqType = contract.UpdateSignerRequestTypeRemove
	}

	programID := solana.MustPublicKeyFromBase58(c.chain.BridgeAddress.String())
	contract.ProgramID = programID

	authorityPda, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte("authority"),
			[]byte(c.chain.Meta.BridgeId),
		},
		programID,
	)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to manually find authority pda")
	}

	instruction := contract.NewUpdateSignerInstruction(
		c.chain.Meta.BridgeId,
		reqType,
		newSigner,
		int64(epochData.StartTime),
		int64(epochData.EndTime),
		nonce,
		sigArray,
		recoveryId,
		authorityPda,
		signer.PublicKey(),
	)

	blockNumber, err := c.getLatestBlockWithRetry(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get latest block number")
	}

	hash, err := c.SendTx(ctx, instruction.Build(), signer)
	if err != nil {
		return "", blockNumber, errors.Wrap(err, "failed to send update signers tx")
	}

	return hash.String(), blockNumber, nil
}