package solana

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/Bridgeless-Project/relayer-svc/internal/core"
	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/pkg/errors"
)

func (c *Client) getWithdrawalPDA(withdrawalHash []byte) (*solana.PublicKey, error) {
	pda, _, err := solana.FindProgramAddress([][]byte{
		[]byte("withdraw"),
		withdrawalHash,
		[]byte(c.chain.Meta.BridgeId),
	}, contract.ProgramID)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve withdrawal PDA contract")
	}

	return &pda, nil
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

func getUid(txHash string, nonce *big.Int) [32]byte {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce.Uint64())
	return sha256.Sum256(append([]byte(txHash), nonceBytes...))
}

func (c *Client) SendTx(ctx context.Context, instruction solana.Instruction) (*solana.Signature, error) {
	recent, err := c.chain.Rpc.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get latest blockhash")
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			instruction,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(c.chain.OperatorWallet.PublicKey()),
	)

	if err != nil {
		return nil, errors.Wrap(err, "unable to create transaction")
	}

	sign, err := tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if c.chain.OperatorWallet.PublicKey().Equals(key) {
				return &c.chain.OperatorWallet.PrivateKey
			}
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign transaction")
	}

	// Send transaction, and wait for confirmation:
	signTx, err := c.chain.Rpc.SendTransaction(ctx, tx)
	if err != nil {
		return &sign[0], errors.Wrap(err, "unable to send transaction")
	}

	return &signTx, nil
}

func processSignature(signature string) ([64]uint8, uint8, error) {
	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return [64]uint8{}, 0, errors.Wrap(err, "invalid signature")
	}

	if len(signatureBytes) != 65 {
		return [64]uint8{}, 0, errors.Wrapf(err, "invalid signature length: %d, expected: 65", len(signatureBytes))
	}

	recId := signatureBytes[64]
	if recId >= 27 {
		recId -= 27
	}

	if recId > 3 {
		return [64]uint8{}, 0, errors.New("Invalid recovery ID after normalization")
	}

	var signatureUint8 [64]uint8
	copy(signatureUint8[:], signatureBytes[:64])
	return signatureUint8, recId, nil
}

func (c *Client) getWithdrawalContext(depositData db.Deposit) (*withdrawalContext, error) {
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

func parseTokenMetadata(data []byte) (*tokenMetadata, error) {
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		panic(err)
	}

	parsedInfo := parsed["parsed"].(map[string]any)["info"].(map[string]any)
	extensions := parsedInfo["extensions"].([]any)

	var tokenMetadata tokenMetadata
	for _, e := range extensions {
		ext := e.(map[string]any)
		extName := ext["extension"].(string)
		state := ext["state"].(map[string]any)

		switch extName {
		case "tokenMetadata":
			tokenMetadata.Name = state["name"].(string)
			tokenMetadata.Symbol = state["symbol"].(string)
			tokenMetadata.Uri = state["uri"].(string)

			addMeta := state["additionalMetadata"].([]any)
			for _, item := range addMeta {
				pair := item.([]any)
				if pair[0] == "nonce" {
					nonce, err := strconv.Atoi(pair[1].(string))
					if err != nil {
						return nil, errors.Wrap(err, "invalid nonce value")
					}

					tokenMetadata.Nonce = uint64(nonce)

				}
			}
		}
	}

	return &tokenMetadata, nil
}

func (c *Client) getSignHash(data db.Deposit) ([]byte, error) {
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
