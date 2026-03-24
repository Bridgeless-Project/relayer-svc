package evm

import (
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

func (c *Client) getWithdrawalTxHash(transactOpts *bind.TransactOpts, data []byte) (string, error) {
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    transactOpts.Nonce.Uint64(),
		Gas:      transactOpts.GasLimit,
		To:       &c.chain.BridgeAddress,
		Value:    transactOpts.Value,
		GasPrice: transactOpts.GasPrice,
		Data:     data,
	})

	signTx, err := transactOpts.Signer(transactOpts.From, tx)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign transaction")
	}

	return signTx.Hash().Hex(), nil

}

func (c *Client) getWithdrawalTxData(method string, depositData *db.Deposit) ([]byte, error) {
	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("failed to parse withdrawal amount")
	}

	receiverAdress := common.HexToAddress(depositData.Receiver)

	signatureBytes, err := hexutil.Decode(depositData.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode signature")
	}

	proof, err := merkleProofParsing(depositData.MerkleProof)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse MerkleProof")
	}

	switch method {
	case withdrawNative:
		return c.abi.Pack(
			withdrawNative,
			amount,
			receiverAdress,
			txHashToBytes32(depositData.TxHash),
			big.NewInt(depositData.TxNonce),
			[][]byte{signatureBytes},
		)

	case withdrawERC20:
		common.HexToAddress(depositData.WithdrawalToken)
		return c.abi.Pack(
			withdrawERC20,
			common.HexToAddress(depositData.WithdrawalToken),
			amount,
			receiverAdress,
			txHashToBytes32(depositData.TxHash),
			big.NewInt(depositData.TxNonce),
			depositData.IsWrappedToken,
			[][]byte{signatureBytes},
		)

	case withdrawERC20Merkelized:
		common.HexToAddress(depositData.WithdrawalToken)
		return c.abi.Pack(
			withdrawERC20Merkelized,
			common.HexToAddress(depositData.WithdrawalToken),
			amount,
			receiverAdress,
			txHashToBytes32(depositData.TxHash),
			big.NewInt(depositData.TxNonce),
			depositData.IsWrappedToken,
			proof,
			[][]byte{signatureBytes},
		)

	case withdrawNativeMerkelized:
		return c.abi.Pack(
			withdrawNativeMerkelized,
			amount,
			receiverAdress,
			txHashToBytes32(depositData.TxHash),
			big.NewInt(depositData.TxNonce),
			proof,
			[][]byte{signatureBytes},
		)

	default:
		return nil, errors.New("unknown method")
	}
}
