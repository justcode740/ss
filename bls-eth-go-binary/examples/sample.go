package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/herumi/bls-eth-go-binary/bls"
)

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
	msg := []byte("message to be signed.")
	fmt.Printf("sec:%x\n", sec.Serialize())
	pub := sec.GetPublicKey()
	fmt.Printf("pub:%x\n", pub.Serialize())
	sig := sec.SignByte(msg)
	fmt.Printf("sig:%x\n", sig.Serialize())
}

// this function return true if sigToVerify is aggregated signature generated by hexstring pk each independently sign the same underlying msg, otherwise return false
func aggregateVerify(msg []byte, rawPks []string, sigToVerify string) bool {
	// fmt.Println(msg)
	// fmt.Println(rawPks)
	// fmt.Println(sigToVerify)
	
	// sig := bls.Sign{}
	// sigByte, _ := hex.DecodeString(sigToVerify)
	// sig.Deserialize(sigByte)
	
	// var pks []bls.PublicKey
	// for _, rawPk := range rawPks {
	// 	pk := bls.PublicKey{}
	// 	pkByte, _ := hex.DecodeString(rawPk)
	// 	pk.Deserialize(pkByte)
	// 	pks = append(pks, pk)
	// }
	// // fmt.Println(pks[0])
	// // fmt.Println(sig)

	// return sig.FastAggregateVerify(pks, msg)
	var sig bls.Sign
	err := sig.DeserializeHexStr(sigToVerify)
	if err != nil {
		error.Error(err)
	}
	
	var pks []bls.PublicKey
	for _, rawPk := range rawPks {
		pk := &bls.PublicKey{}
		pkByte, _ := hex.DecodeString(rawPk)
		pk.Deserialize(pkByte)
		pks = append(pks, *pk)
	}
	return sig.FastAggregateVerify(pks, msg)
}


func verfiyAttestationByValidatorAndBlock(validatorIdx uint, blockSlot uint) {
	
}

// func testClient(){
// 	eth2Client := newEth2Client()
// 	root_data, pks, sig, _ := eth2Client.GetAttestationsForBlock(1041108, 90250)
	
// 	fmt.Println(aggregateVerify(root_data, pks, sig))
// 	// eth2Client.GetValidatorPubKey([])
// }

func check_excel() {
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

const validatorIdsLoc = "validatorinfo/validatorIds.json"
func find_all_validators_of_interest(){
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
			res := readBlockInfo(uint(blockSlot))
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
	fmt.Println(len(res))
	ioutil.WriteFile("valtest50/validators", bytes, 0644)
}

func check_excel_with_local_data(){
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
			res := readBlockInfo(uint(blockSlot))
			for _, attestation := range (res.Data) {
				if contains(attestation.Validators, validator) {
					// fmt.Print(len(attestation.Validators))
					// get data root
					signing_root := getSigningRoot2(attestation.Beaconblockroot[2:], attestation.Signature[2:], uint(attestation.Slot), uint(attestation.Committeeindex), uint(attestation.SourceEpoch), uint(attestation.TargetEpoch))
					// data, _ := hex.DecodeString(attestation.Beaconblockroot[2:])
					// fmt.Print(msg)
					// get pubkeys
					var pks []string
					for _, idx := range attestation.Validators {
						// fmt.Println(idx, all_validators[idx], all_validators[idx]])
						
						pks = append(pks, all_validators[idx])
					}
					// get sig to verify
					sig := attestation.Signature
					check:=aggregateVerify(signing_root[:], pks, sig)

					// fmt.Print(" " + strconv.FormatBool(check)+"  ")
					if check{
						if len(attestation.Validators) < 3 {
							fmt.Println("slot number", attestation.Slot, blockSlot, validator)
						}
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
			res := readBlockInfo(uint(blockSlot))
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
					fmt.Print(strconv.FormatBool(check)+"  ")
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

func main() {
	bls.Init(bls.BLS12_381)
	bls.SetETHmode(bls.EthModeDraft07)
	// data_root := "0x1895bac33c8cecdf1c07a21fa0c1be683283a1c99f5964ed7bfeccc33104ecba"
	// pks := []string{
	// 	"0x80468c2579b577e50b1e69755f06e72ebf6ae9caaed311882236f843567f07edfead2fa0f448861b09fe7685910bc241",
	// 	"0x9131c281f29c3abf34bdcd3a344bb22fa5b07291fbae97ba189be9242e1fdb0474a2ed17a98a87fecd4427714a53a8da",
	// 	"0xa6eb4ebdf9b217e2db544dfa205a90ab27b71269149368854ba61d7c52e39fe3d9c47529468fa1c00d8d85b10057df0a",
	// }
	// pk1 := "0x930eb7c8dd107d2e1c444b17863c3209dc153af4ed4f3222f9c8be3608c772bc94e5ceab0b3940feefa79a13c87d5896"
	// // pk2 := "a323ec9163f8564020c15edf442c95240342fcaaf9cc06d9e3367c63e85af6173727b8b0bacc689acd41859301f70bcd"
	// // pk3 := "0x81690c370330ac17a48cdcfae58c89ab2061f7222409896d06134f4d89f896d2d375f567a541f11de9d2a6966ae256e7"
	// // getSigningRoot2("")
	// data, _ := hex.DecodeString("397f63851a66cd9bee5291a86356a11a4f0f314316ebd10c9b5455b28b2783ff")
	// signing_root := getSigningRoot2("397f63851a66cd9bee5291a86356a11a4f0f314316ebd10c9b5455b28b2783ff",
	// "a12cb3d04c4cd05b869a2d9cef5e11115348bb1e27fec217bcdf26fb42c74c69ec5f595c9692d4d2f8858a4f0ecd310a0e71c528bff539945b83693a8fb371f86725bd7294fc3f29296adcbf9d63b08952c148862fdceaf0065d72681b072da8",
	// 918914,
	// 23,
	// 28715,
	// 28716)
	// fmt.Println(len(data[:]))
	// fmt.Println(len(signing_root[:]))
	// pks := []string{pk1}
	// sig := "0xa12cb3d04c4cd05b869a2d9cef5e11115348bb1e27fec217bcdf26fb42c74c69ec5f595c9692d4d2f8858a4f0ecd310a0e71c528bff539945b83693a8fb371f86725bd7294fc3f29296adcbf9d63b08952c148862fdceaf0065d72681b072da8"
	// fmt.Println(aggregateVerify(data, pks, sig))
	// fmt.Println(aggregateVerify(signing_root[:], pks, sig))
	// t := []byte{97 ,54, 190}
	// fmt.Println(aggregateVerify(t, pks, sig))
	testLib2()


	// // c:=newEth2Client()
	// // c.check()
	// sign_root := getSigningRoot()
	// data, _:=hex.DecodeString("1895bac33c8cecdf1c07a21fa0c1be683283a1c99f5964ed7bfeccc33104ecba")
	// fmt.Println(aggregateVerify(sign_root[:], pks, sig))

	// {
	// 	"aggregationbits": "0x00000000000000000020",
	// 	"beaconblockroot": "0x397f63851a66cd9bee5291a86356a11a4f0f314316ebd10c9b5455b28b2783ff",
	// 	"block_index": 37,
	// 	"block_root": "0xa9c6d860db41e2e46c04c094e7d25efd7edab8bcef9572bb9a30e45cba1d3072",
	// 	"block_slot": 918938,
	// 	"committeeindex": 23,
	// 	"signature": "0xa12cb3d04c4cd05b869a2d9cef5e11115348bb1e27fec217bcdf26fb42c74c69ec5f595c9692d4d2f8858a4f0ecd310a0e71c528bff539945b83693a8fb371f86725bd7294fc3f29296adcbf9d63b08952c148862fdceaf0065d72681b072da8",
	// 	"slot": 918914,
	// 	"source_epoch": 28715,
	// 	"source_root": "0x1b41dff838043578538eac4b3ae63e3ec7af663682acde21a156a0cd29050215",
	// 	"target_epoch": 28716,
	// 	"target_root": "0xf4a1753603b00f6b9ce28e421a45b3ebda6e792f18db1f72540e94af680cb337",
	// 	"validators": [
	// 	 57356
	// 	]
	// }
	// check_excel_with_local_data()
	// sample1()
	
	// sample4()
	// check_excel()
	// writeBlockInfo()
	// readBlockInfo(0)
	// check_excel_with_local_data()
	// testClient()
	// fetchValidatorInfo("validatorinfo")
	// fetchValidatorInfo()
	// testClient()
}