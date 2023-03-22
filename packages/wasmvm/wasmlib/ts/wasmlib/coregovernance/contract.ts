// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the schema definition file instead

import * as wasmlib from '../index';
import * as sc from './index';

export class AddAllowedStateControllerAddressCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableAddAllowedStateControllerAddressParams = new sc.MutableAddAllowedStateControllerAddressParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncAddAllowedStateControllerAddress);
    }
}

export class AddCandidateNodeCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableAddCandidateNodeParams = new sc.MutableAddCandidateNodeParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncAddCandidateNode);
    }
}

export class ChangeAccessNodesCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableChangeAccessNodesParams = new sc.MutableChangeAccessNodesParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncChangeAccessNodes);
    }
}

export class ClaimChainOwnershipCall {
    func: wasmlib.ScFunc;

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncClaimChainOwnership);
    }
}

export class DelegateChainOwnershipCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableDelegateChainOwnershipParams = new sc.MutableDelegateChainOwnershipParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncDelegateChainOwnership);
    }
}

export class RemoveAllowedStateControllerAddressCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableRemoveAllowedStateControllerAddressParams = new sc.MutableRemoveAllowedStateControllerAddressParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncRemoveAllowedStateControllerAddress);
    }
}

export class RevokeAccessNodeCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableRevokeAccessNodeParams = new sc.MutableRevokeAccessNodeParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncRevokeAccessNode);
    }
}

export class RotateStateControllerCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableRotateStateControllerParams = new sc.MutableRotateStateControllerParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncRotateStateController);
    }
}

export class SetFeePolicyCall {
    func:   wasmlib.ScFunc;
    params: sc.MutableSetFeePolicyParams = new sc.MutableSetFeePolicyParams(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScFuncCallContext) {
        this.func = new wasmlib.ScFunc(ctx, sc.HScName, sc.HFuncSetFeePolicy);
    }
}

export class GetAllowedStateControllerAddressesCall {
    func:    wasmlib.ScView;
    results: sc.ImmutableGetAllowedStateControllerAddressesResults = new sc.ImmutableGetAllowedStateControllerAddressesResults(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScViewCallContext) {
        this.func = new wasmlib.ScView(ctx, sc.HScName, sc.HViewGetAllowedStateControllerAddresses);
    }
}

export class GetChainInfoCall {
    func:    wasmlib.ScView;
    results: sc.ImmutableGetChainInfoResults = new sc.ImmutableGetChainInfoResults(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScViewCallContext) {
        this.func = new wasmlib.ScView(ctx, sc.HScName, sc.HViewGetChainInfo);
    }
}

export class GetChainNodesCall {
    func:    wasmlib.ScView;
    results: sc.ImmutableGetChainNodesResults = new sc.ImmutableGetChainNodesResults(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScViewCallContext) {
        this.func = new wasmlib.ScView(ctx, sc.HScName, sc.HViewGetChainNodes);
    }
}

export class GetChainOwnerCall {
    func:    wasmlib.ScView;
    results: sc.ImmutableGetChainOwnerResults = new sc.ImmutableGetChainOwnerResults(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScViewCallContext) {
        this.func = new wasmlib.ScView(ctx, sc.HScName, sc.HViewGetChainOwner);
    }
}

export class GetFeePolicyCall {
    func:    wasmlib.ScView;
    results: sc.ImmutableGetFeePolicyResults = new sc.ImmutableGetFeePolicyResults(wasmlib.ScView.nilProxy);

    public constructor(ctx: wasmlib.ScViewCallContext) {
        this.func = new wasmlib.ScView(ctx, sc.HScName, sc.HViewGetFeePolicy);
    }
}

export class ScFuncs {
    static addAllowedStateControllerAddress(ctx: wasmlib.ScFuncCallContext): AddAllowedStateControllerAddressCall {
        const f = new AddAllowedStateControllerAddressCall(ctx);
        f.params = new sc.MutableAddAllowedStateControllerAddressParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    // access nodes
    static addCandidateNode(ctx: wasmlib.ScFuncCallContext): AddCandidateNodeCall {
        const f = new AddCandidateNodeCall(ctx);
        f.params = new sc.MutableAddCandidateNodeParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    static changeAccessNodes(ctx: wasmlib.ScFuncCallContext): ChangeAccessNodesCall {
        const f = new ChangeAccessNodesCall(ctx);
        f.params = new sc.MutableChangeAccessNodesParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    // chain owner
    static claimChainOwnership(ctx: wasmlib.ScFuncCallContext): ClaimChainOwnershipCall {
        return new ClaimChainOwnershipCall(ctx);
    }

    static delegateChainOwnership(ctx: wasmlib.ScFuncCallContext): DelegateChainOwnershipCall {
        const f = new DelegateChainOwnershipCall(ctx);
        f.params = new sc.MutableDelegateChainOwnershipParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    static removeAllowedStateControllerAddress(ctx: wasmlib.ScFuncCallContext): RemoveAllowedStateControllerAddressCall {
        const f = new RemoveAllowedStateControllerAddressCall(ctx);
        f.params = new sc.MutableRemoveAllowedStateControllerAddressParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    static revokeAccessNode(ctx: wasmlib.ScFuncCallContext): RevokeAccessNodeCall {
        const f = new RevokeAccessNodeCall(ctx);
        f.params = new sc.MutableRevokeAccessNodeParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    // state controller
    static rotateStateController(ctx: wasmlib.ScFuncCallContext): RotateStateControllerCall {
        const f = new RotateStateControllerCall(ctx);
        f.params = new sc.MutableRotateStateControllerParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    // fees
    static setFeePolicy(ctx: wasmlib.ScFuncCallContext): SetFeePolicyCall {
        const f = new SetFeePolicyCall(ctx);
        f.params = new sc.MutableSetFeePolicyParams(wasmlib.newCallParamsProxy(f.func));
        return f;
    }

    // state controller
    static getAllowedStateControllerAddresses(ctx: wasmlib.ScViewCallContext): GetAllowedStateControllerAddressesCall {
        const f = new GetAllowedStateControllerAddressesCall(ctx);
        f.results = new sc.ImmutableGetAllowedStateControllerAddressesResults(wasmlib.newCallResultsProxy(f.func));
        return f;
    }

    // chain info
    static getChainInfo(ctx: wasmlib.ScViewCallContext): GetChainInfoCall {
        const f = new GetChainInfoCall(ctx);
        f.results = new sc.ImmutableGetChainInfoResults(wasmlib.newCallResultsProxy(f.func));
        return f;
    }

    // access nodes
    static getChainNodes(ctx: wasmlib.ScViewCallContext): GetChainNodesCall {
        const f = new GetChainNodesCall(ctx);
        f.results = new sc.ImmutableGetChainNodesResults(wasmlib.newCallResultsProxy(f.func));
        return f;
    }

    // chain owner
    static getChainOwner(ctx: wasmlib.ScViewCallContext): GetChainOwnerCall {
        const f = new GetChainOwnerCall(ctx);
        f.results = new sc.ImmutableGetChainOwnerResults(wasmlib.newCallResultsProxy(f.func));
        return f;
    }

    // fees
    static getFeePolicy(ctx: wasmlib.ScViewCallContext): GetFeePolicyCall {
        const f = new GetFeePolicyCall(ctx);
        f.results = new sc.ImmutableGetFeePolicyResults(wasmlib.newCallResultsProxy(f.func));
        return f;
    }
}
