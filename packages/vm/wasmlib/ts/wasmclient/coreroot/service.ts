// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

import * as wasmclient from "wasmclient"

const ArgDeployer = "dp";
const ArgDescription = "ds";
const ArgHname = "hn";
const ArgName = "nm";
const ArgProgramHash = "ph";

const ResContractFound = "cf";
const ResContractRecData = "dt";
const ResContractRegistry = "r";

///////////////////////////// deployContract /////////////////////////////

export class DeployContractFunc extends wasmclient.ClientFunc {
	private args: wasmclient.Arguments = new wasmclient.Arguments();
	
	public description(v: string): void {
		this.args.setString(ArgDescription, v);
	}
	
	public name(v: string): void {
		this.args.setString(ArgName, v);
	}
	
	public programHash(v: wasmclient.Hash): void {
		this.args.setHash(ArgProgramHash, v);
	}
	
	public async post(): Promise<wasmclient.RequestID> {
		this.args.mandatory(ArgName);
		this.args.mandatory(ArgProgramHash);
		return await super.post(0x28232c27, this.args);
	}
}

///////////////////////////// grantDeployPermission /////////////////////////////

export class GrantDeployPermissionFunc extends wasmclient.ClientFunc {
	private args: wasmclient.Arguments = new wasmclient.Arguments();
	
	public deployer(v: wasmclient.AgentID): void {
		this.args.setAgentID(ArgDeployer, v);
	}
	
	public async post(): Promise<wasmclient.RequestID> {
		this.args.mandatory(ArgDeployer);
		return await super.post(0xf440263a, this.args);
	}
}

///////////////////////////// revokeDeployPermission /////////////////////////////

export class RevokeDeployPermissionFunc extends wasmclient.ClientFunc {
	private args: wasmclient.Arguments = new wasmclient.Arguments();
	
	public deployer(v: wasmclient.AgentID): void {
		this.args.setAgentID(ArgDeployer, v);
	}
	
	public async post(): Promise<wasmclient.RequestID> {
		this.args.mandatory(ArgDeployer);
		return await super.post(0x850744f1, this.args);
	}
}

///////////////////////////// findContract /////////////////////////////

export class FindContractView extends wasmclient.ClientView {
	private args: wasmclient.Arguments = new wasmclient.Arguments();
	
	public hname(v: wasmclient.Hname): void {
		this.args.setHname(ArgHname, v);
	}

	public async call(): Promise<FindContractResults> {
		this.args.mandatory(ArgHname);
		return new FindContractResults(await this.callView("findContract", this.args));
	}
}

export class FindContractResults extends wasmclient.ViewResults {

	contractFound(): wasmclient.Bytes {
		return this.res.getBytes(ResContractFound);
	}

	contractRecData(): wasmclient.Bytes {
		return this.res.getBytes(ResContractRecData);
	}
}

///////////////////////////// getContractRecords /////////////////////////////

export class GetContractRecordsView extends wasmclient.ClientView {

	public async call(): Promise<GetContractRecordsResults> {
		return new GetContractRecordsResults(await this.callView("getContractRecords", null));
	}
}

export class GetContractRecordsResults extends wasmclient.ViewResults {

	contractRegistry(): wasmclient.Bytes {
		return this.res.getBytes(ResContractRegistry);
	}
}

///////////////////////////// CoreRootService /////////////////////////////

export class CoreRootService extends wasmclient.Service {

	public constructor(cl: wasmclient.ServiceClient) {
		super(cl, 0xcebf5908, new Map());
	}

	public deployContract(): DeployContractFunc {
		return new DeployContractFunc(this);
	}

	public grantDeployPermission(): GrantDeployPermissionFunc {
		return new GrantDeployPermissionFunc(this);
	}

	public revokeDeployPermission(): RevokeDeployPermissionFunc {
		return new RevokeDeployPermissionFunc(this);
	}

	public findContract(): FindContractView {
		return new FindContractView(this);
	}

	public getContractRecords(): GetContractRecordsView {
		return new GetContractRecordsView(this);
	}
}
