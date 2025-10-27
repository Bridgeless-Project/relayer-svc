package evm

import (
	"math/big"

	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
)

func txHashToBytes32(txHash string) [32]byte {
	var res [32]byte
	hashBytes, err := hexutil.Decode(txHash)
	if err != nil || len(hashBytes) != 32 {
		bytes := crypto.Keccak256(([]byte)(txHash))
		copy(res[:], bytes)
		return res
	}

	copy(res[:], hashBytes)
	return res
}

func (c *Client) getWithdrawalTxHash(method string, transactOpts *bind.TransactOpts, deposit db.Deposit) (string, error) {
	chainId, ok := big.NewInt(0).SetString(deposit.ChainId, 10)
	if !ok {
		return "", errors.New("failed to set chain id to big")
	}

	data, err := c.getWithdrawalTxData(method, deposit)
	if err != nil {
		return "", errors.Wrap(err, "failed to get withdrawal tx data")
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     transactOpts.Nonce.Uint64(),
		GasTipCap: transactOpts.GasTipCap,
		GasFeeCap: transactOpts.GasFeeCap,
		Gas:       transactOpts.GasLimit,
		To:        &c.chain.BridgeAddress,
		Value:     transactOpts.Value,
		Data:      data,
	})

	signer := types.LatestSigner(&params.ChainConfig{ChainID: chainId})
	signTx, err := types.SignTx(tx, signer, c.chain.OperatorPrivKey)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign transaction")
	}

	return signTx.Hash().Hex(), nil

}

func (c *Client) getWithdrawalTxData(method string, depositData db.Deposit) ([]byte, error) {
	amount, ok := new(big.Int).SetString(depositData.WithdrawalAmount, 10)
	if !ok {
		return nil, errors.New("failed to parse withdrawal amount")
	}

	receiverAdress := common.HexToAddress(depositData.Receiver)

	signatureBytes, err := hexutil.Decode(depositData.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode signature")
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

	default:
		return nil, errors.New("unknown method")
	}
}
