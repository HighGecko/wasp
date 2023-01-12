// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the schema definition file instead

#![allow(dead_code)]
#![allow(unused_imports)]

use wasmlib::*;
use solotutorial::*;
use crate::*;

const EXPORT_MAP: ScExportMap = ScExportMap {
    names: &[
        FUNC_STORE_STRING,
        VIEW_GET_STRING,
    ],
    funcs: &[
        func_store_string_thunk,
    ],
    views: &[
        view_get_string_thunk,
    ],
};

pub fn on_dispatch(index: i32) {
    EXPORT_MAP.dispatch(index);
}

pub struct StoreStringContext {
    pub params: ImmutableStoreStringParams,
    pub state:  MutableSoloTutorialState,
}

fn func_store_string_thunk(ctx: &ScFuncContext) {
    ctx.log("solotutorial.funcStoreString");
    let f = StoreStringContext {
        params: ImmutableStoreStringParams::new(),
        state:  MutableSoloTutorialState::new(),
    };
    ctx.require(f.params.str().exists(), "missing mandatory param: str");
    func_store_string(ctx, &f);
    ctx.log("solotutorial.funcStoreString ok");
}

pub struct GetStringContext {
    pub results: MutableGetStringResults,
    pub state:   ImmutableSoloTutorialState,
}

fn view_get_string_thunk(ctx: &ScViewContext) {
    ctx.log("solotutorial.viewGetString");
    let f = GetStringContext {
        results: MutableGetStringResults::new(),
        state:   ImmutableSoloTutorialState::new(),
    };
    view_get_string(ctx, &f);
    ctx.results(&f.results.proxy);
    ctx.log("solotutorial.viewGetString ok");
}
