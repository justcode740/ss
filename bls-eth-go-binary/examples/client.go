package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"io/ioutil"
)
type eth2Client struct {
	
}
	
type BlockAttestationsRes struct {
	Status string `json:"status"`
	Data   []struct {
		Aggregationbits string `json:"aggregationbits"`
		Beaconblockroot string `json:"beaconblockroot"`
		BlockIndex      int    `json:"block_index"`
		BlockRoot       string `json:"block_root"`
		BlockSlot       int    `json:"block_slot"`
		Committeeindex  int    `json:"committeeindex"`
		Signature       string `json:"signature"`
		Slot            int    `json:"slot"`
		SourceEpoch     int    `json:"source_epoch"`
		SourceRoot      string `json:"source_root"`
		TargetEpoch     int    `json:"target_epoch"`
		TargetRoot      string `json:"target_root"`
		Validators      []int  `json:"validators"`
	} `json:"data"`
}

type ValidatorBatchRes struct {
	Status string `json:"status"`
	Data   []struct {
		Activationeligibilityepoch int         `json:"activationeligibilityepoch"`
		Activationepoch            int         `json:"activationepoch"`
		Balance                    int64       `json:"balance"`
		Effectivebalance           int64       `json:"effectivebalance"`
		Exitepoch                  int64       `json:"exitepoch"`
		Lastattestationslot        int         `json:"lastattestationslot"`
		Name                       interface{} `json:"name"`
		Pubkey                     string      `json:"pubkey"`
		Slashed                    bool        `json:"slashed"`
		Status                     string      `json:"status"`
		Validatorindex             int         `json:"validatorindex"`
		Withdrawableepoch          int64       `json:"withdrawableepoch"`
		Withdrawalcredentials      string      `json:"withdrawalcredentials"`
	} `json:"data"`
}

type ValidatorRes struct {
	Status string `json:"status"`
	Data   struct {
		Activationeligibilityepoch int         `json:"activationeligibilityepoch"`
		Activationepoch            int         `json:"activationepoch"`
		Balance                    int64       `json:"balance"`
		Effectivebalance           int64       `json:"effectivebalance"`
		Exitepoch                  int64       `json:"exitepoch"`
		Lastattestationslot        int         `json:"lastattestationslot"`
		Name                       interface{} `json:"name"`
		Pubkey                     string      `json:"pubkey"`
		Slashed                    bool        `json:"slashed"`
		Status                     string      `json:"status"`
		Validatorindex             int         `json:"validatorindex"`
		Withdrawableepoch          int64       `json:"withdrawableepoch"`
		Withdrawalcredentials      string      `json:"withdrawalcredentials"`
	} `json:"data"`
}

func newEth2Client() *eth2Client {
	return &eth2Client{}
}

// Batch retrive a slice of validator's pubkeys from idxs
func (eth2Client *eth2Client) GetValidatorPubKey(validatorIdxs []int) []string {
	url := fmt.Sprintf("https://beaconcha.in/api/v1/validator/%s", toCommaStr(validatorIdxs))
	resp, err := http.Get(url)
	if err != nil {
        fmt.Println("No response from request")
    }
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if len(validatorIdxs) <= 1 {
		var result ValidatorRes
		if err := json.Unmarshal(body, &result); err != nil {  // Parse []byte to the go struct pointer
			fmt.Println("Can not unmarshal JSON")
		}
		var res []string
		res = append(res, result.Data.Pubkey)
		return res
	}else{
		var result ValidatorBatchRes
		if err := json.Unmarshal(body, &result); err != nil {  // Parse []byte to the go struct pointer
			fmt.Println("Can not unmarshal JSON")
		}
		var res []string
		for _, rec := range result.Data {
			res = append(res, rec.Pubkey)
		}
		return res
	}
}

func (eth2Client *eth2Client) GetAttestationsForBlock(blockSlot uint, validator int) (string, []string, string, error) {
	resp, err := http.Get(fmt.Sprintf("http://beaconcha.in/api/v1/block/%d/attestations", blockSlot))
	if err != nil {
        fmt.Println("No response from request")
    }
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var result BlockAttestationsRes
    if err := json.Unmarshal(body, &result); err != nil {  // Parse []byte to the go struct pointer
        fmt.Println("Can not unmarshal JSON")
    }
	for _, rec := range result.Data {
		// if rec.Committeeindex == 24 {
		// 	fmt.Println(rec.Validators)
		// }
		if (contains(rec.Validators, validator)) {
			var pubKeys []string
			i := 0
			n := len(rec.Validators)
			for(i < n) {
				endIdx := min(i + 100, n)
				keys := eth2Client.GetValidatorPubKey(rec.Validators[i: endIdx])
				pubKeys = append(pubKeys, keys...) 
				// fmt.Println(pubKeys)
				i = i+100
				time.Sleep(1 * time.Second)
			}
			if (len(pubKeys)!=n){
				fmt.Println(len(pubKeys), n)
				panic("miss some keys")
			}
			// if (len(rec.Validators) > 100) {
			// 	fmt.Println(len(rec.Validators))
			// 	return "", nil, "", errors.New("validator len exceed max api batch size 100")
			// }
			// pubkeys := eth2Client.GetValidatorPubKey(rec.Validators)
			return rec.Beaconblockroot, pubKeys, rec.Signature, nil
		}
	}
	// fmt.Println(result.Data)
	return "", nil, "", errors.New(fmt.Sprintf("validator %d not found in the %d block's attestation, maybe source data is wrong?", validator, blockSlot))
}

// -------------------------helper ------------------------------------------
func contains(ids []int, id int) bool {
	for _, idd := range ids {
		if idd == id {
			return true
		}
	}
	return false
}

func toCommaStr(ids []int) string {
	var IDs []string
	for _, i := range ids {
		IDs = append(IDs, strconv.Itoa(i))
	}
	return strings.Join(IDs, ",")
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

