// Copyright (c) 2024 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mixing

import (
	"bytes"
	"encoding/hex"
	"errors"
	"sort"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func TestSessionID(t *testing.T) {
	// PR hashes must be lexicographically sorted before call to
	// deriveSessionID.
	prs := []chainhash.Hash{{0}, {1}, {2}, {3}, {4}, {5}}
	sid := deriveSessionID(prs, 0)

	t.Logf("sid: %x", sid[:])
	for i := range prs {
		t.Logf("%d: %s", i, prs[i])
	}

	xorPRs := append(prs[:0:0], prs...)
	for i := range xorPRs {
		xor(xorPRs[i][:], sid[:])
	}
	for i := range xorPRs {
		t.Logf("xorPRs[%d] before sort: %s", i, xorPRs[i])
	}

	sort.Slice(xorPRs, func(i, j int) bool {
		return bytes.Compare(xorPRs[i][:], xorPRs[j][:]) == -1
	})
	for i := range xorPRs {
		t.Logf("xorPRs[%d] after sort:  %s", i, xorPRs[i])
	}

	prPosition := make(map[chainhash.Hash]int)
	for i := range xorPRs {
		originalHash := xorPRs[i]
		xor(originalHash[:], sid[:])
		prPosition[originalHash] = i
	}

	orderedPRs := make([]chainhash.Hash, len(prs))
	for prHash, pos := range prPosition {
		orderedPRs[pos] = prHash
	}
	for i := range orderedPRs {
		t.Logf("orderedPRs[%d]: %s", i, orderedPRs[i])
	}

	ke := &wire.MsgMixKeyExchange{
		SessionID: sid,
		Epoch:     0,
		Run:       0,
		SeenPRs:   orderedPRs,
	}
	if err := ValidateSession(ke); err != nil {
		t.Errorf("ValidateSession: %v", err)
	}
}

func TestInvalidSession(t *testing.T) {
	// Session and PR order from TestSessionID.
	prs := []chainhash.Hash{{4}, {5}, {2}, {3}, {0}, {1}}
	sidBytes, _ := hex.DecodeString("06808fdf9a834ddbcf2ae371e6297e0c3f006a52260372884baa3e6b3ca50456")
	sid := *(*[32]byte)(sidBytes)

	ke := &wire.MsgMixKeyExchange{
		SessionID: sid,
		Epoch:     0,
		Run:       0,
		SeenPRs:   prs,
	}
	if err := ValidateSession(ke); err != nil {
		t.Fatalf("ValidateSession: %v", err)
	}

	ke.SessionID[16] ^= 0xFF
	if err := ValidateSession(ke); !errors.Is(err, errInvalidSessionID) {
		t.Errorf("ValidateSession unexpected error, got %v, want %v",
			err, errInvalidSessionID)
	}

	ke.SessionID = sid
	ke.SeenPRs[0], ke.SeenPRs[1] = ke.SeenPRs[1], ke.SeenPRs[0]
	if err := ValidateSession(ke); !errors.Is(err, errInvalidPROrder) {
		t.Errorf("ValidateSession unexpected error, got %v, want %v",
			err, errInvalidPROrder)
	}
}
