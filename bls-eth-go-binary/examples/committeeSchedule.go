package main

import (
	// "context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	// "github.com/prysmaticlabs/prysm/v3/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/v3/config/params"
	"github.com/prysmaticlabs/prysm/v3/crypto/hash"

	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
)
func getValidatorIdx(slot int, committeeIdx int) {
	// att.slot + att.communityIdx -> uniquely identify a community
	// 
	// var validators []types.ValidatorIndex
	// for i := 0; i <= 21062; i++ {
	// 	validators = append(validators, types.ValidatorIndex(i))
	// }	
	// validators := totalValidatorIndex()
	// fmt.Println(len(validators))
	// fmt.Println(len(validators))
	// bytes, _ := hex.DecodeString("11d84448ae8fd84292d51a0eb718e6318bce5055ac2bcabd8016f61719f9fbe2")
	// f, err := os.Open("examples/randaos_slots.csv")
	
    // if err != nil {
        
    // }
    // defer f.Close()

	// csvReader := csv.NewReader(f)
	// records, err := csvReader.ReadAll()
	// if err != nil{
	// 	fmt.Println("ERR")
	// }
	// // each row is an att
	// for i := 1; i < len(records); i++{
	// 	// fmt.Println(records[i])
	// 	str := records[i][1][2:]
	// 	bytes, _ := hex.DecodeString(str)
		
	// 	seed, _ := seed(bytes, types.Epoch(i+4), params.BeaconConfig().DomainBeaconAttester)
	// 	fmt.Println( getIndex(0, 21063, seed))
		

	// 	// for i := 20000; i < 50000; i++{
	// 	// 	// 33667
	// 	// if getIndex(0, uint64(i), seed)==16876 && getIndex(1, uint64(i), seed)==20508 {
	// 	// 	fmt.Println(i, seed)
	// 	// }
	// }
	
	// bytes, _ := hex.DecodeString("bf606fd377e1dae753845576d5452d6a591a7bf3bbb6409dc6d61e382e3ad749")
	// bytes, _ := hex.DecodeString("d8ab7b20cd0dbf12ce09670551e14e30860b658835885e8c794fa6f05da25b25")
	// bytes, _ := hex.DecodeString("9f9b1ed30050bcca663b41062db496c5d797b922ea459eb566bc0d46c1d66b24")
	// bytes, _ := hex.DecodeString("ae37c35f939f3ea47cebeb2e6689b31247215477ffeffb18fd4b7dbbce7aaf48")
	bytes, _ := hex.DecodeString("d8ab7b20cd0dbf12ce09670551e14e30860b658835885e8c794fa6f05da25b25")


	
	seed, _ := seed(bytes, types.Epoch(3199), params.BeaconConfig().DomainBeaconAttester)
	fmt.Println(getIndex(0, 33667, seed))
	// for i:=0; i<10; i++{
	// 	fmt.Println(getIndex(uint64(i), 21063, seed))

	// }
	
	
}
// seed start from 0x
func getFirstIndex(randaoMix string, epoch int, indexCount int){
	bytes, _ := hex.DecodeString(randaoMix[2:])
	seed, _ := seed(bytes, types.Epoch(epoch), params.BeaconConfig().DomainBeaconAttester)
	fmt.Println(getIndex(0, uint64(indexCount), seed))
}


	// for i:=20000; i<50000; i++{
	// 	if getIndex(0, uint64(i), seed)==16876{
	// 		fmt.Println(i)
	// 	}
		
	// }

	
	// fmt.Println(getIndex(12054, 21063, seed))
	// fmt.Println(seed)
	// validators := []types.ValidatorIndex{10, 9, 8}
	// committee, err := helpers.BeaconCommittee(
	// 	context.TODO(),
	// 	validators,
	// 	seed,
	// 	types.Slot(slot),
	// 	types.CommitteeIndex(committeeIdx),
	// )
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(committee)
	
	// fmt.Println(params.BeaconConfig().EpochsPerHistoricalVector)
	// fmt.Println(params.BeaconConfig().MinSeedLookahead - 1)
	




func totalValidatorIndex() []types.ValidatorIndex {
	// reader, _ := readFile(srv, f.Id)
	f, err := os.Open("examples/validator102368tmp.csv")
    if err != nil {
        
    }
    defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil{
		fmt.Println("ERR")
	}
	var validators []types.ValidatorIndex
	// each row is an att
	for i := 0; i < len(records); i++{
		// fmt.Println(records[i][0])
		idx, _ := strconv.Atoi(records[i][0])
		validators = append(validators, types.ValidatorIndex(idx))
	}

	return validators
}

func seed(randaoMix []byte, epoch types.Epoch, domain [bls.DomainByteLength]byte) ([32]byte, error) {
	seed := append(domain[:], bytesutil.Bytes8(uint64(epoch))...)
	seed = append(seed, randaoMix...)

	seed32 := hash.Hash(seed)

	return seed32, nil
}