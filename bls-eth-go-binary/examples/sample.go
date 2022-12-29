package main

import (
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"

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
var mp map[string]bool

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

func verfiyAttestationByValidatorAndBlock(validator int, blockSlot int) ([]bool, []string) {
	block := readBlockInfo(blockInfoDataFolder, uint(blockSlot))
	res := []bool{}
	sigs := []string{}
	for _, attestation := range block.Data {
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
			sigs = append(sigs, attestation.Signature)
		}
	}

	return res, sigs
}

func verifyAttestation(beaconBlockRoot string, sig string, slot uint, committeeIdx uint, sourceEpoch uint, targetEpoch uint, sourceRoot string, targetRoot string, validators []int) bool {
	// get data root
	signing_root := getSigningRoot2(beaconBlockRoot[2:], sig[2:], slot, committeeIdx, sourceEpoch, targetEpoch, sourceRoot[2:], targetRoot[2:])
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
	for _, attestation := range block.Data {
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
	for _, attestation := range block.Data {
		if contains(attestation.Validators, validator) {
			// get data root
			signing_root := getSigningRoot2(attestation.Beaconblockroot[2:], attestation.Signature[2:], uint(attestation.Slot), uint(attestation.Committeeindex), uint(attestation.SourceEpoch), uint(attestation.TargetEpoch), attestation.SourceRoot[2:], attestation.TargetRoot[2:])
			// get sig to verify
			sig := attestation.Signature[2:]
			var wg sync.WaitGroup
			wg.Add(1)
			found := false
			for idx, pubkey := range allValidators {
				go func(idx int, pubkey string) {
					if found {
						return
					}
					var pks []string
					pks = append(pks, pubkey[2:])
					r := aggregateVerify(signing_root[:], pks, sig)
					if r {
						wg.Done()
						idxRes = idx
						pkRes = pubkey
						fmt.Println(sig, idx, pubkey)
						found = true
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
	if err != nil {
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
			} else {
				fmt.Print(strconv.FormatBool(aggregateVerify(nil, pks, sig)) + "    ")
			}
			time.Sleep(1 * time.Second)

		}
		fmt.Println()
	}
}

var c map[string]bool

func checkDoubleVoteWithLocalData() {
	f, err := excelize.OpenFile("./data/unslashed_double_votes.xlsx")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("dd")

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
			res, sigs := verfiyAttestationByValidatorAndBlock(validator, blockSlot)
			for i := 0; i < len(res); i++ {
				if !res[i] {
					c[sigs[i]] = true
				}
				fmt.Print(strconv.FormatBool(res[i]) + "  ")
			}

		}
		fmt.Println()
	}
	fmt.Println(len(c))
	// file, _ := json.MarshalIndent(mp, "", " ")
	// _ = ioutil.WriteFile("slot:committeeIdx.json", file, 0644)

}

// func key(committeeIdx int, slot int) string {
// 	strconv.FormatInt(slot, 10) + ":" + strconv
// }

func checkSurroundVoteWithLocalData() {
	f, err := os.Open("./data/surround_vote")
	if err != nil {
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
		res, sigs := verfiyAttestationByValidatorAndBlock(idx, blockslot)
		for i := 0; i < len(res); i++ {
			if !res[i] {
				c[sigs[i]] = true
			}
			fmt.Print(strconv.FormatBool(res[i]) + "  ")
		}
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println(len(c))
}

func checkalgo() {
	f, err := excelize.OpenFile("./data/unslashed_double_votes.xlsx")
	if err != nil {
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
		c, t, f := 0, 0, 0

		for _, blockSlot := range blocks {
			res := readBlockInfo(blockInfoDataFolder, uint(blockSlot))
			for _, attestation := range res.Data {
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
					check := aggregateVerify(nil, pks, sig)
					// fmt.Print(strconv.FormatBool(check)+"  ")
					if check {
						t++
					} else {
						f++
					}
					c++

				}
			}
		}
		if c == 2 {
			count++
			if t == 0 {
				both_false++
			} else if t == 2 {
				both_true++
			} else {
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

func allFalseAttestations() {
	items, _ := ioutil.ReadDir("verificationResult")
	c := 0
	f, _ := os.Create("allFalse.txt")
	defer f.Close()

	for _, item := range items {
		fileLoc := "verificationResult/" + item.Name()
		readFile, err := os.Open(fileLoc)

		if err != nil {
			fmt.Println(err)
		}
		fileScanner := bufio.NewScanner(readFile)

		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			c++
			// fmt.Println(fileScanner.Text())
			f.WriteString(fileScanner.Text())
			f.WriteString("\n")
		}

		readFile.Close()

	}
	fmt.Println(c)
}

func fetchepochs(interestedBlock int) {
	start := time.Now()
	blocks := []uint{}
	startBlock := (interestedBlock / 32) * 32
	for block := startBlock; block < startBlock+32; block++ {
		blocks = append(blocks, uint(block))
	}
	batchWriteBlockInfo("epochs/", blocks)
	fmt.Println(time.Since(start))

}

type Schedule struct {
	Data []struct {
		Index      string   `json:"index"`
		Slot       string   `json:"slot"`
		Validators []string `json:"validators"`
	} `json:"data"`
	ExecutionOptimistic bool `json:"execution_optimistic"`
}

func unvotedValidators(interestedBlock int) []int {
	// epoch's blocks under epochs/, committee schedule under schedule/
	file, _ := ioutil.ReadFile(fmt.Sprintf("schedule/%d.json", interestedBlock))
	var schedule Schedule
	_ = json.Unmarshal([]byte(file), &schedule)
	allValidators := map[int]bool{}
	for _, val := range schedule.Data {
		for _, validator := range val.Validators {
			idx, _ := strconv.Atoi(validator)
			if _, exist := allValidators[idx]; !exist {
				allValidators[idx] = true
			}
		}
	}
	// based on block_slot search the schedule
	// unique validators in that epoch
	votedValidators := map[int]bool{}
	startBlock := (interestedBlock / 32) * 32
	for block := startBlock; block < startBlock+32; block++ {
		atts := readBlockInfo("epochs/", uint(block))
		for _, val := range atts.Data {
			for _, idx := range val.Validators {
				if idx == 21955 {
					fmt.Println(val)
					os.Exit(0)
				}
				if _, exist := votedValidators[idx]; !exist {
					votedValidators[idx] = true
				}
			}
		}
	}

	unvotedValidators := []int{}
	for k := range allValidators {
		if _, voted := votedValidators[k]; !voted {
			unvotedValidators = append(unvotedValidators, k)
		}
	}

	fmt.Println(len(allValidators), len(votedValidators))
	fmt.Println(len(unvotedValidators))
	fmt.Println(contains(unvotedValidators, 56119))
	return unvotedValidators

	// search all voted validators in that epoch by traverse files epoch/

}

func smartSearchPubkeys(interestedBlock int) {
	searchSpace := unvotedValidators(interestedBlock)

	// traverse all false att, correct each by brute force search space
	readFile, _ := os.Open("allFalse.txt")
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	f, _ := os.Create(fmt.Sprintf("correction/%d.txt", interestedBlock))
	defer f.Close()

	for fileScanner.Scan() {
		fa := fileScanner.Text()
		row := strings.Split(fa, ",")
		blockslot, _ := strconv.ParseUint(row[4], 10, 64)
		if blockslot != uint64(interestedBlock) {
			continue
		}

		bbr := row[1]
		sig := row[6]
		slot, _ := strconv.ParseUint(row[7], 10, 64)

		cidx, _ := strconv.ParseUint(row[5], 10, 64)
		sourceEpoch, _ := strconv.ParseUint(row[8], 10, 64)
		sourceRoot := row[9]
		targetEpoch, _ := strconv.ParseUint(row[10], 10, 64)
		targetRoot := row[11]
		if row[12][0] == '"' {
			row[12] = strings.Trim(row[12], "\"\"")
		}
		validators := strtouints(strings.Trim(row[12], "\\[\\]"))
		// fmt.Println(validators)

		// get data root
		signing_root := getSigningRoot2(bbr[2:], sig[2:], uint(slot), uint(cidx), uint(sourceEpoch), uint(targetEpoch), sourceRoot[2:], targetRoot[2:])
		// fmt.Println(signing_root)
		// get sig to verify
		sig = sig[2:]

		if len(validators) == 1 {
			// searchSinglePubKey(validators[0], interestedBlock)
			for _, valIdx := range searchSpace {
				pubkey := allValidators[valIdx]
				var pks []string
				pks = append(pks, pubkey[2:])
				r := aggregateVerify(signing_root[:], pks, sig)
				if r {
					fmt.Println(pubkey)
					f.WriteString(fmt.Sprintf("%s, %d, %s", sig, valIdx, pubkey))

				}

			}
		}
	}

	readFile.Close()

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
	c = map[string]bool{}
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
	verifyAllAttestationInBlockCmd := &cobra.Command{
		Use:   "verifyAll --blockId=<blockslot>",
		Short: "verify all attestations in a specified block",
		Long:  "verifyAllAttestationInBlock [blockid] if the block doesn't exist locally, fetch the block from beaconcha.in under test/ and verify",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(verifyAllAttestationInBlock(int(blockId)))
		},
	}
	verifyAllAttestationInBlockCmd.Flags().Uint64Var(&blockId, "blockId", 0, "Required argument")
	verifyAllAttestationInBlockCmd.MarkFlagRequired("blockId")

	// Note currently only suport file in blockinfo folder, change to fetch later
	var validatorIdx uint64
	verifyAttestationByValidatorAndBlockCmd := &cobra.Command{
		Use:   "verify --valIdx=<validatorIdx> --blockId=<blockslot>",
		Short: "verify attestations that contains validatorIdx in a specified block",
		Long:  "verfiyAttestationByValidatorAndBlock [blockid] verify att that involves [valIdx]",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(verfiyAttestationByValidatorAndBlock(int(validatorIdx), int(blockId)))
		},
	}
	verifyAttestationByValidatorAndBlockCmd.Flags().Uint64Var(&blockId, "blockId", 0, "Required argument")
	verifyAttestationByValidatorAndBlockCmd.Flags().Uint64Var(&validatorIdx, "valIdx", 0, "Required argument")
	verifyAttestationByValidatorAndBlockCmd.MarkFlagRequired("blockId")
	verifyAttestationByValidatorAndBlockCmd.MarkFlagRequired("valIdx")

	var typ string
	checkCmd := &cobra.Command{
		Use:   "check --type=double|surround",
		Short: "check for all interested attestations in double / surround vote detected from beaconcha.in",
		Long:  "check for interested blocks in potential double vote / surround vote detected from beaconcha.in",
		Run: func(cmd *cobra.Command, args []string) {
			switch typ {
			case "double":
				checkDoubleVoteWithLocalData()
			case "surround":
				checkSurroundVoteWithLocalData()
			case "both":
				checkDoubleVoteWithLocalData()
				checkSurroundVoteWithLocalData()
			}
		},
	}
	checkCmd.Flags().StringVar(&typ, "type", "", "Required argument")
	checkCmd.MarkFlagRequired("typ")

	searchCmd := &cobra.Command{
		Use:   "search --valIdx=<validatorIdx> --blockId=<blockId>",
		Short: "brute-force search for correct signer for sinlge validator case",
		Long:  "brute-force search for correct signer for sinlge validator case",
		Run: func(cmd *cobra.Command, args []string) {
			idx, pubkey := searchSinglePubKey(int(validatorIdx), int(blockId))
			fmt.Println(idx, pubkey)
		},
	}
	searchCmd.Flags().Uint64Var(&blockId, "blockId", 0, "Required argument")
	searchCmd.Flags().Uint64Var(&validatorIdx, "valIdx", 0, "Required argument")
	searchCmd.MarkFlagRequired("blockId")
	searchCmd.MarkFlagRequired("valIdx")

	verifyAllCmd := &cobra.Command{
		Use:   "verifyDrive",
		Short: "fetch from google drive and verify all atts for all blocks, output failed verification to verificationResult/",
		Long:  "fetch from google drive and verify all atts",
		Run: func(cmd *cobra.Command, args []string) {
			verifyAllFromDrive()
		},
	}

	allFalseAttestationCmd := &cobra.Command{
		Use:   "allFalse",
		Short: "fetch from google drive and verify all atts for all blocks, output failed verification to verificationResult/",
		Long:  "fetch from google drive and verify all atts",
		Run: func(cmd *cobra.Command, args []string) {
			allFalseAttestations()
		},
	}

	fetchEpochByInterestedBlockCmd := &cobra.Command{
		Use:   "fetch epoch for interested block",
		Short: "fetch epoch for interested block under epochs/",
		Long:  "fetch epoch for interested block under epochs/",
		Run: func(cmd *cobra.Command, args []string) {
			fetchepochs(608067)
		},
	}

	unvotedValidatorsCmd := &cobra.Command{
		Use:   "unvote validators in epoch defined by interested block",
		Short: "unvote validators in epoch defined by interested block",
		Long:  "unvote validators in epoch defined by interested block",
		Run: func(cmd *cobra.Command, args []string) {
			unvotedValidators(608067)
		},
	}
	smartSearchCmd := &cobra.Command{
		Use:   "ss smart search pubkey for correct signing validator sets",
		Short: "ss smart search pubkey for correct signing validator sets",
		Long:  "ss smart search pubkey for correct signing validator sets",
		Run: func(cmd *cobra.Command, args []string) {
			smartSearchPubkeys(608067)
		},
	}
	signingRootCmd := &cobra.Command{
		Use:   "sr validators in epoch defined by interested block",
		Short: "sr validators in epoch defined by interested block",
		Long:  "sr validators in epoch defined by interested block",
		Run: func(cmd *cobra.Command, args []string) {
			getSigningRoot()
		},
	}
	bfPubkeyCmd := &cobra.Command{
		Use:   "bf brufeforce search pubkey for each 1-val case",
		Short: "bf brufeforce search pubkey for each 1-val case",
		Long:  "bf brufeforce search pubkey for each 1-val case",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			bruteforceSearch()
			fmt.Println(time.Since(start))

		},
	}

	duplicateVoteCmd := &cobra.Command{
		Use:   "dv duplicated vote for same target epoch with diff data",
		Short: "dv duplicated vote for same target epoch with diff data",
		Long:  "dv duplicated vote for same target epoch with diff data",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			searchDuplicateVote()
			fmt.Println(time.Since(start))

		},
	}

	correctionCmd := &cobra.Command{
		Use:   "show correction map",
		Short: "show correction map",
		Long:  "show correction map",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			fmt.Println(readCorrection())
			fmt.Println(time.Since(start))

		},
	}

	verifyDVCmd := &cobra.Command{
		Use:   "vdv",
		Short: "verify duplicate votes",
		Long:  "verify duplicate votes",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			verifyDv()
			fmt.Println(time.Since(start))
		},
	}

	shuffleCmd := &cobra.Command{
		Use:   "shuffle",
		Short: "shuffle duplicate votes",
		Long:  "shuffle duplicate votes",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			// var seed [32]byte
			// bytes, _ := hex.DecodeString( "2c054da59da41371a75148181be79b97671e2ed4355fb0c524aaa238d7e42db5")
			// copy(seed[:], bytes)

			// b := []byte("aef8ab944f63cced7f0f2d2dbe7e8eee3a4f84cbc2e257fb5dcf053e7f61b828dbca786e24b118119dc76d4040ce0ee30f4247305229a427893b810ffccf1d38b37e80e925fb37452184a886f5fed04974e7bdd5065951816a92e0618dcae471")

			// c := []byte(
			// 	"b0ea7c45822ff047f04630e3611656f146ec7f309f6199a37c609bcb7b836554f1235b3f9c2217220d6012c77a983f7018607b052002d2d1f863b7a3b0829bdff0156eb26973a0a80b9bd0d0e4c41e69d30672ca2611278ca12bfd904bee9d88")
			// d := []byte(
			// 	"a808cd9950ad09752e961307ad93ca0cfcef95bd69924f289a2dec204f7ee6a5117d34386e3e8dcfae599fcba8a6b04511e5326350841a0bb5682b5e3991b5c3ec0830327c58c5836b603079e71c31c1a3cbf0a938a64fd1b0358e79bf06236c")
			// e := []byte(
			// 	"99f3d02265f4c8b0132f4dc15b1c81d28709257225495027cb396835b1b376565d69bd96b89e44019087d46f30539a9c18d5069fd963af05b2d3c976b31a98592eabbebbbf4c12cd7e6bee60253a7d597b877f8bc04cf73d4792e6f73a066823")
			// f := []byte(
			// 	"b32db77c1ead597f86eae75f603c30f0b39bc7d48e35940217149782b43f0d29987d66301a91702cb443c0fe8c35174218946336844eae3f3cb0053ab1b136be1b751c14823860abac03bbb5dc11a7726c90b946d259fc90e82910575330b7d2")

			// for i := 0; i < 32; i++ {
			// 	seed[i] = seed[i] ^ b[i]
			// }
			// for i := 0; i < 32; i++ {
			// 	seed[i] = seed[i] ^ c[i]
			// }
			// for i := 0; i < 32; i++ {
			// 	seed[i] = seed[i] ^ d[i]
			// }
			// for i := 0; i < 32; i++ {
			// 	seed[i] = seed[i] ^ e[i]
			// }
			// for i := 0; i < 32; i++ {
			// 	seed[i] = seed[i] ^ f[i]
			// }
			// fmt.Println(seed)
			// 0xc98428d231efde4e344843712fe0d328c1b862f0b6b40ab985ef8cc2e825842b
			// lower := ((608090-608064+1)*23 + 11) * 131
			// upper := ((608090-608064+1)*23 + 11) * 132
			// fmt.Println(lower, upper)
			// minidx := 0
			// mini := 0
			// min := 100000000000000000

			f, err := os.Open("randao.csv")
			if err != nil {

			}
			defer f.Close()

			csvReader := csv.NewReader(f)
			records, err := csvReader.ReadAll()

			for i := 20000; i <= 22000; i++ {
				for j := 1; j < len(records); j++ {
					row := records[j]
					if len(row[1]) < 2 {
						continue
					}
					seedstr := row[1][2:]
					// fmt.Println(seedstr)
					bytes, _ := hex.DecodeString(seedstr)

					var seed [32]byte
					copy(seed[:], bytes)
					idx := getIndex(131, uint64(i), seed)
					if idx == 0 {
						fmt.Println(i, j, idx)
					}
				}
				// idx := reverseIndex(0, uint64(i), seed)
				// id := getIndex(5596, uint64(i), seed)
				// fmt.Println(idx)
				// if idx >= uint64(lower) {
				// 	diff := int(idx) - lower
				// 	if diff < min {
				// 		min = diff
				// 		minidx = int(idx)
				// 		mini = i
				// 	}
				// }
				// if idx == 80113 {
				// 	fmt.Println(idx, i)
				// }
				// fmt.Println(idx, id)
				// if  idx >= uint64(lower) && idx <= uint64(upper) {

				// }
			}
			// getIndex(100,50,seed)
			// fmt.Println(minidx, mini)

			fmt.Println(time.Since(start))
		},
	}

	nocorrectionDvCmd := &cobra.Command{
		Use:   "ndv",
		Short: "ndv duplicate votes",
		Long:  "ndv duplicate votes",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			searchDuplicateVoteWithoutCorrection()
			fmt.Println(time.Since(start))
		},
	}

	committeeScheduleCmd := &cobra.Command{
		Use:   "csc",
		Short: "csc duplicate votes",
		Long:  "csc duplicate votes",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			// getCommitteeSchedule(608065, 2)
			fmt.Println(time.Since(start))
		},
	}

	specializedCmd := &cobra.Command{
		Use:   "f",
		Short: "f duplicate votes",
		Long:  "f duplicate votes",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			// put your stuff here
			// getFirstIndex("", )
			fmt.Println(time.Since(start))
		},
	}

	recoverCmd := &cobra.Command{
		Use:   "recover",
		Short: "recover duplicate votes",
		Long:  "recover duplicate votes",
		Run: func(cmd *cobra.Command, args []string) {
			start := time.Now()
			recover()
			fmt.Println(time.Since(start))
		},
	}
	// [25838 82995 62200 14972 29786 4562 54057 24424 11095 4306 67871 65557 77003 51987 27434 64025 31467 29367 5635 96106 66679 18607 55579 11322 38818 16880 26324 82677 50214 17402 4675 18507 72646 43328 37267 68087 15316 47509 53837 70595 31500 36641 47611 74666 67516 50158 72865 71930 8439 57120 23306 62075 68426 77238 39762 74187 49158 87238 25138 91680 91463 45001 52639 42813 83778 378 70168 48546 56119 3282 25280 94017 36392 15024 9401 80150 44706 6148 57694 92182 83944 69714 1385 63150 72470 14075 90758 22935 73161 8517 92735 10346 52401 14525 51551 89784 48244 83176 76113 78182 44429 10512 60095 17832 52523 61570 17834 5144 32874 48660 77622 52914 63642 30141 33353 45075 30686 9306 7813 81368 41411 26909 17633 27967 80442 10710 81238 17307 57107 78546 74953 39480]
	// [34175 71854 15777 43044 83060 68488 74930 6866 30612 22209 8203 38744 32006 6912 31483 76150 78495 72680 6207 52072 3438 52970 70005 50055 9219 49034 79536 3287 37151 69549 9629 10487 80366 49597 74729 72369 1928 58422 62686 10015 14855 45584 3812 44536 7086 65008 54080 35452 66333 88545 3640 16838 26948 18648 73315 96711 76089 40684 59195 82214 60912 85479 41853 34251 13415 92975 1803 51066 70187 55434 78032 24796 22988 45349 24761 37273 55862 15586 51153 93790 29410 58326 65794 63952 34262 46855 76112 38487 73618 92635 69464 72965 84040 4503 52880 801 22414 56797 14229 77431 10967 8636 17392 96585 81022 11332 45034 18700 19963 78320 79005 17647 96857 30491 75924 4755 11631 89925 78491 31324 58005 25528 92912 72759 14422 5282 46414 16399 15734 55847 48083]
	// [41031 74135 17027 92871 2287 18220 3587 40119 52887 39619 11495 15072 24714 61426 30083 42836 7818 18021 45388 23525 28553 12244 33570 25506 69711 1715 77863 67409 60550 57159 74089 4298 26181 50728 93409 50865 2789 96288 11560 22575 73460 30993 83755 92868 73411 34631 51333 47688 27123 51654 48658 30975 49324 54691 13896 41899 95559 1387 2888 40714 71821 1297 64024 90685 9863 27063 4581 72056 32003 3452 59225 416 94605 79015 48776 44613 76264 62060 50971 89324 23952 66746 13727 40452 64424 85363 75754 44971 14953 68122 32291 94646 36046 88232 10545 78519 9924 49036 51119 71247 77776 80152 44189 14638 23318 42216 45490 51760 51649 38857 45592 92794 19228 47195 21947 55958 77008 61939 18850 59694 56974 19442 42578 79160 26970 32704 12395 41786 35098 38640 77717]
	// [1055 6526 14825 10490 15565 5824 19244 3786 7205 14484 17109 5285 3212 20400 16842 6759 19882 16292 16345 19679 4118 7834 20934 3952 17225 5671 8651 6303 3034 502 662 19667 11116 19211 2594 17419 12164 11703 14754 6858 19678 8847 12591 11728 12835 11050 10042 1976 18290 11721 11690 16311 9494 4633 2622 2949 9060 8096 15304 12867 4031 7322 11942 374 6304 6570 20568 17254 15282 18012 5108 10436 4913 16381 20629 15391 7825 8100 20757 1221 10285 4120 3745 2404 2920 2634 11317 1583 16491 8912 4370 18696 312 7068 4026 15590 8730 17976 5903 3008 14147 3980 11605 9907 13025 19989 14245 9392 2031 16053 868 84 18827 6410 16146 16191 11214 13031 20253 8557 8280 3228 12525 11632 20134 16190 16883 16827 11972 16748 7649 11221]

	// [16264 20662 17562 5586 16174 5826 19247 21042 657 5790 19843 814 12353 13179 20446 11397 11437 15998 10434 15218 916 9782 3461 6119 5358 10617 12746 1945 3990 15788 20878 13370 13381 14734 14840 5081 13144 1589 18544 19262 5653 11596 17906 21054 94 12421 15840 3782 574 11367 19403 4541 223 9898 1015 19332 3233 12046 15915 1653 5090 10180 6961 6564 467 161 13666 2442 16009 15084 18103 19779 19426 3560 7790 6287 18088 313 696 7695 9837 4645 5828 13884 11895 3409 15383 7394 925 8693 11297 7643 6713 7297 19027 11978 5956 3641 14892 10826 17206 16130 17257 3013 20262 20850 5476 19336 2433 8818 6806 1107 1484 21001 14568 14967 16796 10981 940 12119 4852 7298 2592 15494 13172 12914 6802 3661 16861 1977 9297 17141]
	// [4973 7331 93582 234 84360 20696 85474 22450 86128 71606 83346 69210 5840 30415 84927 9713 2783 31455 18682 90196 62996 45959 84891 32147 62416 62441 14845 2818 30982 47362 44946 82705 45109 42340 84242 56964 85187 48930 40601 12142 77112 48700 16888 54993 25424 311 2277 18395 20809 12885 84227 43855 36159 31525 61004 40506 29405 77394 21962 51794 53549 12699 73421 73145 83826 33629 27945 45491 31904 34978 17420 38049 28428 51237 21259 68381 27996 81638 6675 44150 71650 36622 60231 51316 84038 15393 88386 52733 23370 79457 48019 67597 46817 43415 70060 48326 63291 27459 41318 75980 24741 7750 22661 40094 50277 7805 47611 11052 91118 58955 24939 63753 6171 46699 35551 11349 64808 46406 91536 57562 83318 8308 85876 35851 34842 48743 55059 81619 57520 62269 39317 96500]
	// [17115 71133 46500 58690 56052 46498 7958 66058 55889 1757 79031 8440 60981 46747 80754 94668 67367 83130 70781 37620 38348 7852 59783 49955 93640 14556 65735 82961 71530 45456 83879 24982 83390 23284 60569 74721 77775 8480 17033 12011 21880 7579 15210 54662 80570 56053 38230 38981 65975 58261 17755 81100 76577 37140 7150 450 31043 21321 72350 70253 2252 79210 96159 50946 47635 39984 88026 92589 59030 10716 703 68821 21582 30895 38853 89314 47388 11188 39217 86899 52049 13791 50033 64025 22754 50817 34756 64991 17106 482 58662 28294 45495 72806 25765 40539 83044 8030 3345 32788 34273 53592 94492 12453 59575 24530 50073 59137 15009 27777 85648 40541 44521 12926 44318 54000 40237 66874 77246 3730 35411 8138 45392 95558 22767 67906 63184 39632 95836 31436 34725 42246]

	rootCmd.AddCommand(argsCmd)
	rootCmd.AddCommand(verifyAllAttestationInBlockCmd)
	rootCmd.AddCommand(verifyAttestationByValidatorAndBlockCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(verifyAllCmd)
	rootCmd.AddCommand(allFalseAttestationCmd)
	rootCmd.AddCommand(fetchEpochByInterestedBlockCmd)
	rootCmd.AddCommand(unvotedValidatorsCmd)
	rootCmd.AddCommand(smartSearchCmd)
	rootCmd.AddCommand(signingRootCmd)
	rootCmd.AddCommand(bfPubkeyCmd)
	rootCmd.AddCommand(duplicateVoteCmd)
	rootCmd.AddCommand(correctionCmd)
	rootCmd.AddCommand(verifyDVCmd)
	rootCmd.AddCommand(shuffleCmd)
	rootCmd.AddCommand(nocorrectionDvCmd)
	rootCmd.AddCommand(committeeScheduleCmd)
	rootCmd.AddCommand(specializedCmd)
	rootCmd.AddCommand(recoverCmd)


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
