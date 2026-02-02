package grpc

import (
	"github.com/Bridgeless-Project/relayer-svc/internal/api/types"
)

var _ types.APIServer = Implementation{}

type Implementation struct{}
