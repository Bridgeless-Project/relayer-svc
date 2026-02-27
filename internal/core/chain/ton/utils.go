package ton

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	tonAddress "github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func getNetworkCell(network string) (*cell.Cell, error) {
	networkCell := cell.BeginCell()
	networkBytes, err := fillBytesToSize(network, networkCellSizeBytes, 0x00)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fill bytes")
	}

	if err = networkCell.StoreSlice(networkBytes, networkCellSizeBit); err != nil {
		return nil, errors.Wrap(err, "failed to store bytes")
	}

	return networkCell.EndCell(), nil
}

func getSignatureCell(signature string) (*cell.Cell, error) {
	signatureCell := cell.BeginCell()

	signatureBytes, err := hexutil.Decode(signature)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding signature")
	}

	err = signatureCell.StoreSlice(signatureBytes, signatureBitSize)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store signature bytes to cell")
	}

	return signatureCell.EndCell(), nil
}

func fillBytesToSize(str string, size int, fill byte) ([]byte, error) {
	if size == 0 {
		size = 32
	}
	if fill == 0 {
		fill = 0x00
	}
	raw := []byte(str)
	if len(raw) > size {
		return nil, errors.New(fmt.Sprintf("\"%s\" is longer than %d bytes", str, size))
	}

	buf := bytes.Repeat([]byte{fill}, size)
	copy(buf, raw)

	return buf, nil
}

func getAddressCell(addr string) (*cell.Cell, error) {
	address, err := tonAddress.ParseAddr(addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse address")
	}

	addressCell := cell.BeginCell()
	if err = addressCell.StoreAddr(address); err != nil {
		return nil, errors.Wrap(err, "failed to store address")
	}

	return addressCell.EndCell(), nil
}

func txHashToBytes32(txHash string) []byte {
	hashBytes, err := hexutil.Decode(txHash)
	if err != nil || len(hashBytes) != 32 {
		return crypto.Keccak256(([]byte)(txHash))
	}
	return hashBytes
}

func getPubkeyFromHex(pubkeyHex string) (*big.Int, *big.Int, error) {
	cleanHex := strings.TrimPrefix(pubkeyHex, "0x")
	pubKeyBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to decode pubkey hex")
	}

	if len(pubKeyBytes) != 65 || pubKeyBytes[0] != 4 {
		return nil, nil, errors.New("bad pubkey format")
	}

	x := new(big.Int).SetBytes(pubKeyBytes[1:33])
	y := new(big.Int).SetBytes(pubKeyBytes[33:65])

	return x, y, nil
}
