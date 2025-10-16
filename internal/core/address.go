package core

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/pkg/errors"
)

type Address string

func (a Address) String() string {
	return string(a)
}

func (a Address) Validate() error {
	_, _, err := bech32.DecodeAndConvert(a.String())

	return errors.Wrap(err, "failed to decode address")
}

func (a Address) Bytes() []byte {
	_, data, err := bech32.DecodeAndConvert(a.String())
	if err != nil {
		panic(err)
	}

	return data
}
