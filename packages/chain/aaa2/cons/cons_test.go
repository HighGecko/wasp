// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/testutil/testpeers"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
)

// Here we run a single consensus instance, step by step with
// regards to the requests to external components (mempool, stateMgr, VM).
func TestBasic(t *testing.T) {
	t.Parallel()
	type test struct {
		n int
		f int
	}
	tests := []test{
		{n: 1, f: 0},  // Low N.
		{n: 2, f: 0},  // Low N.
		{n: 3, f: 0},  // Low N.
		{n: 4, f: 1},  // Smallest reasonable config.
		{n: 10, f: 3}, // Typical config?
		{n: 12, f: 3}, // Non-optimal N/F.
	}
	if !testing.Short() {
		tests = append(tests, test{n: 31, f: 10}) // Large cluster.
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v", test.n, test.f),
			func(tt *testing.T) { testBasic(tt, test.n, test.f) },
		)
	}
}

func testBasic(t *testing.T, n, f int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Node Identities and shared key.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	//
	// Construct the chain on L1.
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	//
	// Construct the chain on L1: Create the accounts.
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Construct the chain on L1: Create the origin TX.
	outs, outIDs := utxoDB.GetUnspentOutputs(originator.Address())
	originTX, chainID, err := transaction.NewChainOriginTransaction(
		originator,
		committeeAddress,
		governor.Address(),
		1_000_000,
		outs,
		outIDs,
	)
	require.NoError(t, err)
	stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	require.NoError(t, err)
	require.NotNil(t, stateAnchor)
	require.NotNil(t, aliasOutput)
	ao0 := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID.UTXOInput())
	err = utxoDB.AddToLedger(originTX)
	require.NoError(t, err)
	//
	// Construct the chain on L1: Create the Init Request TX.
	outs, outIDs = utxoDB.GetUnspentOutputs(originator.Address())
	initTX, err := transaction.NewRootInitRequestTransaction(
		originator,
		chainID,
		"my test chain",
		outs,
		outIDs,
	)
	require.NoError(t, err)
	require.NotNil(t, initTX)
	err = utxoDB.AddToLedger(initTX)
	require.NoError(t, err)
	//
	// Construct the chain on L1: Find the requests (the init request).
	initReqs := []isc.Request{}
	initReqRefs := []*isc.RequestRef{}
	outs, _ = utxoDB.GetUnspentOutputs(chainID.AsAddress())
	for outID, out := range outs {
		if out.Type() == iotago.OutputAlias {
			zeroAliasID := iotago.AliasID{}
			outAsAlias := out.(*iotago.AliasOutput)
			if outAsAlias.AliasID == *chainID.AsAliasID() {
				continue // That's our alias output, not the request, skip it here.
			}
			if outAsAlias.AliasID == zeroAliasID {
				implicitAliasID := iotago.AliasIDFromOutputID(outID)
				if implicitAliasID == *chainID.AsAliasID() {
					continue // That's our origin alias output, not the request, skip it here.
				}
			}
		}
		req, err := isc.OnLedgerFromUTXO(out, outID.UTXOInput())
		if err != nil {
			continue
		}
		initReqs = append(initReqs, req)
		initReqRefs = append(initReqRefs, isc.RequestRefFromRequest(req))
	}
	//
	// Construct the nodes.
	consInstID := []byte{1, 2, 3} // ID of the consensus.
	procConfig := coreprocessors.Config().WithNativeContracts(inccounter.Processor)
	procCache := processors.MustNew(procConfig)
	nodeIDs := nodeIDsFromPubKeys(testpeers.PublicKeys(peerIdentities))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.Named(string(nid))
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareProviders[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		nodes[nid] = cons.New(*chainID, nid, nodeSK, nodeDKShare, procCache, consInstID, nodeIDFromPubKey, nodeLog).AsGPA()
	}
	tc := gpa.NewTestContext(nodes)
	//
	// Provide inputs.
	t.Logf("############ Provide Inputs.")
	now := time.Now()
	inputs := map[gpa.NodeID]gpa.Input{}
	for _, nid := range nodeIDs {
		inputs[nid] = ao0
	}
	tc.WithInputs(inputs).RunAll()
	tc.PrintAllStatusStrings("After Inputs", t.Logf)
	//
	// Provide SM and MP responses on proposals (and time data).
	t.Logf("############ Provide TimeData and Proposals from SM/MP.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.NotNil(t, out.NeedMempoolProposal)
		require.NotNil(t, out.NeedStateMgrStateProposal)
		tc.WithMessage(cons.NewMsgMempoolProposal(nid, initReqRefs))
		tc.WithMessage(cons.NewMsgStateMgrProposalConfirmed(nid, ao0))
		tc.WithMessage(cons.NewMsgTimeData(nid, now))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM proposals", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Provide Decided Data from SM/MP.")
	kvStore := mapdb.NewMapDB()
	virtualStateAccess, err := state.CreateOriginState(kvStore, chainID)
	require.NoError(t, err)
	chainStateSync := coreutil.NewChainStateSync()
	chainStateSync.SetSolidIndex(0)
	stateBaseline := chainStateSync.GetSolidIndexBaseline()
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.NotNil(t, out.NeedMempoolRequests)
		require.NotNil(t, out.NeedStateMgrDecidedState)
		tc.WithMessage(cons.NewMsgMempoolRequests(nid, initReqs))
		tc.WithMessage(cons.NewMsgStateMgrDecidedVirtualState(nid, ao0, stateBaseline, virtualStateAccess))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("After MP/SM data", t.Logf)
	//
	// Provide Decided data from SM and MP.
	t.Logf("############ Run VM, validate the result.")
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Running, out.State)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.NotNil(t, out.NeedVMResult)
		out.NeedVMResult.Log = out.NeedVMResult.Log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelError)).Sugar() // Decrease VM logging.
		require.NoError(t, runvm.NewVMRunner().Run(out.NeedVMResult))
		tc.WithMessage(cons.NewMsgVMResult(nid, out.NeedVMResult))
	}
	tc.RunAll()
	tc.PrintAllStatusStrings("All done.", t.Logf)
	for nid, node := range nodes {
		out := node.Output().(*cons.Output)
		require.Equal(t, cons.Completed, out.State)
		require.True(t, out.Terminated)
		require.Nil(t, out.NeedMempoolProposal)
		require.Nil(t, out.NeedStateMgrStateProposal)
		require.Nil(t, out.NeedMempoolRequests)
		require.Nil(t, out.NeedStateMgrDecidedState)
		require.Nil(t, out.NeedVMResult)
		require.NotNil(t, out.ResultTransaction)
		require.NotNil(t, out.ResultNextAliasOutput)
		require.NotNil(t, out.ResultState)
		block, err := out.ResultState.ExtractBlock()
		require.NoError(t, err)
		require.NotNil(t, block)
		if nid == nodeIDs[0] { // Just do this once.
			require.NoError(t, utxoDB.AddToLedger(out.ResultTransaction))
		}
	}
}

// Run several consensus instances in a chain, receiving inputs from each other.
// This test case has much less of synchronization, because we don't wait for
// all messages to be delivered before responding to the instance requests to
// mempool, stateMgr and VM.
func TestChained(t *testing.T) {
	t.Parallel()
	type test struct {
		n int
		f int
		b int
	}
	var tests []test
	if testing.Short() {
		tests = []test{
			{n: 1, f: 0, b: 10}, // Low N
			{n: 2, f: 0, b: 10}, // Low N
			{n: 3, f: 0, b: 10}, // Low N
			{n: 4, f: 1, b: 10}, // Smallest possible resilient config.
			{n: 10, f: 3, b: 5}, // Maybe a typical config.
			{n: 12, f: 3, b: 3}, // Check a non-optimal N/F combinations.
		}
	} else {
		tests = []test{ // Block counts chosen to keep test time similar in all cases.
			{n: 1, f: 0, b: 700}, // Low N
			{n: 2, f: 0, b: 500}, // Low N
			{n: 3, f: 0, b: 300}, // Low N
			{n: 4, f: 1, b: 250}, // Smallest possible resilient config.
			{n: 10, f: 3, b: 50}, // Maybe a typical config.
			{n: 12, f: 3, b: 35}, // Check a non-optimal N/F combinations.
			{n: 31, f: 10, b: 2}, // A large cluster.
		}
	}
	for _, test := range tests {
		t.Run(
			fmt.Sprintf("N=%v,F=%v,Blocks=%v", test.n, test.f, test.b),
			func(tt *testing.T) { testChained(tt, test.n, test.f, test.b) },
		)
	}
}

func testChained(t *testing.T, n, f, b int) {
	t.Parallel()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	//
	// Node Identities, shared key and ledger.
	_, peerIdentities := testpeers.SetupKeys(uint16(n))
	committeeAddress, dkShareProviders := testpeers.SetupDkgTrivial(t, n, f, peerIdentities, nil)
	nodeIDs := nodeIDsFromPubKeys(testpeers.PublicKeys(peerIdentities))
	utxoDB := utxodb.New(utxodb.DefaultInitParams())
	//
	// Create the accounts.
	scClient := cryptolib.NewKeyPair()
	governor := cryptolib.NewKeyPair()
	originator := cryptolib.NewKeyPair()
	_, err := utxoDB.GetFundsFromFaucet(governor.Address())
	require.NoError(t, err)
	_, err = utxoDB.GetFundsFromFaucet(originator.Address())
	require.NoError(t, err)
	//
	// Construct the chain on L1 and prepare requests.
	tcl := newTestChainLedger(t, utxoDB, governor, originator)
	originAO := tcl.txChainOrigin(committeeAddress)
	allRequests := map[int][]isc.Request{}
	allRequests[0] = tcl.txChainInit()
	if b > 1 {
		_, err = utxoDB.GetFundsFromFaucet(scClient.Address(), 5_000_000)
		require.NoError(t, err)
		allRequests[1] = append(tcl.txAccountsDeposit(scClient), tcl.txDeployIncCounterContract()...)
	}
	incTotal := 0
	for i := 2; i < b; i++ {
		reqs := []isc.Request{}
		reqPerBlock := 3
		for ii := 0; ii < reqPerBlock; ii++ {
			scRequest := isc.NewOffLedgerRequest(
				tcl.chainID,
				inccounter.Contract.Hname(),
				inccounter.FuncIncCounter.Hname(),
				dict.New(), uint64(i*reqPerBlock+ii),
			).WithGasBudget(15000).Sign(scClient)
			reqs = append(reqs, scRequest)
			incTotal++
		}
		allRequests[i] = reqs
	}
	//
	// Construct the nodes for each instance.
	procConfig := coreprocessors.Config().WithNativeContracts(inccounter.Processor)
	procCache := processors.MustNew(procConfig)
	doneCHs := map[gpa.NodeID]chan *testInstInput{}
	for _, nid := range nodeIDs {
		doneCHs[nid] = make(chan *testInstInput, 1)
	}
	testNodeStates := map[gpa.NodeID]*testNodeState{}
	for _, nid := range nodeIDs {
		testNodeStates[nid] = newTestNodeState(t, tcl.chainID)
	}
	testChainInsts := make([]testConsInst, b)
	for i := range testChainInsts {
		ii := i // Copy.
		doneCB := func(nextInput *testInstInput) {
			if ii == b-1 {
				doneCHs[nextInput.nodeID] <- nextInput
				return
			}
			testChainInsts[ii+1].input(nextInput)
		}
		testChainInsts[i] = *newTestConsInst(
			t, tcl.chainID, committeeAddress, i, procCache, nodeIDs,
			testNodeStates, peerIdentities, dkShareProviders,
			allRequests[i], doneCB, log,
		)
	}
	// Start the threads for each instance.
	for i := range testChainInsts {
		go testChainInsts[i].run()
	}
	// Start the process by providing input to the first instance.
	for _, nid := range nodeIDs {
		t.Logf("Going to provide inputs.")
		testChainInsts[0].input(&testInstInput{
			nodeID:             nid,
			baseAliasOutput:    originAO,
			stateBaseline:      testNodeStates[nid].getStateBaseline(),
			virtualStateAccess: testNodeStates[nid].getVirtualStateAccess(),
		})
	}
	// Wait for all the instances to output.
	t.Logf("Waiting for DONE for the last in the chain.")
	doneVals := map[gpa.NodeID]*testInstInput{}
	for nid, doneCH := range doneCHs {
		doneVals[nid] = <-doneCH
	}
	t.Logf("Waiting for all instances to terminate.")
	for _, tci := range testChainInsts {
		<-tci.tcTerminated
	}
	t.Logf("Done, last block was output and all instances terminated.")
	for _, doneVal := range doneVals {
		require.Equal(t, int64(incTotal), inccounter.NewStateAccess(doneVal.virtualStateAccess.KVStore()).GetMaintenanceStatus())
	}
}

////////////////////////////////////////////////////////////////////////////////
// testChainLedger

type testChainLedger struct {
	t           *testing.T
	utxoDB      *utxodb.UtxoDB
	governor    *cryptolib.KeyPair
	originator  *cryptolib.KeyPair
	chainID     *isc.ChainID
	fetchedReqs map[iotago.Address]map[iotago.OutputID]bool
}

func newTestChainLedger(t *testing.T, utxoDB *utxodb.UtxoDB, governor, originator *cryptolib.KeyPair) *testChainLedger {
	return &testChainLedger{
		t:           t,
		utxoDB:      utxoDB,
		governor:    governor,
		originator:  originator,
		fetchedReqs: map[iotago.Address]map[iotago.OutputID]bool{},
	}
}

func (tcl *testChainLedger) txChainOrigin(committeeAddress iotago.Address) *isc.AliasOutputWithID {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(tcl.originator.Address())
	originTX, chainID, err := transaction.NewChainOriginTransaction(
		tcl.originator,
		committeeAddress,
		tcl.governor.Address(),
		1_000_000,
		outs,
		outIDs,
	)
	require.NoError(tcl.t, err)
	stateAnchor, aliasOutput, err := transaction.GetAnchorFromTransaction(originTX)
	require.NoError(tcl.t, err)
	require.NotNil(tcl.t, stateAnchor)
	require.NotNil(tcl.t, aliasOutput)
	originAO := isc.NewAliasOutputWithID(aliasOutput, stateAnchor.OutputID.UTXOInput())
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(originTX))
	tcl.chainID = chainID
	return originAO
}

func (tcl *testChainLedger) txChainInit() []isc.Request {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(tcl.originator.Address())
	initTX, err := transaction.NewRootInitRequestTransaction(tcl.originator, tcl.chainID, "my test chain", outs, outIDs)
	require.NoError(tcl.t, err)
	require.NotNil(tcl.t, initTX)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(initTX))
	return tcl.findChainRequests(initTX)
}

func (tcl *testChainLedger) txAccountsDeposit(account *cryptolib.KeyPair) []isc.Request {
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(account.Address())
	tx, err := transaction.NewRequestTransaction(
		transaction.NewRequestTransactionParams{
			SenderKeyPair:    account,
			SenderAddress:    account.Address(),
			UnspentOutputs:   outs,
			UnspentOutputIDs: outIDs,
			Request: &isc.RequestParameters{
				TargetAddress:                 tcl.chainID.AsAddress(),
				FungibleTokens:                isc.NewFungibleBaseTokens(2_000_000),
				AdjustToMinimumStorageDeposit: false,
				Metadata: &isc.SendMetadata{
					TargetContract: accounts.Contract.Hname(),
					EntryPoint:     accounts.FuncDeposit.Hname(),
					Allowance:      isc.NewEmptyAllowance().AddBaseTokens(10000), // TODO: ...
					GasBudget:      10_000,
				},
			},
			// NFT: par.NFT,
		},
	)
	require.NoError(tcl.t, err)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(tx))
	return tcl.findChainRequests(tx)
}

func (tcl *testChainLedger) txDeployIncCounterContract() []isc.Request {
	sender := tcl.originator
	outs, outIDs := tcl.utxoDB.GetUnspentOutputs(sender.Address())
	tx, err := transaction.NewRequestTransaction(
		transaction.NewRequestTransactionParams{
			SenderKeyPair:    sender,
			SenderAddress:    sender.Address(),
			UnspentOutputs:   outs,
			UnspentOutputIDs: outIDs,
			Request: &isc.RequestParameters{
				TargetAddress:                 tcl.chainID.AsAddress(),
				FungibleTokens:                isc.NewFungibleBaseTokens(2_000_000),
				AdjustToMinimumStorageDeposit: false,
				Metadata: &isc.SendMetadata{
					TargetContract: root.Contract.Hname(),
					EntryPoint:     root.FuncDeployContract.Hname(),
					Params: codec.MakeDict(map[string]interface{}{
						root.ParamProgramHash: inccounter.Contract.ProgramHash,
						root.ParamDescription: "inccounter",
						root.ParamName:        inccounter.Contract.Name,
						inccounter.VarCounter: 0,
					}),
					Allowance: isc.NewEmptyAllowance().AddBaseTokens(10000),
					GasBudget: 10_000,
				},
			},
		},
	)
	require.NoError(tcl.t, err)
	require.NoError(tcl.t, tcl.utxoDB.AddToLedger(tx))
	return tcl.findChainRequests(tx)
}

func (tcl *testChainLedger) findChainRequests(tx *iotago.Transaction) []isc.Request {
	reqs := []isc.Request{}
	outs, err := tx.OutputsSet()
	require.NoError(tcl.t, err)
	for outID, out := range outs {
		// If that's alias output of the chain, then it is not a request.
		if out.Type() == iotago.OutputAlias {
			zeroAliasID := iotago.AliasID{}
			outAsAlias := out.(*iotago.AliasOutput)
			if outAsAlias.AliasID == *tcl.chainID.AsAliasID() {
				continue // That's our alias output, not the request, skip it here.
			}
			if outAsAlias.AliasID == zeroAliasID {
				implicitAliasID := iotago.AliasIDFromOutputID(outID)
				if implicitAliasID == *tcl.chainID.AsAliasID() {
					continue // That's our origin alias output, not the request, skip it here.
				}
			}
		}
		//
		// Otherwise check the receiving address.
		outAddr := out.UnlockConditionSet().Address()
		if outAddr == nil {
			continue
		}
		if !outAddr.Address.Equal(tcl.chainID.AsAddress()) {
			continue
		}
		req, err := isc.OnLedgerFromUTXO(out, outID.UTXOInput())
		if err != nil {
			continue
		}
		reqs = append(reqs, req)
	}
	return reqs
}

////////////////////////////////////////////////////////////////////////////////
// testNodeState

type testNodeState struct {
	kvStore   kvstore.KVStore
	stateSync coreutil.ChainStateSync
	vsAccess  state.VirtualStateAccess
}

func newTestNodeState(t *testing.T, chainID *isc.ChainID) *testNodeState {
	kvStore := mapdb.NewMapDB()
	stateSync := coreutil.NewChainStateSync()
	stateSync.SetSolidIndex(0)
	vsAccess, err := state.CreateOriginState(kvStore, chainID)
	require.NoError(t, err)
	return &testNodeState{
		kvStore:   kvStore,
		stateSync: stateSync,
		vsAccess:  vsAccess,
	}
}

func (tns *testNodeState) getVirtualStateAccess() state.VirtualStateAccess {
	return tns.vsAccess
}

func (tns *testNodeState) getStateBaseline() coreutil.StateBaseline {
	return tns.stateSync.GetSolidIndexBaseline()
}

////////////////////////////////////////////////////////////////////////////////
// testConsInst

type testInstInput struct {
	nodeID             gpa.NodeID
	baseAliasOutput    *isc.AliasOutputWithID
	stateBaseline      coreutil.StateBaseline
	virtualStateAccess state.VirtualStateAccess
}

type testConsInst struct {
	t            *testing.T
	nodes        map[gpa.NodeID]gpa.GPA
	nodeStates   map[gpa.NodeID]*testNodeState
	stateIndex   int
	requests     []isc.Request
	tc           *gpa.TestContext
	tcInputCh    chan map[gpa.NodeID]gpa.Input // These channels are for sending data to TC.
	tcMessageCh  chan gpa.Message              // These channels are for sending data to TC.
	tcTerminated chan interface{}
	//
	// Inputs received from the previous instance.
	lock              *sync.RWMutex                 // inputs value is checked in the TC thread and written in TCI.
	messagePipe       chan gpa.Message              // This queue is used to send message from TCI/TC to TCI.
	messagePipeClosed *atomic.Bool                  // Can be closed from TC or TCI.
	inputCh           chan *testInstInput           // These channels are to send data to TCI.
	inputs            map[gpa.NodeID]*testInstInput // Inputs received to this TCI.
	//
	// The latest output of the consensus instances.
	outLatest map[gpa.NodeID]*cons.Output
	//
	// What has been provided to the consensus instances.
	handledNeedMempoolProposal       map[gpa.NodeID]bool
	handledNeedStateMgrStateProposal map[gpa.NodeID]bool
	handledNeedMempoolRequests       map[gpa.NodeID]bool
	handledNeedStateMgrDecidedState  map[gpa.NodeID]bool
	handledNeedVMResult              map[gpa.NodeID]bool
	//
	// Result of this instance provided to the next instance.
	done   map[gpa.NodeID]bool
	doneCB func(nextInput *testInstInput)
}

func newTestConsInst(
	t *testing.T,
	chainID *isc.ChainID,
	committeeAddress iotago.Address,
	stateIndex int,
	procCache *processors.Cache,
	nodeIDs []gpa.NodeID,
	nodeStates map[gpa.NodeID]*testNodeState,
	peerIdentities []*cryptolib.KeyPair,
	dkShareProviders []registry.DKShareRegistryProvider,
	requests []isc.Request,
	doneCB func(nextInput *testInstInput),
	log *logger.Logger,
) *testConsInst {
	consInstID := []byte(fmt.Sprintf("testConsInst-%v", stateIndex))
	nodes := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nodeLog := log.Named(string(nid))
		nodeSK := peerIdentities[i].GetPrivateKey()
		nodeDKShare, err := dkShareProviders[i].LoadDKShare(committeeAddress)
		require.NoError(t, err)
		nodes[nid] = cons.New(*chainID, nid, nodeSK, nodeDKShare, procCache, consInstID, nodeIDFromPubKey, nodeLog).AsGPA()
	}
	tci := &testConsInst{
		t:                                t,
		nodes:                            nodes,
		nodeStates:                       nodeStates,
		stateIndex:                       stateIndex,
		requests:                         requests,
		tcInputCh:                        make(chan map[gpa.NodeID]gpa.Input, len(nodeIDs)),
		tcMessageCh:                      make(chan gpa.Message, len(nodeIDs)),
		tcTerminated:                     make(chan interface{}),
		lock:                             &sync.RWMutex{},
		messagePipe:                      make(chan gpa.Message, len(nodeIDs)*10),
		messagePipeClosed:                &atomic.Bool{},
		inputCh:                          make(chan *testInstInput, len(nodeIDs)),
		inputs:                           map[gpa.NodeID]*testInstInput{},
		outLatest:                        map[gpa.NodeID]*cons.Output{},
		handledNeedMempoolProposal:       map[gpa.NodeID]bool{},
		handledNeedStateMgrStateProposal: map[gpa.NodeID]bool{},
		handledNeedMempoolRequests:       map[gpa.NodeID]bool{},
		handledNeedStateMgrDecidedState:  map[gpa.NodeID]bool{},
		handledNeedVMResult:              map[gpa.NodeID]bool{},
		done:                             map[gpa.NodeID]bool{},
		doneCB:                           doneCB,
	}
	tci.tc = gpa.NewTestContext(nodes).
		WithOutputHandler(tci.outputHandler).
		WithInputChannel(tci.tcInputCh).
		WithMessageChannel(tci.tcMessageCh)
	return tci
}

func (tci *testConsInst) run() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		tci.tc.RunAll()
		close(tci.tcTerminated)
		cancel()
	}()
	ticks := time.After(10 * time.Millisecond)
	tickSent := 0
	tickClose := false
	var timeForStatus <-chan time.Time
	for {
		select {
		case inp := <-tci.inputCh:
			tci.lock.Lock()
			if _, ok := tci.inputs[inp.nodeID]; ok {
				tci.lock.Unlock()
				panic("duplicate input")
			}
			tci.inputs[inp.nodeID] = inp
			inputsLen := len(tci.inputs)
			tci.lock.Unlock()
			tci.tcInputCh <- map[gpa.NodeID]gpa.Input{inp.nodeID: inp.baseAliasOutput}
			timeForStatus = time.After(3 * time.Second)
			if inputsLen == len(tci.nodes) {
				close(tci.tcInputCh)
			}
			tci.tryHandleOutput(inp.nodeID)
		case msg, ok := <-tci.messagePipe:
			if !ok {
				tickClose = true
				if tickSent > 0 {
					close(tci.tcMessageCh)
				}
				tci.messagePipe = nil
				continue
			}
			tci.tcMessageCh <- msg
		case <-ctx.Done():
			return
		case <-ticks:
			if tickClose && tickSent > 0 {
				continue // tci.tcMessageCh already closed.
			}
			tickSent++
			for nodeID := range tci.nodes {
				tci.tcMessageCh <- cons.NewMsgTimeData(nodeID, time.Now())
			}
			if tickClose {
				close(tci.tcMessageCh)
				continue
			}
			ticks = time.After(20 * time.Millisecond)
		case <-timeForStatus:
			tci.tc.PrintAllStatusStrings(fmt.Sprintf("TCI[%v] timeForStatus", tci.stateIndex), tci.t.Logf)
			timeForStatus = time.After(3 * time.Second)
		}
	}
}

func (tci *testConsInst) input(input *testInstInput) {
	tci.inputCh <- input
}

func (tci *testConsInst) outputHandler(nodeID gpa.NodeID, out gpa.Output) {
	tci.lock.Lock()
	tci.outLatest[nodeID] = out.(*cons.Output)
	tci.lock.Unlock()
	tci.tryHandleOutput(nodeID)
}

// Here we respond to the node requests to other components (provided via the output).
// This can be executed in the TCI (on input) and TC (on output) threads.
func (tci *testConsInst) tryHandleOutput(nodeID gpa.NodeID) {
	tci.lock.Lock()
	defer tci.lock.Unlock()
	out, ok := tci.outLatest[nodeID]
	if !ok {
		return
	}
	inp, ok := tci.inputs[nodeID]
	if !ok {
		return
	}
	switch out.State {
	case cons.Completed:
		if tci.done[nodeID] {
			return
		}
		tci.doneCB(&testInstInput{
			nodeID:             nodeID,
			baseAliasOutput:    out.ResultNextAliasOutput,
			stateBaseline:      tci.nodeStates[nodeID].getStateBaseline(),
			virtualStateAccess: out.ResultState,
		})
		tci.done[nodeID] = true
		return
	case cons.Skipped:
		if tci.done[nodeID] {
			return
		}
		tci.doneCB(inp)
		tci.done[nodeID] = true
		return
	}
	tci.tryHandledNeedMempoolProposal(nodeID, out, inp)
	tci.tryHandledNeedStateMgrStateProposal(nodeID, out, inp)
	tci.tryHandledNeedMempoolRequests(nodeID, out)
	tci.tryHandledNeedStateMgrDecidedState(nodeID, out, inp)
	tci.tryHandledNeedVMResult(nodeID, out)
	allClosed := true
	for nid := range tci.nodes {
		if tci.handledNeedMempoolProposal[nid] &&
			tci.handledNeedStateMgrStateProposal[nid] &&
			tci.handledNeedMempoolRequests[nid] &&
			tci.handledNeedStateMgrDecidedState[nid] &&
			tci.handledNeedVMResult[nid] {
			continue
		}
		allClosed = false
	}
	if allClosed {
		tci.tryCloseMessagePipe()
	}
}

func (tci *testConsInst) tryHandledNeedMempoolProposal(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedMempoolProposal != nil && !tci.handledNeedMempoolProposal[nodeID] {
		require.Equal(tci.t, out.NeedMempoolProposal, inp.baseAliasOutput)
		reqRefs := []*isc.RequestRef{}
		for _, r := range tci.requests {
			reqRefs = append(reqRefs, isc.RequestRefFromRequest(r))
		}
		tci.messagePipe <- cons.NewMsgMempoolProposal(nodeID, reqRefs)
		tci.handledNeedMempoolProposal[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrStateProposal(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedStateMgrStateProposal != nil && !tci.handledNeedStateMgrStateProposal[nodeID] {
		require.Equal(tci.t, out.NeedStateMgrStateProposal, inp.baseAliasOutput)
		tci.messagePipe <- cons.NewMsgStateMgrProposalConfirmed(nodeID, inp.baseAliasOutput)
		tci.handledNeedStateMgrStateProposal[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedMempoolRequests(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedMempoolRequests != nil && !tci.handledNeedMempoolRequests[nodeID] {
		requests := []isc.Request{}
		for _, reqRef := range out.NeedMempoolRequests {
			for _, req := range tci.requests {
				if reqRef.IsFor(req) {
					requests = append(requests, req)
					break
				}
			}
		}
		if len(requests) == len(out.NeedMempoolRequests) {
			tci.messagePipe <- cons.NewMsgMempoolRequests(nodeID, requests)
		} else {
			// TODO: Otherwise we have to sync between mempools.
			tci.t.Errorf("TEST: we have to sync between mempools.")
		}
		tci.handledNeedMempoolRequests[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedStateMgrDecidedState(nodeID gpa.NodeID, out *cons.Output, inp *testInstInput) {
	if out.NeedStateMgrDecidedState != nil && !tci.handledNeedStateMgrDecidedState[nodeID] {
		if *out.NeedStateMgrDecidedState == inp.baseAliasOutput.OutputID() {
			tci.messagePipe <- cons.NewMsgStateMgrDecidedVirtualState(nodeID, inp.baseAliasOutput, inp.stateBaseline, inp.virtualStateAccess)
		} else {
			// TODO: Otherwise we have to sync between state managers.
			tci.t.Errorf("TEST: we have to sync between state managers.")
		}
		tci.handledNeedStateMgrDecidedState[nodeID] = true
	}
}

func (tci *testConsInst) tryHandledNeedVMResult(nodeID gpa.NodeID, out *cons.Output) {
	if out.NeedVMResult != nil && !tci.handledNeedVMResult[nodeID] {
		out.NeedVMResult.Log = out.NeedVMResult.Log.Desugar().WithOptions(zap.IncreaseLevel(logger.LevelError)).Sugar() // Decrease VM logging.
		require.NoError(tci.t, runvm.NewVMRunner().Run(out.NeedVMResult))
		tci.messagePipe <- cons.NewMsgVMResult(nodeID, out.NeedVMResult)
		tci.handledNeedVMResult[nodeID] = true
	}
}

func (tci *testConsInst) tryCloseMessagePipe() {
	if !tci.messagePipeClosed.Swap(true) {
		close(tci.messagePipe)
	}
}

////////////////////////////////////////////////////////////////////////////////
// Helper functions.

func nodeIDsFromPubKeys(pubKeys []*cryptolib.PublicKey) []gpa.NodeID {
	ret := make([]gpa.NodeID, len(pubKeys))
	for i := range pubKeys {
		ret[i] = nodeIDFromPubKey(pubKeys[i])
	}
	return ret
}

func nodeIDFromPubKey(pubKey *cryptolib.PublicKey) gpa.NodeID {
	return gpa.NodeID("N#" + pubKey.String()[:6])
}
