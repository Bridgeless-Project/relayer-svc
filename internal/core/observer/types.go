package observer

import (
	"github.com/Bridgeless-Project/relayer-svc/internal/db"
	"github.com/pkg/errors"
)

const (
	eventDepositSubmitted = "DEPOSIT_SUBMITTED"
	eventEpochUpdated     = "EPOCH_UPDATED"
)

var skippedDeposit = errors.New("skipped")

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type Event struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type MsgEvent struct {
	MsgIndex int     `json:"msg_index"`
	Events   []Event `json:"events"`
}

type BlockEvents struct {
	Deposits []*db.Deposit
	Epochs   []*db.Epoch
}
