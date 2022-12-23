package main

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
)

func tt(valIdx uint64, indexCount uint64, seed [32]byte) uint64 {
	shuffedIdx, err := helpers.ShuffledIndex(
		types.ValidatorIndex(valIdx),
		indexCount,
		seed,
	)
	if err != nil{
		panic(err)
	}
	return uint64(shuffedIdx)
}