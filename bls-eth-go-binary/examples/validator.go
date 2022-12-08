package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

// read aray of validator idx, batch fetch validator address, write the map to a json
func fetchValidatorInfo(idloc string){
	file, _ := ioutil.ReadFile(idloc+"/validators.json")
 
	var validators []int
 
	_ = json.Unmarshal([]byte(file), &validators)
	// the validator api return is sorted by idx, not in query order
	sort.Ints(validators)
	eth2Client := newEth2Client()
	id2addr := make(map[int]string)
	i := 0
	n := len(validators)
	fmt.Println("len", n)
	reqTime := 0

	for(i < n) {
		endIdx := min(i + 100, n)
		reqTime += 1
		keys := eth2Client.GetValidatorPubKey(validators[i: endIdx])
		idx := 0
		for j := i; j  < endIdx; j++{
			id2addr[validators[j]] = keys[idx]
			idx++
		}
		i = i+100
		if reqTime % 10 == 0 {
			time.Sleep(60 * time.Second)
		}
	}

	val, _ := json.MarshalIndent(id2addr, "", " ")
	_ = ioutil.WriteFile(idloc+"/res.json", val, 0644)	
}

const validatorIdsLoc = "validatorinfo/validatorIds.json"
func findValidatorsOfInterestDoubleVote(){
	f, err := excelize.OpenFile("./data/unslashed_double_votes.xlsx")
	if err!=nil {
		fmt.Println(err)
	}
	var all_validators []int
	// 405
	for i := 2; i < 10; i++ {
		idx := strconv.Itoa(i)
		validatorIdx, _ := f.GetCellValue("unslashed_double_votes", "B"+idx)
		validator, _ := strconv.Atoi(validatorIdx)
		blocksStr, _ := f.GetCellValue("unslashed_double_votes", "C"+idx)
		var blocks []int
		err := json.Unmarshal([]byte(blocksStr), &blocks)
		if err != nil {
			fmt.Println(err)
		}
		blocks = removeDuplicateInt(blocks)
		for _, blockSlot := range blocks {
			res := readBlockInfo(blockInfoDataFolder, uint(blockSlot))
			c := 0
			for _, attestation := range (res.Data) {
				if contains(attestation.Validators, validator) {
					all_validators = append(all_validators, attestation.Validators...)
					c++
				}
			}
			if c==2 {
				fmt.Print(strconv.FormatInt(int64(validator), 10)+ "  ")
			}
			fmt.Print(strconv.FormatInt(int64(c), 10)+ "  ")	
		}
		fmt.Println()
	}
	res := removeDuplicateInt(all_validators)
	bytes, _ := json.Marshal(res)
	ioutil.WriteFile("valtest50/validators", bytes, 0644)
}

func writeValidatorDict(d map[int]string){
	bytes, _ := json.Marshal(d)
	ioutil.WriteFile("data/allValidators", bytes, 0644)
}

// read dict of id2addr to memory as golang map
func getValidatorMap() map[int]string {
	file, _ := ioutil.ReadFile("validatorinfo/res.json")
 
	data := map[int]string{}
 
	_ = json.Unmarshal([]byte(file), &data)
	return data
}

// read 125mb csv
func getValidatorMap2() map[int]string {
	file, err := os.Open("data/validator_key_all.csv")
    if err != nil {
        panic(err)
    }
	data := map[int]string{}

    parser := csv.NewReader(file)
	line := 0
	for {
		if line == 0 {
			line ++
			continue
		}
		record, err := parser.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		idx, _ := strconv.Atoi(record[10])
		data[idx] = record[7]
		line ++
	}
	return data
}

func getValidatorMap3() map[int]string {
	file, _ := ioutil.ReadFile("data/allValidators.json")
	var validators map[int]string
	_  = json.Unmarshal([]byte(file), &validators)
	return validators
}

