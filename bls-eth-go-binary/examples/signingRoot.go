package main

import (
	"encoding/hex"
	"fmt"

	"github.com/prysmaticlabs/prysm/v3/beacon-chain/core/signing"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v3/testing/util"
)

func getSigningRoot() [32]byte {
	blockroot, _ := hex.DecodeString("1895bac33c8cecdf1c07a21fa0c1be683283a1c99f5964ed7bfeccc33104ecba")
	sourceroot, _ := hex.DecodeString("ddead563586b895ec62d57511327b189535ac6209de9e76ff20816146d780711")
	targetroot, _ := hex.DecodeString("70e6b618b1c11e1615faa06272b5d02761b9c157e342596d571bc37e8cf4a0da")
	sig, _ := hex.DecodeString("89761ec8e2a086fb073204f9d48ed1a8e805e9871de9be0bac91a5f76aa5f261c2f748f50f48eb4a7b40a32998227f440c6e89302c9260cdb08837cbedd2d50ab92e53155679905ddf1a7a828d8fcc56a4d56c4543caf0eaa66aaa6778ac8c30")
	att1 := util.HydrateIndexedAttestation(&ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Slot: types.Slot(99999),
			CommitteeIndex: types.CommitteeIndex(7),
			BeaconBlockRoot: blockroot,
			Source: &ethpb.Checkpoint{
				Epoch: types.Epoch(3123),
				Root: sourceroot,},
			Target: &ethpb.Checkpoint{Epoch: types.Epoch(3124),
			Root: targetroot,},
		},
		Signature: sig,
		AttestingIndices: []uint64{21521, 128},
	})
	previousVersion, _ := hex.DecodeString("00000000")
	currentVersion, _ :=  hex.DecodeString("00000000")
	fork := &ethpb.Fork{
		PreviousVersion: previousVersion,
		CurrentVersion: currentVersion,
		Epoch: types.Epoch(0),
	}
	genesisValidatorsRoot, _ := hex.DecodeString("4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95")
	// fmt.Println("domain type",  params.BeaconConfig().DomainBeaconAttester)
	
	domain, err := signing.Domain(fork, att1.Data.Target.Epoch, params.BeaconConfig().DomainBeaconAttester, genesisValidatorsRoot)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// fmt.Println("att", att1.Data)
	
	signing_root, err:= signing.ComputeSigningRoot(att1.Data, domain)
	fmt.Println("test", hex.EncodeToString(signing_root[:]))
	if err != nil {
		panic(err)
	}

	return signing_root
}


// beaconroot hex string without 0x, sig hex string without 0x, 
func getSigningRoot2(beaconroot string, signature string, slot uint, commmitteeIndex uint, sourceEpoch uint, targetEpoch uint, source string, target string) [32]byte {
	blockroot, _ := hex.DecodeString(beaconroot)
	sourceroot, _ := hex.DecodeString(source)
	targetroot, _ := hex.DecodeString(target)
	sig, _ := hex.DecodeString(signature)
	att1 := util.HydrateIndexedAttestation(&ethpb.IndexedAttestation{
		Data: &ethpb.AttestationData{
			Slot: types.Slot(slot),
			CommitteeIndex: types.CommitteeIndex(commmitteeIndex),
			BeaconBlockRoot: blockroot,
			Source: &ethpb.Checkpoint{
				Epoch: types.Epoch(sourceEpoch),
				Root: sourceroot,
			},
			Target: &ethpb.Checkpoint{
				Epoch: types.Epoch(targetEpoch),
				Root: targetroot,
			},
		},
		Signature: sig,
		AttestingIndices: []uint64{0, 1},
	})
	// slot 0-(74239+1)*32-1 is 0,0, 
	// 74240*32-(144895+1)*32-1 is 0,1, 
	// 144896*32-last is 1,2
	var previousVersion []byte
	var currentVersion []byte
	if slot >=0 && slot <= 74240 * 32 - 1 {
		previousVersion, _ = hex.DecodeString("00000000")
		currentVersion, _ =  hex.DecodeString("00000000")
	}else if slot >= 74240 * 32 && slot <= 144896*32-1 {
		previousVersion, _ = hex.DecodeString("00000000")
		currentVersion, _ =  hex.DecodeString("01000000")
	}else if slot >= 144896*32 {
		previousVersion, _ = hex.DecodeString("01000000")
		currentVersion, _ =  hex.DecodeString("02000000")
	}
	
	fork := &ethpb.Fork{
		PreviousVersion: previousVersion,
		CurrentVersion: currentVersion,
		Epoch: types.Epoch(0),
	}
	genesisValidatorsRoot, _ := hex.DecodeString("4b363db94e286120d76eb905340fdd4e54bfe9f06bf33ff6cf5ad27f511bfe95")
	// fmt.Println("domain type",  params.BeaconConfig().DomainBeaconAttester)

	domain, err := signing.Domain(fork, types.Epoch(0), params.BeaconConfig().DomainBeaconAttester, genesisValidatorsRoot)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// fmt.Println("att", att1.Data)
	
	signing_root, err:= signing.ComputeSigningRoot(att1.Data, domain)
	// fmt.Println("test", hex.EncodeToString(signing_root[:]))
	if err != nil {
		panic(err)
	}

	return signing_root
}
