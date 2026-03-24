package evm

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

const hashlen = 32

func txHashToBytes32(txHash string) [32]byte {
	var res [32]byte
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
func merkleProofParsing(merkleProof string) ([][32]byte, error) {
	if merkleProof == "" {
		return make([][32]byte, 0), nil
	}

	var proof [][32]byte
	var proofsAsString []string
	if err := json.Unmarshal([]byte(merkleProof), &proofsAsString); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal merkle proof JSON")
	}

	for _, s := range proofsAsString {
		proofBytes, err := hexutil.Decode(s)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode hex")
		}
		if len(proofBytes) != hashlen {
			return nil, errors.New("invalid hash length, expected exactly 32 bytes")
		}

		var element [32]byte
		copy(element[:], proofBytes)
		proof = append(proof, element)
	}

	return proof, nil
}
