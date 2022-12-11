package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"fmt"
	// "log"
	"sync"

	// "io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/spf13/cobra"
)

var allValidators map[int]string
var mp  map[string]bool


func sample1() {
	fmt.Printf("sample1\n")
	var sec bls.SecretKey
	sec.SetByCSPRNG()
	msg := []byte("abc")
	pub := sec.GetPublicKey()
	fmt.Println(pub)
	sig := sec.SignByte(msg)
	fmt.Println(sig)
	fmt.Printf("verify=%v\n", sig.VerifyByte(pub, msg))
}

func sample2() {
	fmt.Printf("sample2\n")
	var sec bls.SecretKey
	sec.SetByCSPRNG()
	fmt.Printf("sec:%s\n", sec.SerializeToHexStr())
	pub := sec.GetPublicKey()
	fmt.Printf("1.pub:%s\n", pub.SerializeToHexStr())
	fmt.Printf("1.pub x=%x\n", pub)
	var P *bls.G1 = bls.CastFromPublicKey(pub)
	bls.G1Normalize(P, P)
	fmt.Printf("2.pub:%s\n", pub.SerializeToHexStr())
	fmt.Printf("2.pub x=%x\n", pub)
	fmt.Printf("P.X=%x\n", P.X.Serialize())
	fmt.Printf("P.Y=%x\n", P.Y.Serialize())
	fmt.Printf("P.Z=%x\n", P.Z.Serialize())
}

func sample3() {
	fmt.Printf("sample3\n")
	var sec bls.SecretKey
	b := make([]byte, 64)
	for i := 0; i < len(b); i++ {
		b[i] = 0xff
	}
	err := sec.SetLittleEndianMod(b)
	if err != nil {
		fmt.Printf("err")
		return
	}
	fmt.Printf("sec=%x\n", sec.Serialize())
}

func sample4() {
	fmt.Printf("sample4\n")
	var sec bls.SecretKey
	secByte, _ := hex.DecodeString("4aac41b5cb665b93e031faa751944b1f14d77cb17322403cba8df1d6e4541a4d")
	sec.Deserialize(secByte)
	fmt.Println(sec)
	msg := []byte("message to be signed.")
	fmt.Printf("sec:%x\n", sec.Serialize())
	pub := sec.GetPublicKey()
	fmt.Printf("pub:%x\n", pub.Serialize())
	sig := sec.SignByte(msg)
	fmt.Printf("sig:%x\n", sig.Serialize())
}

// this function return true if sigToVerify is aggregated signature generated by hexstring pk each independently sign the same underlying msg, otherwise return false
func aggregateVerify(msg []byte, rawPks []string, sigToVerify string) bool {
	sig := bls.Sign{}
	sig.DeserializeHexStr(sigToVerify)
	
	var pks []bls.PublicKey
	for _, rawPk := range rawPks {
		pk := bls.PublicKey{}
		pk.DeserializeHexStr(rawPk)
		pks = append(pks, pk)
	}

	return sig.FastAggregateVerify(pks, msg)
}


func verfiyAttestationByValidatorAndBlock(validator int, blockSlot int) []bool {
	block := readBlockInfo(blockInfoDataFolder, uint(blockSlot))
	res := []bool{}
	for _, attestation := range (block.Data) {
		if contains(attestation.Validators, validator) {
			// get data root
			signing_root := getSigningRoot2(attestation.Beaconblockroot[2:], attestation.Signature[2:], uint(attestation.Slot), uint(attestation.Committeeindex), uint(attestation.SourceEpoch), uint(attestation.TargetEpoch), attestation.SourceRoot[2:], attestation.TargetRoot[2:])
			// get sig to verify
			sig := attestation.Signature[2:]
			var pks []string
			for _, idx := range attestation.Validators {
				pks = append(pks, allValidators[idx][2:])
			}
			r := aggregateVerify(signing_root[:], pks, sig)
			res = append(res, r)
		}
	}
	
	return res
}

func verifyAttestation(beaconBlockRoot string, sig string, slot uint, committeeIdx uint, sourceEpoch uint, targetEpoch uint, sourceRoot string, targetRoot string, validators []int) bool {
	// get data root
	signing_root := getSigningRoot2(beaconBlockRoot[2:], sig[2:], slot, committeeIdx, sourceEpoch, targetEpoch, sourceRoot[2:],  targetRoot[2:])
	var pks []string
	for _, idx := range validators {
		pks = append(pks, allValidators[idx][2:])
	}
	r := aggregateVerify(signing_root[:], pks, sig[2:])
	return r
}

func verifyAllAttestationInBlock(blockSlot int) []bool {
	folderPath := "test/"
	if !blockInfoExist(folderPath, uint(blockSlot)) {
		batchWriteBlockInfo(folderPath, []uint{uint(blockSlot)})
	}
	block := readBlockInfo(folderPath, uint(blockSlot))
	res := []bool{}
	for _, attestation := range (block.Data) {
		// key := strconv.Itoa(attestation.Slot) + ":" + strconv.Itoa(attestation.Committeeindex)

		// get data root
		signing_root := getSigningRoot2(attestation.Beaconblockroot[2:], attestation.Signature[2:], uint(attestation.Slot), uint(attestation.Committeeindex), uint(attestation.SourceEpoch), uint(attestation.TargetEpoch), attestation.SourceRoot[2:], attestation.TargetRoot[2:])
		// get sig to verify
		sig := attestation.Signature[2:]
		var pks []string
		for _, idx := range attestation.Validators {
			pks = append(pks, allValidators[idx][2:])
		}
		r := aggregateVerify(signing_root[:], pks, sig)
		res = append(res, r)
	}
	
	return res
}

func searchSinglePubKey(validator int, blockSlot int) (int, string) {
	block := readBlockInfo(blockInfoDataFolder, uint(blockSlot))
	var idxRes int
	var pkRes string
	for _, attestation := range (block.Data) {
		if contains(attestation.Validators, validator) {
			// get data root
			signing_root := getSigningRoot2(attestation.Beaconblockroot[2:], attestation.Signature[2:], uint(attestation.Slot), uint(attestation.Committeeindex), uint(attestation.SourceEpoch), uint(attestation.TargetEpoch), attestation.SourceRoot[2:], attestation.TargetRoot[2:])
			// get sig to verify
			sig := attestation.Signature[2:]
			var wg sync.WaitGroup
			wg.Add(1)
			for idx, pubkey := range allValidators{
				go func(idx int, pubkey string) {
					var pks []string
					pks = append(pks, pubkey[2:])
					r := aggregateVerify(signing_root[:], pks, sig)
					if r {
						wg.Done()
						idxRes = idx
						pkRes = pubkey
					}
				}(idx, pubkey)
				

			}
			wg.Wait()
			break	
		}
	}
	return idxRes, pkRes

	
}

func checkExcel() {
	f, err := excelize.OpenFile("./data/unslashed_double_votes.xlsx")
	if err!=nil {
		fmt.Println(err)
	}
	eth2Client := newEth2Client()
	for i := 2; i < 200; i++ {
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
			_, pks, sig, err := eth2Client.GetAttestationsForBlock(uint(blockSlot), validator)
			if err != nil {
				fmt.Print(err)
			}else{
				fmt.Print(strconv.FormatBool(aggregateVerify(nil, pks, sig)) + "    ")
			}
			time.Sleep(1 * time.Second)
			
		}
		fmt.Println()
	}
}

func checkDoubleVoteWithLocalData(){
	f, err := excelize.OpenFile("./data/unslashed_double_votes.xlsx")
	if err!=nil {
		fmt.Println(err)
	}

	defer f.Close()
	
	// 405
	for i := 2; i < 406; i++ {
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
		// c, t, f:=0,0,0
		
		for _, blockSlot := range blocks {
			res := verfiyAttestationByValidatorAndBlock(validator, blockSlot)
			for _, r:= range res {
				fmt.Print(strconv.FormatBool(r)+"  ")
			}
		}
		fmt.Println()
	}
	// file, _ := json.MarshalIndent(mp, "", " ")
	// _ = ioutil.WriteFile("slot:committeeIdx.json", file, 0644)
	
}

func checkSurroundVoteWithLocalData(){
	f, err := os.Open("./data/surround_vote")
	if err!=nil {
		fmt.Println(err)
	}

	defer f.Close()

    scanner := bufio.NewScanner(f)
    // optionally, resize scanner's capacity for lines over 64K, see next example
	first := true
    for scanner.Scan() {
		if first {
			first = false
			continue
		}
		vals := strings.Fields(scanner.Text())
		idx, _ := strconv.Atoi(vals[0])
		blockslot, _ := strconv.Atoi(vals[1])
		res := verfiyAttestationByValidatorAndBlock(idx, blockslot)
        for _, r:= range res {
			fmt.Print(strconv.FormatBool(r)+"  ")
		}
		fmt.Println()
    }

    if err := scanner.Err(); err != nil {
        panic(err)
    }
}

func checkalgo(){
	f, err := excelize.OpenFile("./data/unslashed_double_votes.xlsx")
	if err!=nil {
		fmt.Println(err)
	}
	all_validators := getValidatorMap()
	both_true := 0
	both_false := 0
	true_false := 0
	count := 0

	// 405
	for i := 2; i < 406; i++ {
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
		c, t, f:=0,0,0
		
		for _, blockSlot := range blocks {
			res := readBlockInfo(blockInfoDataFolder, uint(blockSlot))
			for _, attestation := range (res.Data) {
				if !contains(attestation.Validators, validator) {
					// get data root
					// msg := attestation.Beaconblockroot
					// get pubkeys
					var pks []string
					for _, idx := range attestation.Validators {
						pks = append(pks, all_validators[idx][2:])
					}
					// get sig to verify
					sig := attestation.Signature
					check:=aggregateVerify(nil, pks, sig)
					// fmt.Print(strconv.FormatBool(check)+"  ")
					if check{
						t++
					}else{
						f++
					}
					c++
					
				}
			}
		}
		if c==2 {
			count++
			if t==0 {
				both_false++
			}else if t==2{
				both_true++
			}else{
				true_false++
			}
		}
		fmt.Println()
	}
	fmt.Println(both_true, both_false, true_false, count)
}

func removeDuplicateInt(intSlice []int) []int {
    allKeys := make(map[int]bool)
    list := []int{}
    for _, item := range intSlice {
        if _, value := allKeys[item]; !value {
            allKeys[item] = true
            list = append(list, item)
        }
    }
    return list
}

func verifyAllFromDrive() {
	start := time.Now()
	t()
	fmt.Println(time.Since(start))
}

// 4219 608065 included_in: 608067 
// 56119 608090
// 0x958391837758f8275e71bf34405d27a509ff1b7de4e7a53d87aa89dbb6800e0f100f57e9fd1a1c1b7f908c3860dfddb4
//25359  988961 included in: 988965
// 99842 988973
//0x857f7ca393da6f6fd816ef37b5cc7f3b33bdf23421625de8e5922399390af1de3596da010ea047cb452ee6eb766005cd
//42021 918914 918922
//98751 918940
// 0xb24928375d7ebb58e50544dc569f0a9e4aa67346ea3d7efa3d93995bc1e11c9ad656410e4d634c60a04e7197b064958d
//56833 918914 918922
// 98778 918925
// 0x8ef8a6aa04403f7a25ad8bbb0845c183be8bd83f9680bed6df396231ee94a2248de1af367a67e16683e69d82e63a8818
//57356  918914 918922
// 75274 918930
// 0x973ea7aa4fd07db36c3be5a075e81fde16172ceb8a207f65812eb08601b78f6bc9756a7af00275a109db02f2197efc4c

func main() {
	bls.Init(bls.BLS12_381)
	bls.SetETHmode(bls.EthModeDraft07)
	
	allValidators = getValidatorMap3()
	// mp = make(map[string]bool)

	// Create the root command
	rootCmd := &cobra.Command{
		Use: "ss",
	}

	// Define a flag for the verbosity level
	// verbosity := rootCmd.Flags().Int("verbosity", 0, "Verbosity level")

	// Define a command for printing the command line arguments
	argsCmd := &cobra.Command{
		Use: "args",
		Run: func(cmd *cobra.Command, args []string) {
			// Print the command line arguments
			fmt.Println("Arguments:")
			for i, arg := range args {
				fmt.Printf("  %d: %s\n", i, arg)
			}
		},
	}


	var blockId uint64
	verifyAllAttestationInBlockCmd := &cobra.Command {
		Use: "verify --blockId=<blockslot>",
		Short: "verifyAllAttestationInBlock --blockId=<blockslot>",
		Long: "verifyAllAttestationInBlock [blockid] if the block doesn't exist locally, fetch the block from beaconcha.in under test/ and verify",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(verifyAllAttestationInBlock(int(blockId)))
		},
	}
	verifyAllAttestationInBlockCmd.Flags().Uint64Var(&blockId, "blockId", 0, "Required argument")
	verifyAllAttestationInBlockCmd.MarkFlagRequired("blockId")

	// Note currently only suport file in blockinfo folder, change to fetch later
	var validatorIdx uint64
	verifyAttestationByValidatorAndBlockCmd := &cobra.Command {
		Use: "verify --valIdx=<validatorIdx> --blockId=<blockslot>",
		Short: "verifyAllAttestationInBlock --blockId=<blockslot>",
		Long: "verifyAllAttestationInBlock [blockid] if the block doesn't exist locally, fetch the block from beaconcha.in under test/ and verify",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(verfiyAttestationByValidatorAndBlock(int(validatorIdx), int(blockId)))
		},
	}
	verifyAttestationByValidatorAndBlockCmd.Flags().Uint64Var(&blockId, "blockId", 0, "Required argument")
	verifyAttestationByValidatorAndBlockCmd.Flags().Uint64Var(&validatorIdx, "valIdx", 0, "Required argument")
	verifyAttestationByValidatorAndBlockCmd.MarkFlagRequired("blockId")
	verifyAttestationByValidatorAndBlockCmd.MarkFlagRequired("valIdx")

	
	var typ string
	checkCmd := &cobra.Command {
		Use: "check --type=double|surround",
		Short: "check --type=double|surround",
		Long: "check for interested blocks in potential double vote / surround vote detected from beaconcha.in",
		Run: func(cmd *cobra.Command, args []string) {
			switch typ {
			case "double":
				checkDoubleVoteWithLocalData()
			case "surround":
				checkSurroundVoteWithLocalData()
			}

		},
	}
	checkCmd.Flags().StringVar(&typ, "blockId", "", "Required argument")
	checkCmd.MarkFlagRequired("typ")

	searchCmd := &cobra.Command {
		Use: "search --valIdx=<validatorIdx> --blockId=<blockId>",
		Short: "search --valIdx=<validatorIdx> --blockId=<blockId>",
		Long: "brute-force search for correct signer for sinlge validator case",
		Run: func(cmd *cobra.Command, args []string) {
			idx, pubkey := searchSinglePubKey(int(validatorIdx), int(blockId))
			fmt.Println(idx, pubkey)
		},
	}

	searchCmd.Flags().Uint64Var(&blockId, "blockId", 0, "Required argument")
	searchCmd.Flags().Uint64Var(&validatorIdx, "valIdx", 0, "Required argument")
	searchCmd.MarkFlagRequired("blockId")
	searchCmd.MarkFlagRequired("valIdx")




	rootCmd.AddCommand(argsCmd)
	rootCmd.AddCommand(verifyAllAttestationInBlockCmd)
	rootCmd.AddCommand(verifyAttestationByValidatorAndBlockCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(searchCmd)


	// Parse the command line flags and arguments
	rootCmd.Execute()
	
	// checkDoubleVoteWithLocalData()
	// a := []int{1041108, 1041109, 1041117}
	// for _, v := range a{
	// 	fmt.Println(verfiyAttestationByValidatorAndBlock(33096, v))
		
	// }
	// verfiyAttestationByValidatorAndBlock()
	// verifyAllAttestationInBlock(3)
	
	
	

	// searchSinglePubKey(88475, 608067)
	// for i := 608064; i<= 608086; i++{
	// 	if !contains(allBlocks, i){
	// 		fmt.Println(i)
	// 		verifyAllAttestationInBlock(i)
	// 	}
		
	// }
	// // verfiyAttestationByValidatorAndBlock(4219, 608067)
	// checkSurroundVoteWithLocalData()
	// file, _ := json.MarshalIndent(mp, "", " ")
	// _ = ioutil.WriteFile("total_slot:committeeIdx.json", file, 0644)
	// r := verfiyAttestationByValidatorAndBlock(4219, 608067)
	// fmt.Println(r)
	
	// for i := 608064; i <= 608095; i++{
	// 	attestations := readBlockInfo("epoch/", uint(i))
	// 	for _, attestation := range(attestations.Data){
	// 		// fmt.Println(attestation.Committeeindex==11)
	// 		if contains(attestation.Validators, 4219){
	// 			fmt.Println(attestation)
	// 		}
	// 	}
	// }

	// getValidatorMap2()
	
}