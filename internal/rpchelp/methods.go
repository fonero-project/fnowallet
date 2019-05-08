// Copyright (c) 2015 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

//+build !generate

package rpchelp

import "github.com/fonero-project/fnod/fnojson"

// Common return types.
var (
	returnsBool        = []interface{}{(*bool)(nil)}
	returnsNumber      = []interface{}{(*float64)(nil)}
	returnsString      = []interface{}{(*string)(nil)}
	returnsStringArray = []interface{}{(*[]string)(nil)}
	returnsLTRArray    = []interface{}{(*[]fnojson.ListTransactionsResult)(nil)}
)

// Methods contains all methods and result types that help is generated for,
// for every locale.
var Methods = []struct {
	Method      string
	ResultTypes []interface{}
}{
	{"accountaddressindex", []interface{}{(*int)(nil)}},
	{"accountsyncaddressindex", nil},
	{"addmultisigaddress", returnsString},
	{"addticket", nil},
	{"consolidate", returnsString},
	{"createmultisig", []interface{}{(*fnojson.CreateMultiSigResult)(nil)}},
	{"createnewaccount", nil},
	{"dumpprivkey", returnsString},
	{"exportwatchingwallet", returnsString},
	{"generatevote", []interface{}{(*fnojson.GenerateVoteResult)(nil)}},
	{"getaccountaddress", returnsString},
	{"getaccount", returnsString},
	{"getaddressesbyaccount", returnsStringArray},
	{"getbalance", []interface{}{(*fnojson.GetBalanceResult)(nil)}},
	{"getbestblockhash", returnsString},
	{"getbestblock", []interface{}{(*fnojson.GetBestBlockResult)(nil)}},
	{"getblockcount", returnsNumber},
	{"getinfo", []interface{}{(*fnojson.InfoWalletResult)(nil)}},
	{"getmasterpubkey", []interface{}{(*string)(nil)}},
	{"getmultisigoutinfo", []interface{}{(*fnojson.GetMultisigOutInfoResult)(nil)}},
	{"getnewaddress", returnsString},
	{"getrawchangeaddress", returnsString},
	{"getreceivedbyaccount", returnsNumber},
	{"getreceivedbyaddress", returnsNumber},
	{"getstakeinfo", []interface{}{(*fnojson.GetStakeInfoResult)(nil)}},
	{"getticketfee", returnsNumber},
	{"gettickets", []interface{}{(*fnojson.GetTicketsResult)(nil)}},
	{"gettransaction", []interface{}{(*fnojson.GetTransactionResult)(nil)}},
	{"getunconfirmedbalance", returnsNumber},
	{"getvotechoices", []interface{}{(*fnojson.GetVoteChoicesResult)(nil)}},
	{"getwalletfee", returnsNumber},
	{"help", append(returnsString, returnsString[0])},
	{"importprivkey", nil},
	{"importscript", nil},
	{"keypoolrefill", nil},
	{"listaccounts", []interface{}{(*map[string]float64)(nil)}},
	{"listaddresstransactions", returnsLTRArray},
	{"listalltransactions", returnsLTRArray},
	{"listlockunspent", []interface{}{(*[]fnojson.TransactionInput)(nil)}},
	{"listreceivedbyaccount", []interface{}{(*[]fnojson.ListReceivedByAccountResult)(nil)}},
	{"listreceivedbyaddress", []interface{}{(*[]fnojson.ListReceivedByAddressResult)(nil)}},
	{"listscripts", []interface{}{(*fnojson.ListScriptsResult)(nil)}},
	{"listsinceblock", []interface{}{(*fnojson.ListSinceBlockResult)(nil)}},
	{"listtransactions", returnsLTRArray},
	{"listunspent", []interface{}{(*fnojson.ListUnspentResult)(nil)}},
	{"lockunspent", returnsBool},
	{"purchaseticket", returnsString},
	{"redeemmultisigout", []interface{}{(*fnojson.RedeemMultiSigOutResult)(nil)}},
	{"redeemmultisigouts", []interface{}{(*fnojson.RedeemMultiSigOutResult)(nil)}},
	{"renameaccount", nil},
	{"rescanwallet", nil},
	{"revoketickets", nil},
	{"sendfrom", returnsString},
	{"sendmany", returnsString},
	{"sendtoaddress", returnsString},
	{"sendtomultisig", returnsString},
	{"setticketfee", returnsBool},
	{"settxfee", returnsBool},
	{"setvotechoice", nil},
	{"signmessage", returnsString},
	{"signrawtransaction", []interface{}{(*fnojson.SignRawTransactionResult)(nil)}},
	{"signrawtransactions", []interface{}{(*fnojson.SignRawTransactionsResult)(nil)}},
	{"stakepooluserinfo", []interface{}{(*fnojson.StakePoolUserInfoResult)(nil)}},
	{"startautobuyer", nil},
	{"stopautobuyer", nil},
	{"sweepaccount", []interface{}{(*fnojson.SweepAccountResult)(nil)}},
	{"ticketsforaddress", returnsBool},
	{"validateaddress", []interface{}{(*fnojson.ValidateAddressWalletResult)(nil)}},
	{"verifymessage", returnsBool},
	{"version", []interface{}{(*map[string]fnojson.VersionResult)(nil)}},
	{"walletinfo", []interface{}{(*fnojson.WalletInfoResult)(nil)}},
	{"walletislocked", returnsBool},
	{"walletlock", nil},
	{"walletpassphrasechange", nil},
	{"walletpassphrase", nil},
}

// HelpDescs contains the locale-specific help strings along with the locale.
var HelpDescs = []struct {
	Locale   string // Actual locale, e.g. en_US
	GoLocale string // Locale used in Go names, e.g. EnUS
	Descs    map[string]string
}{
	{"en_US", "EnUS", helpDescsEnUS}, // helpdescs_en_US.go
}
