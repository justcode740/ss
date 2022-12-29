package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


// sig -> corrected validator idxs
func readCorrection(filename string) map[string][]int{
	// correction/all2.txt
	file, err := os.Open(filename)
    if err != nil {
       
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
	mp := map[string][]int{}
    // optionally, resize scanner's capacity for lines over 64K, see next example
    for scanner.Scan() {
        line := scanner.Text()
		splitted := strings.Split(line, ",")
		sig := strings.Trim(splitted[0], " ")
		vals := strings.Trim(splitted[1], " ")
		idxs := strings.Split(vals[1:len(vals)-1], " ")
		var varidxs []int 
		for _, idx := range idxs {
			id, _ := strconv.Atoi(idx)
			varidxs = append(varidxs, id)
		}
		fmt.Println(varidxs)
		mp[sig]=varidxs
    }

    if err := scanner.Err(); err != nil {
        
    }
	return mp
}

func bruteforceSearch() {
	readFile, _ := os.Open("allFalse.txt")
	fileScanner := bufio.NewScanner(readFile)
 
    fileScanner.Split(bufio.ScanLines)
	
	f, _ := os.Create(fmt.Sprintf("correction/all%s.txt", time.Now().String()))
    defer f.Close()

	c := 0
  
	for fileScanner.Scan() {
		fa := fileScanner.Text()
		i := strings.Index(fa, "\"")
		row := strings.Split(fa, ",")

		var validators []int
		if i > -1 {
			validatorsstr := fa[i+1:len(fa)-1]
			validators = strtouints(strings.Trim(validatorsstr, "\\[\\]"))
		}else{
			validators = strtouints(strings.Trim(row[12], "\\[\\]"))
		}
		bbr := row[1]
		sig := row[6]
		slot, _ := strconv.ParseUint(row[7], 10, 64)
		
		cidx, _ := strconv.ParseUint(row[5], 10, 64)
		sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
		sourceRoot := row[9]
		targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
		targetRoot := row[11]
		
		// get data root
		signing_root := getSigningRoot2(bbr[2:], sig[2:], uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot[2:], targetRoot[2:])
		// fmt.Println(signing_root)
		// get sig to verify
		sig = sig[2:]
		// only search for 1-val case
		if len(validators)==1{
			c++
			fmt.Println(fmt.Sprintf("search for validator %d", validators[0]))
			found := false
			var wg sync.WaitGroup
			wg.Add(1)
			for idx, pubkey := range allValidators{
				go func(valIdx int, pubkey string) {
					if found { return }
					var pks []string
					pks = append(pks, pubkey[2:])
					r := aggregateVerify(signing_root[:], pks, sig)
					if r {
						f.WriteString(fmt.Sprintf("%s, %d, %s\n", sig, valIdx, pubkey))
						fmt.Println(sig, valIdx, pubkey)
						found = true
						wg.Done()
					}
				}(idx, pubkey)
				

			}
			wg.Wait()
			if !found {
				fmt.Println(validators)
			}
		
		}
		


	}
	readFile.Close()
	fmt.Println(c)
}

func recover() {
	readFile, _ := os.Open("allFalse.txt")
	// readFile, _ := os.Open("examples/data-correction/t.csv")

	fileScanner := bufio.NewScanner(readFile)
 
    fileScanner.Split(bufio.ScanLines)
	
	f, _ := os.Create(fmt.Sprintf("correction/all%s.txt", time.Now().String()))
    defer f.Close()

	randaomap := readRandao()
  
	for fileScanner.Scan() {
		fa := fileScanner.Text()
		// i := strings.Index(fa, "\"")
		row := strings.Split(fa, ",")

		// var validators []int
		// if i > -1 {
		// 	validatorsstr := fa[i+1:len(fa)-1]
		// 	validators = strtouints(strings.Trim(validatorsstr, "\\[\\]"))
		// }else{
		// 	validators = strtouints(strings.Trim(row[12], "\\[\\]"))
		// }
		// fmt.Println(row[0][2:])
		res := ""
		aggregationBits := row[0][2:]
		for i:=0; i< len(aggregationBits)-1; i+=2{
			r, _ := hexToBin(aggregationBits[i:i+2])
			res = res + r
		}
		fmt.Println(res)

		// fmt.Println(res)

		blockslot, _ := strconv.Atoi(row[4])
		sig := row[6]
		slot, _ := strconv.ParseUint(row[7], 10, 64)
		
		cidx, _ := strconv.ParseUint(row[5], 10, 64)
		// sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
		// sourceRoot := row[9]
		// targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
		// targetRoot := row[11]
		
		// get data root
		// signing_root := getSigningRoot2(bbr[2:], sig[2:], uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot[2:], targetRoot[2:])
		// fmt.Println(signing_root)
		// get sig to verify
		sig = sig[2:]
		
		randaoMix := randaomap[blockslot][2:]

		epoch := blockslot / 32

		committeeSchedule := getCommitteeSchedule(
			int(slot), 
			int(cidx), 
			randaoMix, 
			fmt.Sprintf("examples/data-correction/validator_epoch_%vtmp.csv", epoch),
			epoch,
		)
		fmt.Println(len(committeeSchedule))
		var correctedValIdx []int
		for i := 0; i < len(res); i++ {
			if res[i] == '1' {
				correctedValIdx = append(correctedValIdx, committeeSchedule[i])
			}
		}
		fmt.Println(correctedValIdx)
		f.WriteString(fmt.Sprintf("%s, %d\n", sig, correctedValIdx))

	}
	readFile.Close()
}
func readRandao() map[int]string {
	f, err := os.Open("examples/data-correction/randao.csv")
    if err != nil {
        
    }
    defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil{
		fmt.Println(err)
	}
	randaoMap := map[int]string{}
	
	for i := 1; i < len(records); i++{
		// fmt.Println(records[i][0])
		blockslot, _ := strconv.Atoi(records[i][0])
		randaoMap[blockslot] = records[i][4] // include 0x
		
	}
	return randaoMap
}

func hexToBin(hex string) (string, error) {
	ui, err := strconv.ParseUint(hex, 16, 64)
	
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	n := len(hex) * 4
	


	// %016b indicates base 2, zero padded, with 16 characters
	tmp := fmt.Sprintf("%08b", ui)
	// fmt.Println(tmp)
	res := ""
	// 00000000 | 00000000 16
	for i := 8; i <= len(tmp); i+=8 {
		res = res + reverse(tmp[i-8:i])
	}
	for len(res) < n {
		res = "0" + res
	}
	return res, nil
}

func reverse(s string) (result string) {
	for _,v := range s {
	  result = string(v) + result
	}
	return 
  }

func rec(epoch int){
	// fetch randao from blockslot
	
	// compute seed
	// get validator index based on committee schedule epoch
	// blockslot // 32 -> validator_epoch_<epoch_num>tmp.csv
	// sorted to get active validator index
	// get committee based on slot and ciidx
	// use aggregation bit to recover validator idx
	// 
}

func verifyCorrected() {
	readFile, _ := os.Open("allFalse.txt")
	fileScanner := bufio.NewScanner(readFile)
 
    fileScanner.Split(bufio.ScanLines)
	
	f, _ := os.Create(fmt.Sprintf("correction/all%s.txt", time.Now().String()))
    defer f.Close()

	correction := readCorrection("correction/res.txt")
  
	for fileScanner.Scan() {
		fa := fileScanner.Text()
		// i := strings.Index(fa, "\"")
		row := strings.Split(fa, ",")

		// if i > -1 {
		// 	validatorsstr := fa[i+1:len(fa)-1]
		// 	validators = strtouints(strings.Trim(validatorsstr, "\\[\\]"))
		// }else{
		// 	validators = strtouints(strings.Trim(row[12], "\\[\\]"))
		// }
		bbr := row[1]
		sig := row[6]
		validators := correction[sig[2:]]
		var pks []string
		for _, val := range validators {
			pks = append(pks, allValidators[val][2:])
		}
		slot, _ := strconv.ParseUint(row[7], 10, 64)
		
		cidx, _ := strconv.ParseUint(row[5], 10, 64)
		sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
		sourceRoot := row[9]
		targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
		targetRoot := row[11]
		
		// get data root
		signing_root := getSigningRoot2(bbr[2:], sig[2:], uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot[2:], targetRoot[2:])
		// fmt.Println(signing_root)
		// get sig to verify
		sig = sig[2:]
		r := aggregateVerify(signing_root[:], pks, sig)
		if r {
			f.WriteString("true\n")
		}else{
			f.WriteString("false\n")
		}
			
		
		


	}
	readFile.Close()
	
}