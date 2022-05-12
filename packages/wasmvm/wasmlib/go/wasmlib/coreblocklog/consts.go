// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package coreblocklog

import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"

const (
	ScName        = "blocklog"
	ScDescription = "Core block log contract"
	HScName       = wasmtypes.ScHname(0xf538ef2b)
)

const (
	ParamBlockIndex    = "n"
	ParamContractHname = "h"
	ParamFromBlock     = "f"
	ParamRequestID     = "u"
	ParamToBlock       = "t"
)

const (
	ResultBlockIndex             = "n"
	ResultBlockInfo              = "i"
	ResultEvent                  = "e"
	ResultGoverningAddress       = "g"
	ResultRequestID              = "u"
	ResultRequestIndex           = "r"
	ResultRequestProcessed       = "p"
	ResultRequestRecord          = "d"
	ResultStateControllerAddress = "s"
)

const (
	ViewControlAddresses           = "controlAddresses"
	ViewGetBlockInfo               = "getBlockInfo"
	ViewGetEventsForBlock          = "getEventsForBlock"
	ViewGetEventsForContract       = "getEventsForContract"
	ViewGetEventsForRequest        = "getEventsForRequest"
	ViewGetLatestBlockInfo         = "getLatestBlockInfo"
	ViewGetRequestIDsForBlock      = "getRequestIDsForBlock"
	ViewGetRequestReceipt          = "getRequestReceipt"
	ViewGetRequestReceiptsForBlock = "getRequestReceiptsForBlock"
	ViewIsRequestProcessed         = "isRequestProcessed"
)

const (
	HViewControlAddresses           = wasmtypes.ScHname(0x796bd223)
	HViewGetBlockInfo               = wasmtypes.ScHname(0xbe89f9b3)
	HViewGetEventsForBlock          = wasmtypes.ScHname(0x36232798)
	HViewGetEventsForContract       = wasmtypes.ScHname(0x682a1922)
	HViewGetEventsForRequest        = wasmtypes.ScHname(0x4f8d68e4)
	HViewGetLatestBlockInfo         = wasmtypes.ScHname(0x084a1760)
	HViewGetRequestIDsForBlock      = wasmtypes.ScHname(0x5a20327a)
	HViewGetRequestReceipt          = wasmtypes.ScHname(0xb7f9534f)
	HViewGetRequestReceiptsForBlock = wasmtypes.ScHname(0x77e3beef)
	HViewIsRequestProcessed         = wasmtypes.ScHname(0xd57d50a9)
)