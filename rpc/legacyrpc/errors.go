// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2016-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package legacyrpc

import (
	"fmt"

	"github.com/fonero-project/fnod/fnojson"
	"github.com/fonero-project/fnowallet/errors"
)

func convertError(err error) *fnojson.RPCError {
	if err, ok := err.(*fnojson.RPCError); ok {
		return err
	}

	code := fnojson.ErrRPCWallet
	if err, ok := err.(*errors.Error); ok {
		switch err.Kind {
		case errors.Bug:
			code = fnojson.ErrRPCInternal.Code
		case errors.Encoding:
			code = fnojson.ErrRPCInvalidParameter
		case errors.Locked:
			code = fnojson.ErrRPCWalletUnlockNeeded
		case errors.Passphrase:
			code = fnojson.ErrRPCWalletPassphraseIncorrect
		case errors.NoPeers:
			code = fnojson.ErrRPCClientNotConnected
		case errors.InsufficientBalance:
			code = fnojson.ErrRPCWalletInsufficientFunds
		}
	}
	return &fnojson.RPCError{
		Code:    code,
		Message: err.Error(),
	}
}

func rpcError(code fnojson.RPCErrorCode, err error) *fnojson.RPCError {
	return &fnojson.RPCError{
		Code:    code,
		Message: err.Error(),
	}
}

func rpcErrorf(code fnojson.RPCErrorCode, format string, args ...interface{}) *fnojson.RPCError {
	return &fnojson.RPCError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Errors variables that are defined once here to avoid duplication.
var (
	errUnloadedWallet = &fnojson.RPCError{
		Code:    fnojson.ErrRPCWallet,
		Message: "request requires a wallet but wallet has not loaded yet",
	}

	errRPCClientNotConnected = &fnojson.RPCError{
		Code:    fnojson.ErrRPCClientNotConnected,
		Message: "disconnected from consensus RPC",
	}

	errNoNetwork = &fnojson.RPCError{
		Code:    fnojson.ErrRPCClientNotConnected,
		Message: "disconnected from network",
	}

	errAccountNotFound = &fnojson.RPCError{
		Code:    fnojson.ErrRPCWalletInvalidAccountName,
		Message: "account not found",
	}

	errAddressNotInWallet = &fnojson.RPCError{
		Code:    fnojson.ErrRPCWallet,
		Message: "address not found in wallet",
	}

	errNotImportedAccount = &fnojson.RPCError{
		Code:    fnojson.ErrRPCWallet,
		Message: "imported addresses must belong to the imported account",
	}

	errNeedPositiveAmount = &fnojson.RPCError{
		Code:    fnojson.ErrRPCInvalidParameter,
		Message: "amount must be positive",
	}

	errWalletUnlockNeeded = &fnojson.RPCError{
		Code:    fnojson.ErrRPCWalletUnlockNeeded,
		Message: "enter the wallet passphrase with walletpassphrase first",
	}

	errReservedAccountName = &fnojson.RPCError{
		Code:    fnojson.ErrRPCInvalidParameter,
		Message: "account name is reserved by RPC server",
	}
)
