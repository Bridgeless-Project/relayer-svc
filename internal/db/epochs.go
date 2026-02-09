package db

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
	Id            uint32
	ChainId       string
	Signature     string
	Signer        string
	StartTime     uint64
	EndTime       uint64
	Nonce         string
	SignatureMode bool
}
