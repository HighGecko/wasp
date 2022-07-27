// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

import * as wasmlib from "wasmlib";
import * as sc from "./index";

const exportMap: wasmlib.ScExportMap = {
	names: [
		sc.FuncF,
	],
	funcs: [
		funcFThunk,
	],
	views: [
	],
};

export function on_call(index: i32): void {
	wasmlib.WasmVMHost.connect();
	wasmlib.ScExports.call(index, exportMap);
}

export function on_load(): void {
	wasmlib.WasmVMHost.connect();
	wasmlib.ScExports.export(exportMap);
}

function funcFThunk(ctx: wasmlib.ScFuncContext): void {
	ctx.log("storage.funcF");
	let f = new sc.FContext();
	ctx.require(f.params.n().exists(), "missing mandatory n");
	sc.funcF(ctx, f);
	ctx.log("storage.funcF ok");
}