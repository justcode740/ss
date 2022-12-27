package main

import (
	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
)

func getIndex(valIdx uint64, indexCount uint64, seed [32]byte) uint64 {
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

func reverseIndex(valIdx uint64, indexCount uint64, seed [32]byte) uint64 {
	originalIdx, err := helpers.ComputeShuffledIndex(types.ValidatorIndex(valIdx), indexCount, seed, false)
	if err != nil {
		panic(err)
	}
	return uint64(originalIdx)
}