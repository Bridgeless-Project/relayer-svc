package types

type EpochStatus int32

const (
	EpochStatus_EPOCH_STATUS_FAILED			        EpochStatus = 0
	EpochStatus_EPOCH_STATUS_PROCESSED          EpochStatus = 1
	EpochStatus_EPOCH_STATUS_PENDING            EpochStatus = 2
)