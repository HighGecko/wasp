// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the schema definition file instead

import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "./index";

export class ImmutableFindContractResults extends wasmtypes.ScProxy {
	// encoded contract record
	contractFound(): wasmtypes.ScImmutableBytes {
		return new wasmtypes.ScImmutableBytes(this.proxy.root(sc.ResultContractFound));
	}

	// encoded contract record
	contractRecData(): wasmtypes.ScImmutableBytes {
		return new wasmtypes.ScImmutableBytes(this.proxy.root(sc.ResultContractRecData));
	}
}

export class MutableFindContractResults extends wasmtypes.ScProxy {
	// encoded contract record
	contractFound(): wasmtypes.ScMutableBytes {
		return new wasmtypes.ScMutableBytes(this.proxy.root(sc.ResultContractFound));
	}

	// encoded contract record
	contractRecData(): wasmtypes.ScMutableBytes {
		return new wasmtypes.ScMutableBytes(this.proxy.root(sc.ResultContractRecData));
	}
}

export class MapHnameToImmutableBytes extends wasmtypes.ScProxy {

	getBytes(key: wasmtypes.ScHname): wasmtypes.ScImmutableBytes {
		return new wasmtypes.ScImmutableBytes(this.proxy.key(wasmtypes.hnameToBytes(key)));
	}
}

export class ImmutableGetContractRecordsResults extends wasmtypes.ScProxy {
	// contract records
	contractRegistry(): sc.MapHnameToImmutableBytes {
		return new sc.MapHnameToImmutableBytes(this.proxy.root(sc.ResultContractRegistry));
	}
}

export class MapHnameToMutableBytes extends wasmtypes.ScProxy {

	clear(): void {
		this.proxy.clearMap();
	}

	getBytes(key: wasmtypes.ScHname): wasmtypes.ScMutableBytes {
		return new wasmtypes.ScMutableBytes(this.proxy.key(wasmtypes.hnameToBytes(key)));
	}
}

export class MutableGetContractRecordsResults extends wasmtypes.ScProxy {
	// contract records
	contractRegistry(): sc.MapHnameToMutableBytes {
		return new sc.MapHnameToMutableBytes(this.proxy.root(sc.ResultContractRegistry));
	}
}
