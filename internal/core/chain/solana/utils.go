package solana

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/Bridgeless-Project/relayer-svc/internal/core/chain/solana/contract"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gagliardetto/solana-go"
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

func getUid(txHash string, nonce *big.Int) [32]byte {
	nonceBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(nonceBytes, nonce.Uint64())
	return sha256.Sum256(append([]byte(txHash), nonceBytes...))
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
