package manifest

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
)

type ResourceValue struct {
	Val types.Int
}

func NewResourceValue(v uint64) ResourceValue {
	val := sdkmath.NewIntFromUint64(v)
	return ResourceValue{Val: val}
}
