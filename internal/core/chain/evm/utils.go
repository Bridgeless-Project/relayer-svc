package evm

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func txHashToBytes32(txHash string) [crypto.DigestLength]byte {
	var res [crypto.DigestLength]byte
	hashBytes, err := hexutil.Decode(txHash)
	if err != nil || len(hashBytes) != crypto.DigestLength {
		bytes := crypto.Keccak256(([]byte)(txHash))
		copy(res[:], bytes)
		return res
	}

	copy(res[:], hashBytes)
	return res
}

// returns empty slice to avoid panic in case merkle proof is non existent
func merkleProofParsing(merkleProof string) ([][crypto.DigestLength]byte, error) {
	if merkleProof == "" {
		return make([][crypto.DigestLength]byte, 0), nil
	}

	var proof [][crypto.DigestLength]byte
	var proofsAsString []string
	if err := json.Unmarshal([]byte(merkleProof), &proofsAsString); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal merkle proof JSON")
	}

	for _, hashStr := range proofsAsString {
		proofBytes, err := hexutil.Decode(hashStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode hex")
		}
		if len(proofBytes) != crypto.DigestLength {
			return nil, errors.New("invalid hash length, expected exactly 32 bytes")
		}

		var element [crypto.DigestLength]byte
		copy(element[:], proofBytes)
		proof = append(proof, element)
	}

	return proof, nil
}
