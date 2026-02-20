package db

import "github.com/Bridgeless-Project/relayer-svc/internal/types"

const (
	AttributeKeyMerkleProof     = "merkle_proof"
	AttributeEpochId            = "epoch_id"
	AttributeTssInfo            = "tss_info"
	AttributeEpochSignature     = "epoch_signature"
	AttributeEpochSignatureData = "epoch_signature_data"
	AttributeEpochSigner        = "epoch_signer"
	AttributeEpochNonce         = "epoch_nonce"
	AttributeEpochStartTime     = "epoch_start_time"
	AttributeEpochEndTime       = "epoch_end_time"
	AttributeEpochSignatureMode = "epoch_signature_mode"

	AttributeEpochChainType        = "epoch_chain_type"
	AttributeChainId               = "chain_id"
	AttributeEpochSignatureAddress = "epoch_signature_address"
)

type Epoch struct {
	Id            uint32           `db:"id"`
	ChainId       string           `db:"chain_id"`
	Signature     string           `db:"signature"`
	Signer        string           `db:"signer"`
	StartTime     uint64           `db:"start_time"`
	EndTime       uint64           `db:"end_time"`
	Nonce         string           `db:"nonce"`
	SignatureMode bool             `db:"signature_mode"`
	Status        types.EpochStatus `db:"status"`
}

type EpochIdentifier struct {
	Id 		  uint32
	ChainId string
	Nonce		string
}

type EpochsQ interface {
	New() EpochsQ
	Insert(Epoch) (err error)
	Get(identifier EpochIdentifier) (*Epoch, error)
	GetWithStatus(status types.EpochStatus) ([]Epoch, error)

	UpdateStatus(EpochIdentifier, types.EpochStatus) error

	Transaction(f func() error) error
}