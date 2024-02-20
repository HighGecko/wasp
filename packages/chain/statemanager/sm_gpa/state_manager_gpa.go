package sm_gpa

import (
	"fmt"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_gpa_utils"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_inputs"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_gpa/sm_messages"
	"github.com/iotaledger/wasp/packages/chain/statemanager/sm_utils"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/pipe"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

type blockInfo struct {
	trieRoot   trie.Hash
	blockIndex uint32
}

type stateManagerGPA struct {
	log                      *logger.Logger
	chainID                  isc.ChainID
	blockCache               sm_gpa_utils.BlockCache
	blocksToFetch            blockFetchers
	blocksFetched            blockFetchers
	loadedSnapshotStateIndex uint32
	nodeRandomiser           sm_utils.NodeRandomiser
	store                    state.Store
	output                   StateManagerOutput
	parameters               StateManagerParameters
	chainOfBlocks            pipe.Deque[*blockInfo]
	lastGetBlocksTime        time.Time
	lastCleanBlockCacheTime  time.Time
	lastCleanRequestsTime    time.Time
	lastStatusLogTime        time.Time
	metrics                  *metrics.ChainStateManagerMetrics
}

var _ gpa.GPA = &stateManagerGPA{}

const (
	numberOfNodesToRequestBlockFromConst = 5
	statusLogPeriodConst                 = 10 * time.Second
)

func New(
	chainID isc.ChainID,
	loadedSnapshotStateIndex uint32,
	nr sm_utils.NodeRandomiser,
	wal sm_gpa_utils.BlockWAL,
	store state.Store,
	metrics *metrics.ChainStateManagerMetrics,
	log *logger.Logger,
	parameters StateManagerParameters,
) (gpa.GPA, error) {
	var err error
	smLog := log.Named("GPA")
	blockCache, err := sm_gpa_utils.NewBlockCache(parameters.TimeProvider, parameters.BlockCacheMaxSize, wal, metrics, smLog)
	if err != nil {
		return nil, fmt.Errorf("error creating block cache: %v", err)
	}
	result := &stateManagerGPA{
		log:                      smLog,
		chainID:                  chainID,
		blockCache:               blockCache,
		blocksToFetch:            newBlockFetchers(newBlockFetchersMetrics(metrics.IncBlocksFetching, metrics.DecBlocksFetching, metrics.StateManagerBlockFetched)),
		blocksFetched:            newBlockFetchers(newBlockFetchersMetrics(metrics.IncBlocksPending, metrics.DecBlocksPending, bfmNopDurationFun)),
		loadedSnapshotStateIndex: loadedSnapshotStateIndex,
		nodeRandomiser:           nr,
		store:                    store,
		output:                   newOutput(),
		parameters:               parameters,
		chainOfBlocks:            nil,
		lastGetBlocksTime:        time.Time{},
		lastCleanBlockCacheTime:  time.Time{},
		lastCleanRequestsTime:    time.Time{},
		lastStatusLogTime:        time.Time{},
		metrics:                  metrics,
	}

	return result, nil
}

// -------------------------------------
// Implementation for gpa.GPA interface
// -------------------------------------

func (smT *stateManagerGPA) Input(input gpa.Input) gpa.OutMessages {
	switch inputCasted := input.(type) {
	case *sm_inputs.ConsensusStateProposal: // From consensus
		return smT.handleConsensusStateProposal(inputCasted)
	case *sm_inputs.ConsensusDecidedState: // From consensus
		return smT.handleConsensusDecidedState(inputCasted)
	case *sm_inputs.ConsensusBlockProduced: // From consensus
		return smT.handleConsensusBlockProduced(inputCasted)
	case *sm_inputs.ChainFetchStateDiff: // From mempool
		return smT.handleChainFetchStateDiff(inputCasted)
	case *sm_inputs.StateManagerBlocksToCommit: // From state manager gpa
		return smT.handleStateManagerBlocksToCommit(inputCasted.GetCommitments())
	case *sm_inputs.StateManagerTimerTick: // From state manager go routine
		return smT.handleStateManagerTimerTick(inputCasted.GetTime())
	default:
		smT.log.Warnf("Unknown input received, ignoring it: type=%T, message=%v", input, input)
		return nil // No messages to send
	}
}

func (smT *stateManagerGPA) Message(msg gpa.Message) gpa.OutMessages {
	switch msgCasted := msg.(type) {
	case *sm_messages.GetBlockMessage:
		return smT.handlePeerGetBlock(msgCasted.Sender(), msgCasted.GetL1Commitment())
	case *sm_messages.BlockMessage:
		return smT.handlePeerBlock(msgCasted.Sender(), msgCasted.GetBlock())
	default:
		smT.log.Warnf("Unknown message received, ignoring it: type=%T, message=%v", msg, msg)
		return nil // No messages to send
	}
}

func (smT *stateManagerGPA) Output() gpa.Output {
	return smT.output
}

func (smT *stateManagerGPA) StatusString() string {
	return fmt.Sprintf(
		"State manager is waiting for %v blocks from other nodes; "+
			"%v blocks are obtained and waiting to be committed; "+
			"%v requests are waiting for response; "+
			"last time blocks were requested from peer nodes: %v (every %v); "+
			"last time outdated requests were cleared: %v (every %v); "+
			"last time block cache was cleaned: %v (every %v).",
		smT.blocksToFetch.getSize(),
		smT.blocksFetched.getSize(),
		smT.getWaitingCallbacksCount(),
		util.TimeOrNever(smT.lastGetBlocksTime), smT.parameters.StateManagerGetBlockRetry,
		util.TimeOrNever(smT.lastCleanRequestsTime), smT.parameters.StateManagerRequestCleaningPeriod,
		util.TimeOrNever(smT.lastCleanBlockCacheTime), smT.parameters.BlockCacheBlockCleaningPeriod,
	)
}

func (smT *stateManagerGPA) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		sm_messages.MsgTypeBlockMessage:    func() gpa.Message { return sm_messages.NewEmptyBlockMessage() },
		sm_messages.MsgTypeGetBlockMessage: func() gpa.Message { return sm_messages.NewEmptyGetBlockMessage() },
	})
}

// -------------------------------------
// Internal functions
// -------------------------------------

func (smT *stateManagerGPA) handlePeerGetBlock(from gpa.NodeID, commitment *state.L1Commitment) gpa.OutMessages {
	// TODO: [KP] Only accept queries from access nodes.
	fromLog := from.ShortString()
	smT.log.Debugf("Message GetBlock %s received from peer %s", commitment, fromLog)
	block := smT.getBlock(commitment)
	if block == nil {
		smT.log.Debugf("Message GetBlock %s: block not found, peer %s request ignored", commitment, fromLog)
		return nil // No messages to send
	}
	smT.log.Debugf("Message GetBlock %s: block index %v found, sending it to peer %s", commitment, block.StateIndex(), fromLog)
	return gpa.NoMessages().Add(sm_messages.NewBlockMessage(block, from))
}

func (smT *stateManagerGPA) handlePeerBlock(from gpa.NodeID, block state.Block) gpa.OutMessages {
	blockIndex := block.StateIndex()
	blockCommitment := block.L1Commitment()
	fromLog := from.ShortString()
	smT.log.Debugf("Message Block index %v %s received from peer %s", blockIndex, blockCommitment, fromLog)
	fetcher := smT.blocksToFetch.takeFetcher(blockCommitment)
	if fetcher == nil {
		smT.log.Debugf("Message Block index %v %s: block is not needed, ignoring it", blockIndex, blockCommitment)
		return nil // No messages to send
	}
	smT.blockCache.AddBlock(block)
	messages := smT.traceBlockChain(fetcher)
	smT.log.Debugf("Message Block index %v %s from peer %s handled", blockIndex, blockCommitment, fromLog)
	return messages
}

func (smT *stateManagerGPA) handleConsensusStateProposal(csp *sm_inputs.ConsensusStateProposal) gpa.OutMessages {
	start := time.Now()
	smT.log.Debugf("Input consensus state proposal index %v %s received...", csp.GetStateIndex(), csp.GetL1Commitment())
	callback := newBlockRequestCallback(
		func() bool {
			return csp.IsValid()
		},
		func() {
			csp.Respond()
			smT.log.Debugf("Input consensus state proposal index %v %s: responded to consensus", csp.GetStateIndex(), csp.GetL1Commitment())
			smT.metrics.ConsensusStateProposalHandled(time.Since(start))
		},
	)
	messages := smT.traceBlockChainWithCallback(csp.GetStateIndex(), csp.GetL1Commitment(), callback)
	smT.log.Debugf("Input consensus state proposal index %v %s handled", csp.GetStateIndex(), csp.GetL1Commitment())
	return messages
}

func (smT *stateManagerGPA) handleConsensusDecidedState(cds *sm_inputs.ConsensusDecidedState) gpa.OutMessages {
	start := time.Now()
	smT.log.Debugf("Input consensus decided state index %v %s received...", cds.GetStateIndex(), cds.GetL1Commitment())
	callback := newBlockRequestCallback(
		func() bool {
			return cds.IsValid()
		},
		func() {
			state, err := smT.store.StateByTrieRoot(cds.GetL1Commitment().TrieRoot())
			if err != nil {
				smT.log.Errorf("Input consensus decided state index %v %s: error obtaining state: %w", cds.GetStateIndex(), cds.GetL1Commitment(), err)
				return
			}
			cds.Respond(state)
			smT.log.Debugf("Input consensus decided state index %v %s: responded to consensus with state index %v",
				cds.GetStateIndex(), cds.GetL1Commitment(), state.BlockIndex())
			smT.metrics.ConsensusDecidedStateHandled(time.Since(start))
		},
	)
	messages := smT.traceBlockChainWithCallback(cds.GetStateIndex(), cds.GetL1Commitment(), callback)
	smT.log.Debugf("Input consensus decided state index %v %s handled", cds.GetStateIndex(), cds.GetL1Commitment())
	return messages
}

func (smT *stateManagerGPA) handleConsensusBlockProduced(input *sm_inputs.ConsensusBlockProduced) gpa.OutMessages {
	start := time.Now()
	stateIndex := input.GetStateDraft().BlockIndex() - 1 // NOTE: as this state draft is complete, the returned index is the one of the next state (which will be obtained, once this state draft is committed); to get the index of the base state, we need to subtract one
	commitment := input.GetStateDraft().BaseL1Commitment()
	smT.log.Debugf("Input block produced on state index %v %s received...", stateIndex, commitment)
	if !smT.store.HasTrieRoot(commitment.TrieRoot()) {
		smT.log.Panicf("Input block produced on state index %v %s: state, on which this block is produced, is not yet in the store", stateIndex, commitment)
	}
	// NOTE: committing already committed block is allowed (see `TestDoubleCommit` test in `packages/state/state_test.go`)
	block := smT.commitStateDraft(input.GetStateDraft())
	blockCommitment := block.L1Commitment()
	smT.blockCache.AddBlock(block)
	input.Respond(block)
	smT.log.Debugf("Input block produced on state index %v %s: state draft has been committed to the store, responded to consensus with resulting block index %v %s",
		stateIndex, commitment, block.StateIndex(), blockCommitment)
	fetcher := smT.blocksToFetch.takeFetcher(blockCommitment)
	var result gpa.OutMessages
	if fetcher != nil {
		result = smT.markFetched(fetcher, false)
	}
	smT.log.Debugf("Input block produced on state index %v %s handled", stateIndex, commitment)
	smT.metrics.ConsensusBlockProducedHandled(time.Since(start))
	return result // No messages to send
}

func (smT *stateManagerGPA) handleChainFetchStateDiff(input *sm_inputs.ChainFetchStateDiff) gpa.OutMessages {
	start := time.Now()
	smT.log.Debugf("Input mempool state request for state index %v %s is received compared to state index %v %s...",
		input.GetNewStateIndex(), input.GetNewL1Commitment(), input.GetOldStateIndex(), input.GetOldL1Commitment())
	oldBlockRequestCompleted := false
	newBlockRequestCompleted := false
	isValidFun := func() bool { return input.IsValid() }
	respondIfNeededFun := func() {
		if oldBlockRequestCompleted && newBlockRequestCompleted {
			smT.handleChainFetchStateDiffRespond(input, start)
		}
	}
	oldRequestCallback := newBlockRequestCallback(isValidFun, func() {
		oldBlockRequestCompleted = true
		smT.log.Debugf("Input mempool state request for state index %v %s: new block request completed",
			input.GetNewStateIndex(), input.GetNewL1Commitment())
		respondIfNeededFun()
	})
	newRequestCallback := newBlockRequestCallback(isValidFun, func() {
		newBlockRequestCompleted = true
		smT.log.Debugf("Input mempool state request for state index %v %s: old block request completed",
			input.GetNewStateIndex(), input.GetNewL1Commitment())
		respondIfNeededFun()
	})
	result := gpa.NoMessages()
	result.AddAll(smT.traceBlockChainWithCallback(input.GetOldStateIndex(), input.GetOldL1Commitment(), oldRequestCallback))
	result.AddAll(smT.traceBlockChainWithCallback(input.GetNewStateIndex(), input.GetNewL1Commitment(), newRequestCallback))
	smT.log.Debugf("Input mempool state request for state index %v %s handled",
		input.GetNewStateIndex(), input.GetNewL1Commitment())
	return result
}

func (smT *stateManagerGPA) handleChainFetchStateDiffRespond(input *sm_inputs.ChainFetchStateDiff, start time.Time) { //nolint:funlen
	makeCallbackFun := func(part string) blockRequestCallback {
		return newBlockRequestCallback(
			func() bool { return input.IsValid() },
			func() {
				smT.log.Debugf("Input mempool state request for state index %v %s: %s block request completed once again",
					input.GetNewStateIndex(), input.GetNewL1Commitment(), part)
				smT.handleChainFetchStateDiffRespond(input, start)
			},
		)
	}
	obtainCommittedPreviousBlockFun := func(block state.Block, part string) state.Block {
		commitment := block.PreviousL1Commitment()
		result := smT.getBlock(commitment)
		if result == nil {
			blockIndex := block.StateIndex() - 1
			smT.log.Debugf("Input mempool state request for state index %v %s: block %v %s in the %s block chain is missing; fetching it",
				input.GetNewStateIndex(), input.GetNewL1Commitment(), blockIndex, commitment, part)
			// NOTE: returned messages are not sent out; only GetBlock messages are possible in this case and
			// 		 these messages will be sent out at the next retry;
			smT.traceBlockChainWithCallback(blockIndex, commitment, makeCallbackFun(part))
		}
		return result
	}
	lastBlockFun := func(blocks []state.Block) state.Block {
		return blocks[len(blocks)-1]
	}
	oldBlock := smT.getBlock(input.GetOldL1Commitment())
	if oldBlock == nil {
		smT.log.Panicf("Input mempool state request for state index %v %s: cannot obtain final old block %s",
			input.GetNewStateIndex(), input.GetNewL1Commitment(), input.GetOldL1Commitment())
		return
	}
	newBlock := smT.getBlock(input.GetNewL1Commitment())
	if newBlock == nil {
		smT.log.Panicf("Input mempool state request for state index %v %s: cannot obtain final new block %s",
			input.GetNewStateIndex(), input.GetNewL1Commitment(), input.GetNewL1Commitment())
		return
	}
	oldChainOfBlocks := []state.Block{oldBlock}
	newChainOfBlocks := []state.Block{newBlock}
	for lastBlockFun(oldChainOfBlocks).StateIndex() > lastBlockFun(newChainOfBlocks).StateIndex() {
		oldBlock = obtainCommittedPreviousBlockFun(lastBlockFun(oldChainOfBlocks), "old")
		if oldBlock == nil {
			return
		}
		oldChainOfBlocks = append(oldChainOfBlocks, oldBlock)
	}
	for lastBlockFun(oldChainOfBlocks).StateIndex() < lastBlockFun(newChainOfBlocks).StateIndex() {
		newBlock = obtainCommittedPreviousBlockFun(lastBlockFun(newChainOfBlocks), "new")
		if newBlock == nil {
			return
		}
		newChainOfBlocks = append(newChainOfBlocks, newBlock)
	}
	for lastBlockFun(oldChainOfBlocks).StateIndex() > 0 {
		if lastBlockFun(oldChainOfBlocks).L1Commitment().Equals(lastBlockFun(newChainOfBlocks).L1Commitment()) {
			break
		}
		oldBlock = obtainCommittedPreviousBlockFun(lastBlockFun(oldChainOfBlocks), "old")
		if oldBlock == nil {
			return
		}
		newBlock = obtainCommittedPreviousBlockFun(lastBlockFun(newChainOfBlocks), "new")
		if newBlock == nil {
			return
		}
		oldChainOfBlocks = append(oldChainOfBlocks, oldBlock)
		newChainOfBlocks = append(newChainOfBlocks, newBlock)
	}
	commonIndex := lastBlockFun(oldChainOfBlocks).StateIndex()
	commonCommitment := lastBlockFun(oldChainOfBlocks).L1Commitment()
	oldChainOfBlocks = lo.Reverse(oldChainOfBlocks[:len(oldChainOfBlocks)-1])
	newChainOfBlocks = lo.Reverse(newChainOfBlocks[:len(newChainOfBlocks)-1])
	newState, err := smT.store.StateByTrieRoot(input.GetNewL1Commitment().TrieRoot())
	if err != nil {
		smT.log.Errorf("Input mempool state request for state index %v %s: error obtaining state: %w",
			input.GetNewStateIndex(), input.GetNewL1Commitment(), err)
		return
	}
	input.Respond(sm_inputs.NewChainFetchStateDiffResults(newState, newChainOfBlocks, oldChainOfBlocks))
	smT.log.Debugf("Input mempool state request for state index %v %s: responded to chain with requested state, "+
		"and block chains of length %v (requested) and %v (old) with common ancestor index %v %s",
		input.GetNewStateIndex(), input.GetNewL1Commitment(), len(newChainOfBlocks), len(oldChainOfBlocks),
		commonIndex, commonCommitment)
	smT.metrics.ChainFetchStateDiffHandled(time.Since(start))
}

func (smT *stateManagerGPA) handleStateManagerBlocksToCommit(commitments []*state.L1Commitment) gpa.OutMessages {
	smT.log.Debugf("Input state manager blocks to commit %s is received", commitments)
	result := gpa.NoMessages()
	for _, commitment := range commitments {
		fetcher := smT.blocksFetched.takeFetcher(commitment)
		if fetcher == nil {
			smT.log.Warnf("Input state manager blocks to commit %s: blocks waiting to be committed does not contain block %s; probably it is has already been committed",
				commitments, commitment)
		} else {
			result.AddAll(smT.markFetched(fetcher, true))
		}
	}
	smT.log.Debugf("Input state manager blocks to commit %s handled", commitments)
	return result
}

func (smT *stateManagerGPA) getBlock(commitment *state.L1Commitment) state.Block {
	block := smT.blockCache.GetBlock(commitment)
	if block != nil {
		return block
	}

	// Check in store (DB).
	if !smT.store.HasTrieRoot(commitment.TrieRoot()) {
		return nil
	}
	var err error
	block, err = smT.store.BlockByTrieRoot(commitment.TrieRoot())
	if err != nil {
		smT.log.Errorf("Loading block %s from the DB failed: %v", commitment, err)
		return nil
	}
	if !commitment.BlockHash().Equals(block.Hash()) {
		smT.log.Errorf("Block %s loaded from the database has hash %s",
			commitment, block.Hash())
		return nil
	}
	if !commitment.TrieRoot().Equals(block.TrieRoot()) {
		smT.log.Errorf("Block %s loaded from the database has trie root %s",
			commitment, block.TrieRoot())
		return nil
	}
	smT.log.Debugf("Block %s with index %v loaded from the database", commitment, block.StateIndex())
	smT.blockCache.AddBlock(block)
	return block
}

func (smT *stateManagerGPA) traceBlockChainWithCallback(index uint32, lastCommitment *state.L1Commitment, callback blockRequestCallback) gpa.OutMessages {
	if smT.store.HasTrieRoot(lastCommitment.TrieRoot()) {
		smT.log.Debugf("Tracing block index %v %s chain: the block is already in the store, calling back", index, lastCommitment)
		callback.requestCompleted()
		return nil // No messages to send
	}
	if smT.blocksToFetch.addCallback(lastCommitment, callback) {
		smT.metrics.IncRequestsWaiting()
		smT.log.Debugf("Tracing block index %v %s chain: the block is already being fetched", index, lastCommitment)
		return nil
	}
	if smT.blocksFetched.addCallback(lastCommitment, callback) {
		smT.metrics.IncRequestsWaiting()
		smT.log.Debugf("Tracing block index %v %s chain: the block is already fetched, but cannot yet be committed", index, lastCommitment)
		return nil
	}
	fetcher := newBlockFetcherWithCallback(index, lastCommitment, callback)
	smT.metrics.IncRequestsWaiting()
	return smT.traceBlockChain(fetcher)
}

// TODO: state manager may ask for several requests at once: the request can be
// formulated as "give me blocks from some commitment till some index". If the
// requested node has the required block committed into the store, it certainly
// has all the blocks before it.
func (smT *stateManagerGPA) traceBlockChain(initFetcher blockFetcher) gpa.OutMessages {
	var fetcher blockFetcher
	var previousCommitment *state.L1Commitment
	for fetcher = initFetcher; !smT.store.HasTrieRoot(fetcher.getCommitment().TrieRoot()); fetcher = newBlockFetcherWithRelatedFetcher(previousCommitment, fetcher) {
		stateIndex := fetcher.getStateIndex()
		commitment := fetcher.getCommitment()
		block := smT.blockCache.GetBlock(commitment)
		if block == nil {
			var stateIndexBoundaryValid bool
			stateIndexBoundary, err := smT.store.LargestPrunedBlockIndex()
			if err != nil {
				smT.log.Warnf("Cannot obtain largest pruned block: %v", err)
				stateIndexBoundary = 0
				stateIndexBoundaryValid = false
			} else {
				stateIndexBoundaryValid = true
			}
			if smT.loadedSnapshotStateIndex > stateIndexBoundary {
				stateIndexBoundary = smT.loadedSnapshotStateIndex
				stateIndexBoundaryValid = true
			}
			if (stateIndex <= stateIndexBoundary) && stateIndexBoundaryValid {
				smT.log.Panicf("Cannot find block index %v %s, because its index is not above boundary %v",
					stateIndex, commitment, stateIndexBoundary)
			}
			smT.blocksToFetch.addFetcher(fetcher)
			smT.log.Debugf("Block %s is missing, starting fetching it", commitment)
			return smT.makeGetBlockRequestMessages(commitment)
		}
		blockIndex := block.StateIndex()
		previousBlockIndex := blockIndex - 1
		previousCommitment = block.PreviousL1Commitment()
		smT.log.Debugf("Tracing block index %v %s -> previous block %v %s", blockIndex, commitment, previousBlockIndex, previousCommitment)
		if previousCommitment == nil {
			result := smT.markFetched(fetcher, true)
			smT.log.Debugf("Traced to the initial block")
			return result
		}
		smT.blocksFetched.addFetcher(fetcher)
		if smT.blocksToFetch.addRelatedFetcher(previousCommitment, fetcher) {
			smT.log.Debugf("Block %v %s is already being fetched", previousBlockIndex, previousCommitment)
			return nil // No messages to send
		}
		if smT.blocksFetched.addRelatedFetcher(previousCommitment, fetcher) {
			smT.log.Debugf("Block %v %s is already fetched, but cannot yet be committed", previousBlockIndex, previousCommitment)
			return nil // No messages to send
		}
	}
	result := smT.markFetched(fetcher, false)
	smT.log.Debugf("Block %s is already committed", fetcher.getCommitment())
	return result
}

func (smT *stateManagerGPA) markFetched(fetcher blockFetcher, doCommit bool) gpa.OutMessages {
	if doCommit {
		commitment := fetcher.getCommitment()
		block := smT.blockCache.GetBlock(commitment)
		if block == nil {
			// Block was previously received but it is no longer in cache and
			// for some unexpected reasons it is not in WAL: rerequest it
			smT.log.Warnf("Block %s was previously obtained, but it can neither be found in cache nor in WAL. Rerequesting it.", commitment)
			smT.blocksToFetch.addFetcher(fetcher)
			return gpa.NoMessages().AddAll(smT.makeGetBlockRequestMessages(commitment))
		}
		blockIndex := block.StateIndex()
		// Commit block
		var stateDraft state.StateDraft
		previousCommitment := block.PreviousL1Commitment()
		if previousCommitment == nil {
			// origin block
			stateDraft = smT.store.NewOriginStateDraft()
		} else {
			var err error
			stateDraft, err = smT.store.NewEmptyStateDraft(previousCommitment)
			if err != nil {
				smT.log.Panicf("Error creating empty state draft to store block index %v %s: %w", blockIndex, commitment, err)
			}
		}
		block.Mutations().ApplyTo(stateDraft)
		committedBlock := smT.commitStateDraft(stateDraft)
		committedCommitment := committedBlock.L1Commitment()
		if !committedCommitment.Equals(commitment) {
			smT.log.Panicf("Block index %v, received after committing (%s), differs from the block, which was committed (%s)",
				blockIndex, committedCommitment, commitment)
		}
		smT.log.Debugf("Block index %v %s has been committed to the store on state %s",
			blockIndex, commitment, previousCommitment)
	}
	relatedFetchers := fetcher.notifyFetched()
	smT.metrics.SubRequestsWaiting(fetcher.getCallbacksCount())
	relatedCommitments := lo.Map(relatedFetchers, func(f blockFetcher, i int) *state.L1Commitment {
		return f.getCommitment()
	})
	smT.log.Debugf("Blocks %s will be committed in the next iteration", relatedCommitments)
	smT.output.addBlocksToCommit(relatedCommitments)
	return nil // No messages to send
}

// Make `numberOfNodesToRequestBlockFromConst` messages to random peers
func (smT *stateManagerGPA) makeGetBlockRequestMessages(commitment *state.L1Commitment) gpa.OutMessages {
	nodeIDs := smT.nodeRandomiser.GetRandomOtherNodeIDs(numberOfNodesToRequestBlockFromConst)
	response := gpa.NoMessages()
	for _, nodeID := range nodeIDs {
		response.Add(sm_messages.NewGetBlockMessage(commitment, nodeID))
	}
	return response
}

func (smT *stateManagerGPA) handleStateManagerTimerTick(now time.Time) gpa.OutMessages {
	start := time.Now()
	result := gpa.NoMessages()
	nextStatusLogTime := smT.lastStatusLogTime.Add(statusLogPeriodConst)
	if now.After(nextStatusLogTime) {
		smT.log.Debugf("State manager gpa status: %s", smT.StatusString())
		smT.lastStatusLogTime = now
	}
	nextGetBlocksTime := smT.lastGetBlocksTime.Add(smT.parameters.StateManagerGetBlockRetry)
	if now.After(nextGetBlocksTime) {
		commitments := smT.blocksToFetch.getCommitments()
		for _, commitment := range commitments {
			result.AddAll(smT.makeGetBlockRequestMessages(commitment))
		}
		smT.lastGetBlocksTime = now
		smT.log.Debugf("Resent getBlock messages for blocks %s, next resend not earlier than %v",
			commitments, smT.lastGetBlocksTime.Add(smT.parameters.StateManagerGetBlockRetry))
	}
	nextCleanBlockCacheTime := smT.lastCleanBlockCacheTime.Add(smT.parameters.BlockCacheBlockCleaningPeriod)
	if now.After(nextCleanBlockCacheTime) {
		smT.blockCache.CleanOlderThan(now.Add(-smT.parameters.BlockCacheBlocksInCacheDuration))
		smT.lastCleanBlockCacheTime = now
		smT.log.Debugf("Block cache cleaned, %v blocks remaining, next cleaning not earlier than %v",
			smT.blockCache.Size(), smT.lastCleanBlockCacheTime.Add(smT.parameters.BlockCacheBlockCleaningPeriod))
	}
	nextCleanRequestsTime := smT.lastCleanRequestsTime.Add(smT.parameters.StateManagerRequestCleaningPeriod)
	if now.After(nextCleanRequestsTime) {
		smT.blocksToFetch.cleanCallbacks()
		smT.blocksFetched.cleanCallbacks()
		smT.lastCleanRequestsTime = now
		waitingCallbacks := smT.getWaitingCallbacksCount()
		smT.metrics.SetRequestsWaiting(waitingCallbacks)
		smT.log.Debugf("Callbacks of block fetchers cleaned, %v waiting callbacks remained, next cleaning not earlier than %v",
			waitingCallbacks, smT.lastCleanRequestsTime.Add(smT.parameters.StateManagerRequestCleaningPeriod))
	}
	smT.metrics.StateManagerTimerTickHandled(time.Since(start))
	return result
}

func (smT *stateManagerGPA) getWaitingCallbacksCount() int {
	return smT.blocksToFetch.getCallbacksCount() + smT.blocksFetched.getCallbacksCount()
}

func (smT *stateManagerGPA) commitStateDraft(stateDraft state.StateDraft) state.Block {
	block := smT.store.Commit(stateDraft)
	stateIndex := block.StateIndex()
	smT.metrics.BlockIndexCommitted(stateIndex)
	if smT.pruningNeeded() {
		smT.pruneStore(block.PreviousL1Commitment(), stateIndex-1)
	}
	smT.output.addBlockCommitted(stateIndex, block.L1Commitment())
	return block
}

func (smT *stateManagerGPA) pruningNeeded() bool {
	return smT.parameters.PruningMinStatesToKeep > 0
}

func (smT *stateManagerGPA) pruneStore(commitment *state.L1Commitment, stateIndex uint32) {
	if commitment == nil {
		return // Nothing to prune
	}
	start := time.Now()

	smT.updateChainOfBlocks(commitment, stateIndex)

	var statesToKeepFromChain int
	chainState, err := smT.store.LatestState()
	if err != nil {
		smT.log.Errorf("Cannot get latest chain state: %v", err)
		statesToKeepFromChain = 0
	} else {
		statesToKeepFromChain = int(governance.NewStateAccess(chainState).GetBlockKeepAmount())
	}
	var statesToKeep int
	if statesToKeepFromChain > smT.parameters.PruningMinStatesToKeep {
		statesToKeep = statesToKeepFromChain
	} else {
		statesToKeep = smT.parameters.PruningMinStatesToKeep
	}

	if statesToKeep > smT.chainOfBlocks.Length() {
		return // Number of states in chain is less than `statesToKeep`
	}

	statesToPrune := smT.chainOfBlocks.Length() - statesToKeep
	if statesToPrune > smT.parameters.PruningMaxStatesToDelete {
		statesToPrune = smT.parameters.PruningMaxStatesToDelete
	}
	i := 0
	for ; i < statesToPrune; i++ {
		bi := smT.chainOfBlocks.PeekStart()
		singleStart := time.Now()
		stats, err := smT.store.Prune(bi.trieRoot)
		if err != nil {
			smT.log.Errorf("Failed to prune trie root %s: %v", bi.trieRoot, err)
			return // Returning in order not to leave gaps of pruned trie roots in between not pruned ones
		}
		smT.chainOfBlocks.RemoveStart()
		smT.metrics.StatePruned(time.Since(singleStart), bi.blockIndex)
		smT.log.Debugf("Block index %v %s pruned: %v nodes and %v values deleted", bi.blockIndex, bi.trieRoot, stats.DeletedNodes, stats.DeletedValues)
	}
	smT.metrics.PruningCompleted(time.Since(start), i)
	smT.log.Debugf("Pruning completed, %v trie roots pruned", i)
}

// updateChainOfBlocks updates chain of blocks to contain trie roots/block indexes
// of all the blocks starting from the one with passed commitment and going back
// to the oldest unpruned block. Usually some block chain is currently known.
// However the passed commitment might be newer and not contained in currently
// known chain of blocks. The function attempts to use currently known chain as
// much as possible: while building new chain of blocks, it attempts to find a
// place to merge it with already known chain. After the merge it checks if the
// end of the merged chain is still what it should be.
// This function is extensively tested in `state_manager_gpa_cob_test.go` file.
func (smT *stateManagerGPA) updateChainOfBlocks(commitment *state.L1Commitment, stateIndex uint32) { //nolint:funlen,gocyclo
	GetPreviousBlockInfoFun := func(bi *blockInfo) (*blockInfo, error) {
		block, err := smT.store.BlockByTrieRoot(bi.trieRoot)
		if err != nil {
			smT.log.Errorf("Failed to retrieve previous block info of %s while pruning: %v", bi.trieRoot, err)
			return nil, err
		}
		com := block.PreviousL1Commitment()
		if com == nil {
			return nil, nil
		}
		return &blockInfo{
			trieRoot:   com.TrieRoot(),
			blockIndex: block.StateIndex() - 1,
		}, nil
	}
	GetLastKnownBlockInfoFun := func() *blockInfo {
		if smT.chainOfBlocks.Length() == 0 {
			return nil
		}
		return smT.chainOfBlocks.PeekEnd()
	}

	cob := pipe.NewDeque[*blockInfo]()
	bi := &blockInfo{
		trieRoot:   commitment.TrieRoot(),
		blockIndex: stateIndex,
	}

	var lastKnownBi *blockInfo
	if smT.chainOfBlocks == nil {
		lastKnownBi = nil
	} else {
		lastKnownBi = GetLastKnownBlockInfoFun()
	}

	var err error
	// Find chain of newest blocks: the ones, that has larger indexes than currently known chain
	if lastKnownBi != nil {
		for err == nil && bi != nil && bi.blockIndex > lastKnownBi.blockIndex && smT.store.HasTrieRoot(bi.trieRoot) {
			cob.AddStart(bi)
			bi, err = GetPreviousBlockInfoFun(bi)
		}
	}
	// Remove blocks from currently known chain, that have larger indexes than
	// the newest block: they are older than the newest block, but on the different
	// branch of the chain. TODO: Instead of removing, the blocks should probably
	// be pruned.
	if err == nil && bi != nil {
		for lastKnownBi != nil && lastKnownBi.blockIndex > bi.blockIndex {
			_ = smT.chainOfBlocks.RemoveEnd()
			lastKnownBi = GetLastKnownBlockInfoFun()
		}
	}
	// Try to find a place to merge newest blocks chain with currently known blocks chain: `bi.trieRoot.Equals(lastKnownBi.trieRoot)``
	for err == nil && bi != nil && lastKnownBi != nil && !bi.trieRoot.Equals(lastKnownBi.trieRoot) && smT.store.HasTrieRoot(bi.trieRoot) {
		// Normally, no iteration of this cycle should occur: once a common index
		// is reached in previous cycles, trie roots should also match. In an unlikely
		// event of chain split, each iteration of this cycle fetches one older block
		// to the newest blocks chain and drops (TODO: maybe it should prune) one
		// newest ("last known") block from currently known blocks chain. Hence,
		// this comparison of block indexes should still hold.
		if bi.blockIndex != lastKnownBi.blockIndex {
			smT.log.Errorf("Oldest fetched block index %v does not match newest known block index %v",
				bi.blockIndex, lastKnownBi.blockIndex)
			return
		}
		cob.AddStart(bi)
		bi, err = GetPreviousBlockInfoFun(bi)
		_ = smT.chainOfBlocks.RemoveEnd()
		lastKnownBi = GetLastKnownBlockInfoFun()
	}
	if err != nil {
		smT.log.Errorf("Failed to obtain previous block info: %v", err)
		return
	}
	if lastKnownBi == nil { // either there were no currently known blocks chain,
		// or newest blocks chain had no common block infos
		// (which is very unlikely): fill the chain from the store.
		for err == nil && bi != nil && smT.store.HasTrieRoot(bi.trieRoot) {
			cob.AddStart(bi)
			bi, err = GetPreviousBlockInfoFun(bi)
		}
		if err != nil {
			smT.log.Errorf("Failed to obtain previous block info: %v", err)
			return
		}
		smT.chainOfBlocks = cob
	} else if bi == nil { // origin block has been reached
		smT.chainOfBlocks = cob
	} else if bi.trieRoot.Equals(lastKnownBi.trieRoot) { // Here is the the place to merge newest blocks chain with currently known blocks chain
		// Normally newest blocks chain should contain only several (usually, 1)
		// block infos and currently known blocks chain should contain at least
		// `PruningMinStatesToKeep` block infos, but on a sudden enabling of pruning
		// might contain millions of them. Therefore it is more efficient to copy
		// newest blocks chain to the currently known one compared to doing it
		// the other way round. Let's merge them this way.
		for cob.Length() > 0 {
			bi = cob.RemoveStart()
			smT.chainOfBlocks.AddEnd(bi)
		}
		// Although it should not happen, let's check if the start of the chain
		// is still correct.
		// The oldest known block should still be in the store
		for bi = smT.chainOfBlocks.PeekStart(); !smT.store.HasTrieRoot(bi.trieRoot); bi = smT.chainOfBlocks.PeekStart() {
			// NOTE: `PeekStart` panic on empty list is not handled here, but
			// this should not happen as at least block, passed in function's
			// parameters should be both in this deque and in the store
			_ = smT.chainOfBlocks.RemoveStart()
		}
		// The oldest known block should be the oldest block in the store
		for bi, err = GetPreviousBlockInfoFun(smT.chainOfBlocks.PeekStart()); err == nil && bi != nil && smT.store.HasTrieRoot(bi.trieRoot); bi, err = GetPreviousBlockInfoFun(bi) {
			smT.chainOfBlocks.AddStart(bi)
		}
		if err != nil {
			smT.log.Errorf("Failed to obtain previous block info: %v", err)
			return
		}
	} else { // !smT.store.HasTrieRoot(bi.trieRoot), which means that this block
		// has already been pruned from the store.
		smT.chainOfBlocks = cob
	}
}
