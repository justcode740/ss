package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

var blockSlots = []uint{1041100, 1041103, 1041107, 1041104, 1267818, 1054792, 1267816, 608067, 918915, 1054791, 1054788, 1041098, 
1054787, 1267819, 1054786, 1041108, 1041106, 1054790, 1054789, 1041105, 1041099, 1041102, 1267820, 918913, 
988962, 918918, 918917, 1041101, 1267817, 918922, 988961, 918882, 608052, 608059, 918901, 988955,
988965, 918883, 988931, 918900, 918894, 918902, 918907, 988935, 918896, 918908, 918892, 918898,
918903, 988959, 988972, 918909, 918895, 918910, 918887, 918891, 918886, 608043, 608040, 918889,
918911, 918881, 988946, 988936, 988974, 918899, 918904, 918905, 918893, 918916, 918906, 988940,
608057, 988930, 918890, 918884}

var all_blocks = []uint{
	1041100,1041108,1041103,1041107,1041104,1267818,1267822,1054792,1054796,1267816,608067,608085,918915,918922,1054791,1054788,1041098,1054787,1267819,1054786,1041109,1041106,1054790,1054789,1041105,1041099,918916,1041102,1041127,1267820,918913,988962,988965,988976,918918,1041117,918917,1041101,1267817,918938,1041116,1041113,1041128,1054798,1054802,988961,1267841,1041114,1041118,1054805,1054815,918882,608052,608059,918901,988955,918883,988931,918900,918894,918902,918907,988935,918896,918908,918892,918898,918903,988959,988972,918909,918895,918910,918887,918891,918886,608043,608040,918889,918911,918881,988946,988936,988974,918899,918904,918905,918893,918906,988940,608057,988930,918890,918884}
var allBlocks = []int{
		1041100,1041108,1041103,1041107,1041104,1267818,1267822,1054792,1054796,1267816,608067,608085,918915,918922,1054791,1054788,1041098,1054787,1267819,1054786,1041109,1041106,1054790,1054789,1041105,1041099,918916,1041102,1041127,1267820,918913,988962,988965,988976,918918,1041117,918917,1041101,1267817,918938,1041116,1041113,1041128,1054798,1054802,988961,1267841,1041114,1041118,1054805,1054815,918882,608052,608059,918901,988955,918883,988931,918900,918894,918902,918907,988935,918896,918908,918892,918898,918903,988959,988972,918909,918895,918910,918887,918891,918886,608043,608040,918889,918911,918881,988946,988936,988974,918899,918904,918905,918893,918906,988940,608057,988930,918890,918884}

const blockInfoDataFolder = "blockinfo/"
const blockInfoFilePrefix = "blockAttestations"
const blockInfoFilePrefix2 = "epochblockAttestations"

func writeBlockInfo(){
	eth2Client := newEth2Client()
	for i := 0; i < len(all_blocks); i++ {
		// every 10 req, wait 1 min
		if (i != 0 && i % 10 == 0){
			time.Sleep(60 * time.Second)
		}
		res := eth2Client.fetchBlockInfo(all_blocks[i])
		file, _ := json.MarshalIndent(res, "", " ")
		_ = ioutil.WriteFile(blockInfoDataFolder + fmt.Sprintf("%s%d.json", blockInfoFilePrefix, all_blocks[i]), file, 0644)
	}
}

func batchWriteBlockInfo(folderPath string, blocks []uint) {
	eth2Client := newEth2Client()
	for i := 0; i < len(blocks); i++ {
		// every 10 req, wait 1 min
		if (i != 0 && i % 10 == 0){
			time.Sleep(60 * time.Second)
		}
		res := eth2Client.fetchBlockInfo(blocks[i])
		file, _ := json.MarshalIndent(res, "", " ")
		_ = ioutil.WriteFile(folderPath + fmt.Sprintf("%s%d.json", blockInfoFilePrefix, blocks[i]), file, 0644)
	}
}

func blockInfoExist(folderPath string, blockSlot uint) bool {
	// Check if the file exists
	_, err := os.Stat(folderPath + fmt.Sprintf("%s%d.json", blockInfoFilePrefix, blockSlot))
	if err != nil {
		if os.IsNotExist(err) {
			// The file does not exist
			return false
		}
		// Other error occurred
		panic(err)
	}

	// The file exists
	return true
}

func readBlockInfo(folderPath string, blockSlot uint) BlockAttestationsRes {
	file, _ := ioutil.ReadFile(folderPath + fmt.Sprintf("%s%d.json", blockInfoFilePrefix, blockSlot))
 
	data := BlockAttestationsRes{}
 
	_ = json.Unmarshal([]byte(file), &data)
	return data
}

// Should be deprecated
func readBlockInfo2(folderPath string, blockSlot uint) BlockAttestationsRes {
	file, _ := ioutil.ReadFile(folderPath + fmt.Sprintf("%s%d.json", blockInfoFilePrefix2, blockSlot))
 
	data := BlockAttestationsRes{}
 
	_ = json.Unmarshal([]byte(file), &data)
	return data
}