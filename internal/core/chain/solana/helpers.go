package solana

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"strconv"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func (c *Client) getSignHash(data *db.Deposit) ([]byte, error) {
	amount, err := strconv.ParseUint(data.WithdrawalAmount, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse withdrawal amount")
	}
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBytes, amount)

	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(data.TxNonce))
	// unique id derived from deposit info
	uid := sha256.Sum256(append([]byte(data.TxHash), nonceBytes...))

	receiver, err := solana.PublicKeyFromBase58(data.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse receiver address")
	}

	buffer := []byte("withdraw")
	buffer = append(buffer, []byte(c.chain.Meta.BridgeId)...)
	buffer = append(buffer, amountBytes...)
	buffer = append(buffer, uid[:]...)
	buffer = append(buffer, receiver.Bytes()...)

	if data.WithdrawalToken != core.DefaultNativeTokenAddress {
		token, err := solana.PublicKeyFromBase58(data.WithdrawalToken)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer, token.Bytes()...)
	}

	hash := sha256.Sum256(buffer)
	return hash[:], nil
}

func (c *Client) getLatestBlockWithRetry(ctx context.Context) (int64, error) {
	var (
		block *rpc.GetLatestBlockhashResult
		err   error
	)

	getBlock := func() error {
		block, err = c.chain.Rpc.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
		if err != nil {
			return err
		}

		return nil
	}

	if err = core.DoWithRetry(ctx, getBlock); err != nil {
		return 0, errors.Wrap(err, "failed to get block")
	}

	return int64(block.Value.LastValidBlockHeight), nil
}

func (c *Client) getWithdrawalContext(depositData *db.Deposit) (*withdrawalContext, error) {
	signatureBytes, recId, err := processSignature(depositData.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "unable to process signature")
	}

	uidBytes := getUid(depositData.TxHash, big.NewInt(depositData.TxNonce))
	uid := [32]uint8{}
	copy(uid[:], uidBytes[:])

	withdrawalHash, err := c.getSignHash(depositData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get withdrawal hash")
	}
	var withdrawalHashBytes [32]uint8
	copy(withdrawalHashBytes[:], withdrawalHash)

	pda, err := c.getWithdrawalPDA(withdrawalHash)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get withdrawal pda")
	}

	amount, ok := big.NewInt(0).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("unable to parse withdrawal amount")
	}

	receiver, err := solana.PublicKeyFromBase58(depositData.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse receiver address")
	}

	var tokenAcccount solana.PublicKey
	if depositData.WithdrawalToken != core.DefaultNativeTokenAddress {
		tokenAcccount, err = solana.PublicKeyFromBase58(depositData.WithdrawalToken)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse withdrawal token")
		}
	}

	authority, _, err := solana.FindProgramAddress([][]byte{
		[]byte("authority"),
		[]byte(c.chain.Meta.BridgeId),
	}, contract.ProgramID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find authority account")
	}

	return &withdrawalContext{
		Receiver:         receiver,
		Token:            tokenAcccount,
		Authority:        authority,
		WithdrawalPDA:    *pda,
		Amount:           amount.Uint64(),
		BridgeID:         c.chain.Meta.BridgeId,
		UID:              uid,
		Sig:              signatureBytes,
		RecID:            recId,
		WithdrawalTxHash: withdrawalHashBytes,
		TxNonce:          uint64(depositData.TxNonce),
	}, nil
}

func (c *Client) getTokenMetadata(tokenAccount solana.PublicKey) (*tokenMetadata, error) {
	accountInfo, err := c.chain.Rpc.GetAccountInfoWithOpts(context.TODO(), tokenAccount, &rpc.GetAccountInfoOpts{
		Encoding: solana.EncodingJSONParsed,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get account info")
	}

	return parseTokenMetadata(accountInfo.Value.Data.GetRawJSON())
}

func (c *Client) getWithdrawalHash(depositData db.Deposit) ([]byte, error) {
	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("could not convert withdrawal amount to big.Int")
	}
	amountBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountBytes, amount.Uint64())

	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, uint64(depositData.TxNonce))

	uid := getUid(depositData.TxHash, big.NewInt(depositData.TxNonce))

	receiver, err := solana.PublicKeyFromBase58(depositData.Receiver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse receiver address")
	}

	buffer := []byte("withdraw")
	buffer = append(buffer, []byte(c.chain.Meta.BridgeId)...)
	buffer = append(buffer, amountBytes...)
	buffer = append(buffer, uid[:]...)
	buffer = append(buffer, receiver.Bytes()...)

	if depositData.WithdrawalToken != core.DefaultNativeTokenAddress {
		token, err := solana.PublicKeyFromBase58(depositData.WithdrawalToken)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer, token.Bytes()...)
	}

	hash := sha256.Sum256(buffer)
	return hash[:], nil
}
