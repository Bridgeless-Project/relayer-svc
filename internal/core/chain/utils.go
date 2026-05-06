package chain

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/pkg/errors"
)

func HexToCoordinates(pubkeyHex string) (*big.Int, *big.Int, error) {
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