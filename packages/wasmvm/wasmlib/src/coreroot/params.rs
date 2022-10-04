// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the schema definition file instead

#![allow(dead_code)]
#![allow(unused_imports)]

use crate::coreroot::*;
use crate::*;

#[derive(Clone)]
pub struct ImmutableDeployContractParams {
	pub(crate) proxy: Proxy,
}

impl ImmutableDeployContractParams {
    // default 'N/A'
    pub fn description(&self) -> ScImmutableString {
		ScImmutableString::new(self.proxy.root(PARAM_DESCRIPTION))
	}

    pub fn name(&self) -> ScImmutableString {
		ScImmutableString::new(self.proxy.root(PARAM_NAME))
	}

    // TODO variable init params for deployed contract
    pub fn program_hash(&self) -> ScImmutableHash {
		ScImmutableHash::new(self.proxy.root(PARAM_PROGRAM_HASH))
	}
}

#[derive(Clone)]
pub struct MutableDeployContractParams {
	pub(crate) proxy: Proxy,
}

impl MutableDeployContractParams {
    // default 'N/A'
    pub fn description(&self) -> ScMutableString {
		ScMutableString::new(self.proxy.root(PARAM_DESCRIPTION))
	}

    pub fn name(&self) -> ScMutableString {
		ScMutableString::new(self.proxy.root(PARAM_NAME))
	}

    // TODO variable init params for deployed contract
    pub fn program_hash(&self) -> ScMutableHash {
		ScMutableHash::new(self.proxy.root(PARAM_PROGRAM_HASH))
	}
}

#[derive(Clone)]
pub struct ImmutableGrantDeployPermissionParams {
	pub(crate) proxy: Proxy,
}

impl ImmutableGrantDeployPermissionParams {
    pub fn deployer(&self) -> ScImmutableAgentID {
		ScImmutableAgentID::new(self.proxy.root(PARAM_DEPLOYER))
	}
}

#[derive(Clone)]
pub struct MutableGrantDeployPermissionParams {
	pub(crate) proxy: Proxy,
}

impl MutableGrantDeployPermissionParams {
    pub fn deployer(&self) -> ScMutableAgentID {
		ScMutableAgentID::new(self.proxy.root(PARAM_DEPLOYER))
	}
}

#[derive(Clone)]
pub struct ImmutableRequireDeployPermissionsParams {
	pub(crate) proxy: Proxy,
}

impl ImmutableRequireDeployPermissionsParams {
    pub fn deploy_permissions_enabled(&self) -> ScImmutableBool {
		ScImmutableBool::new(self.proxy.root(PARAM_DEPLOY_PERMISSIONS_ENABLED))
	}
}

#[derive(Clone)]
pub struct MutableRequireDeployPermissionsParams {
	pub(crate) proxy: Proxy,
}

impl MutableRequireDeployPermissionsParams {
    pub fn deploy_permissions_enabled(&self) -> ScMutableBool {
		ScMutableBool::new(self.proxy.root(PARAM_DEPLOY_PERMISSIONS_ENABLED))
	}
}

#[derive(Clone)]
pub struct ImmutableRevokeDeployPermissionParams {
	pub(crate) proxy: Proxy,
}

impl ImmutableRevokeDeployPermissionParams {
    pub fn deployer(&self) -> ScImmutableAgentID {
		ScImmutableAgentID::new(self.proxy.root(PARAM_DEPLOYER))
	}
}

#[derive(Clone)]
pub struct MutableRevokeDeployPermissionParams {
	pub(crate) proxy: Proxy,
}

impl MutableRevokeDeployPermissionParams {
    pub fn deployer(&self) -> ScMutableAgentID {
		ScMutableAgentID::new(self.proxy.root(PARAM_DEPLOYER))
	}
}

#[derive(Clone)]
pub struct ImmutableSubscribeBlockContextParams {
	pub(crate) proxy: Proxy,
}

impl ImmutableSubscribeBlockContextParams {
    pub fn close_func(&self) -> ScImmutableHname {
		ScImmutableHname::new(self.proxy.root(PARAM_CLOSE_FUNC))
	}

    pub fn open_func(&self) -> ScImmutableHname {
		ScImmutableHname::new(self.proxy.root(PARAM_OPEN_FUNC))
	}
}

#[derive(Clone)]
pub struct MutableSubscribeBlockContextParams {
	pub(crate) proxy: Proxy,
}

impl MutableSubscribeBlockContextParams {
    pub fn close_func(&self) -> ScMutableHname {
		ScMutableHname::new(self.proxy.root(PARAM_CLOSE_FUNC))
	}

    pub fn open_func(&self) -> ScMutableHname {
		ScMutableHname::new(self.proxy.root(PARAM_OPEN_FUNC))
	}
}

#[derive(Clone)]
pub struct ImmutableFindContractParams {
	pub(crate) proxy: Proxy,
}

impl ImmutableFindContractParams {
    pub fn hname(&self) -> ScImmutableHname {
		ScImmutableHname::new(self.proxy.root(PARAM_HNAME))
	}
}

#[derive(Clone)]
pub struct MutableFindContractParams {
	pub(crate) proxy: Proxy,
}

impl MutableFindContractParams {
    pub fn hname(&self) -> ScMutableHname {
		ScMutableHname::new(self.proxy.root(PARAM_HNAME))
	}
}
