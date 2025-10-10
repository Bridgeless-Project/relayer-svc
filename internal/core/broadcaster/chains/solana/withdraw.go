package solana

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"strconv"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/broadcaster"
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/gagliardetto/solana-go"

	"github.com/pkg/errors"
)

func (p *Client) WithdrawalAmountValid(amount *big.Int) bool {
	// Solana token amounts are uint64, bigger (or negative) numbers are invalid
	if !amount.IsUint64() {
		return false
	}
	return amount.Cmp(broadcaster.ZeroAmount) == 1
}

func (p *Client) GetSignHash(data db.Deposit) ([]byte, error) {
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
	buffer = append(buffer, []byte(p.chain.Meta.BridgeId)...)
	buffer = append(buffer, amountBytes...)
	buffer = append(buffer, uid[:]...)
	buffer = append(buffer, receiver.Bytes()...)

	if data.WithdrawalToken != broadcaster.DefaultNativeTokenAddress {
		token, err := solana.PublicKeyFromBase58(data.WithdrawalToken)
		if err != nil {
			return nil, err
		}
		buffer = append(buffer, token.Bytes()...)
	}

	hash := sha256.Sum256(buffer)
	return hash[:], nil
}

func (p *Client) WithdrawNative(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *Client) WithdrawToken(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *Client) WithdrawWrapped(depositData db.Deposit) (txHash string, err error) {
	//TODO implement me
	panic("implement me")
}
