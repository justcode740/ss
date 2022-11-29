package main

import (
	"encoding/hex"
	"fmt"

	// "github.com/phoreproject/bls"
	bls2 "github.com/prysmaticlabs/prysm/v3/crypto/bls"
	"github.com/prysmaticlabs/prysm/v3/crypto/bls/common"
)
func testLib2(){
	// sig := bls2.NewAggregateSignature("")
	sigbytes, _ := hex.DecodeString("89761ec8e2a086fb073204f9d48ed1a8e805e9871de9be0bac91a5f76aa5f261c2f748f50f48eb4a7b40a32998227f440c6e89302c9260cdb08837cbedd2d50ab92e53155679905ddf1a7a828d8fcc56a4d56c4543caf0eaa66aaa6778ac8c30")
	fmt.Println(sigbytes)
	sig, _ := bls2.SignatureFromBytes(sigbytes)
	msg := getSigningRoot()
	// msg, _ := hex.DecodeString("1895bac33c8cecdf1c07a21fa0c1be683283a1c99f5964ed7bfeccc33104ecba")
	pkbytes, _ := hex.DecodeString("87acda545d5c84996d757f6222567d5ec579ef55dbc4214ef49aa2b49cfef063e0c39b4943cdfec8eba58c92c4eca96b")
	pk, _  := bls2.PublicKeyFromBytes(pkbytes)
	fmt.Println(pk)
	// proposerbytes, _ := hex.DecodeString(
		// "81690c370330ac17a48cdcfae58c89ab2061f7222409896d06134f4d89f896d2d375f567a541f11de9d2a6966ae256e7")
	// proposerpk, _ := bls2.PublicKeyFromBytes(proposerbytes)
	var pks []common.PublicKey
	pks = append(pks, pk)
	// pks = append(pks, proposerpk)
	var msgs []byte
	fmt.Println(sig)
	copy(msgs[:], msg[:])
	fmt.Println(sig.Verify(pk, msgs))
	// fmt.Println(sig.Eth2FastAggregateVerify(pks, msg))
}