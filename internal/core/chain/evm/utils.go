package evm

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

const hashlen = crypto.DigestLength

func txHashToBytes32(txHash string) [hashlen]byte {
	var res [hashlen]byte
	hashBytes, err := hexutil.Decode(txHash)
	if err != nil || len(hashBytes) != hashlen {
		bytes := crypto.Keccak256(([]byte)(txHash))
		copy(res[:], bytes)
		return res
	}

	copy(res[:], hashBytes)
	return res
}

// returns empty slice to avoid panic in case merkle proof is non existent
func merkleProofParsing(merkleProof string) ([][hashlen]byte, error) {
	if merkleProof == "" {
		return make([][hashlen]byte, 0), nil
	}

	var proof [][hashlen]byte
	var proofsAsString []string
	if err := json.Unmarshal([]byte(merkleProof), &proofsAsString); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal merkle proof JSON")
	}

	for _, hashStr := range proofsAsString {
		proofBytes, err := hexutil.Decode(hashStr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode hex")
		}
		if len(proofBytes) != hashlen {
			return nil, errors.New("invalid hash length, expected exactly 32 bytes")
		}

		var element [hashlen]byte
		copy(element[:], proofBytes)
		proof = append(proof, element)
	}

	return proof, nil
}
