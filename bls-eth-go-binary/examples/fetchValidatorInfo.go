package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
	"sort"
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

// read dict of id2addr to memory as golang map
func getValidatorMap() map[int]string {
	file, _ := ioutil.ReadFile("validatorinfo/res.json")
 
	data := map[int]string{}
 
	_ = json.Unmarshal([]byte(file), &data)
	return data
}


