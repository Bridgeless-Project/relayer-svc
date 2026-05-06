package evm

import (
	"crypto/ecdsa"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure/v3"
)

const (
	withdrawERC20             = "withdrawERC20"
	withdrawNative            = "withdrawNative"
	withdrawERC20Merkelized   = "withdrawERC20Merkelized"
	withdrawNativeMerkelized  = "withdrawNativeMerklelized"
	notAvailableBlockReceipts = "the method eth_getBlockReceipts does not exist/is not available"
)

type signerInfo struct {
	address    common.Address
	privateKey *ecdsa.PrivateKey
}

var EVMHooks = figure.Hooks{
	"[]*ecdsa.PrivateKey": func(raw interface{}) (reflect.Value, error) {
		switch value := raw.(type) {
		case []string:
			keys := make([]*ecdsa.PrivateKey, len(value))
			for i, str := range value {
				kp, err := crypto.HexToECDSA(str)
				if err != nil {
					return reflect.Value{}, errors.Wrap(err, "failed to init keypair")
				}
				keys[i] = kp
			}

			return reflect.ValueOf(keys), nil
		default:
			return reflect.Value{}, errors.Errorf("cant init keypair from type: %T", value)
		}
	},
}
