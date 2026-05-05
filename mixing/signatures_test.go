// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mixing

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

var (
	testPrivKey = secp256k1.PrivKeyFromBytes([]byte{31: 1})
	testPubKey  = testPrivKey.PubKey()
)

func hexDecode(tb testing.TB, s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		tb.Fatalf("Hex decode failed: %v", err)
	}
	return b
}

func fakePR(tb testing.TB) *wire.MsgMixPairReq {
	var id [33]byte
	copy(id[:], testPubKey.SerializeCompressed())
	utxos := []wire.MixPairReqUTXO{{
		OutPoint: wire.OutPoint{
			Hash:  [32]byte{31: 1},
			Index: 2,
			// Tree:  3,
		},
		Script:    []byte{4, 5, 6},
		PubKey:    []byte{31: 7},
		Signature: []byte{31: 8},
		Opcode:    9,
	}}
	change := &wire.TxOut{
		Value: 10,
		// Version:  11,
		PkScript: []byte{12, 13, 14},
	}
	pr, err := wire.NewMsgMixPairReq(id, 15, 16, "script class", 17, 18, 19, 20, utxos, change, 21, 22)
	if err != nil {
		tb.Fatal(err)
	}
	return pr
}

const wantSig = "3a88093cc784a934c90bd41ff96d59e0e13d2111c0430a0a734d76f1cd4de91b50e250487b8fdb3ceebfee34166babd395c512c4273c25d8f2c482b6695aae76"

func TestSignMessage(t *testing.T) {
	pr := fakePR(t)

	if err := SignMessage(pr, testPrivKey); err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if !bytes.Equal(pr.Signature[:], hexDecode(t, wantSig)) {
		t.Fatalf("Signature does not match expected (got: %x want: %s)", pr.Signature[:], wantSig)
	}
}

func TestVerifySignedMessage(t *testing.T) {
	pr := fakePR(t)

	if err := SignMessage(pr, testPrivKey); err != nil {
		t.Fatalf("Failed to sign message: %v", err)
	}

	if !VerifySignedMessage(pr) {
		t.Fatalf("VerifySignedMessage invalid signature %x", pr.Signature[:])
	}
}

func BenchmarkSignMessage(b *testing.B) {
	pr := fakePR(b)

	for b.Loop() {
		err := SignMessage(pr, testPrivKey)
		if err != nil {
			b.Fatalf("Failed to sign message: %v", err)
		}
	}
}

func BenchmarkVerifySignedMessage(b *testing.B) {
	pr := fakePR(b)
	err := SignMessage(pr, testPrivKey)
	if err != nil {
		b.Fatalf("Failed to sign message: %v", err)
	}

	for b.Loop() {
		if !VerifySignedMessage(pr) {
			b.Fatalf("VerifySignedMessage invalid signature %x", pr.Signature[:])
		}
	}
}
