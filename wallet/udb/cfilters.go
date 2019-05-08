// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package udb

import (
	"github.com/fonero-project/fnod/chaincfg/chainhash"
	"github.com/fonero-project/fnod/gcs"
	"github.com/fonero-project/fnod/gcs/blockcf"
	"github.com/fonero-project/fnowallet/wallet/walletdb"
)

// CFilter returns the saved regular compact filter for a block.
func (s *Store) CFilter(dbtx walletdb.ReadTx, blockHash *chainhash.Hash) (*gcs.Filter, error) {
	ns := dbtx.ReadBucket(wtxmgrBucketKey)
	v, err := fetchRawCFilter(ns, blockHash[:])
	if err != nil {
		return nil, err
	}
	vc := make([]byte, len(v)) // Copy for FromNBytes which stores passed slice
	copy(vc, v)
	return gcs.FromNBytes(blockcf.P, vc)
}
