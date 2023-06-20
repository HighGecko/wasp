package isc

import (
	"errors"
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type RequestKind rwutil.Kind

const (
	requestKindOnLedger RequestKind = iota
	requestKindOffLedgerISC
	requestKindOffLedgerEVMTx
	requestKindOffLedgerEVMCall
)

func RequestFromBytes(data []byte) (Request, error) {
	rr := rwutil.NewBytesReader(data)
	return RequestFromReader(rr), rr.Err
}

func RequestFromReader(rr *rwutil.Reader) (ret Request) {
	kind := rr.ReadKind()
	switch RequestKind(kind) {
	case requestKindOnLedger:
		ret = new(onLedgerRequestData)
	case requestKindOffLedgerISC:
		ret = new(offLedgerRequestData)
	case requestKindOffLedgerEVMTx:
		ret = new(evmOffLedgerTxRequest)
	case requestKindOffLedgerEVMCall:
		ret = new(evmOffLedgerCallRequest)
	default:
		if rr.Err == nil {
			rr.Err = errors.New("invalid Request kind")
			return nil
		}
	}
	rr.PushBack().WriteKind(kind)
	rr.Read(ret)
	return ret
}

// region RequestID //////////////////////////////////////////////////////////////////

type RequestID iotago.OutputID

const RequestIDDigestLen = 6

type RequestRef struct {
	ID   RequestID
	Hash hashing.HashValue
}

const RequestRefKeyLen = iotago.OutputIDLength + 32

type RequestRefKey [RequestRefKeyLen]byte

func (rrk RequestRefKey) String() string {
	return iotago.EncodeHex(rrk[:])
}

func RequestRefFromRequest(req Request) *RequestRef {
	return &RequestRef{ID: req.ID(), Hash: RequestHash(req)}
}

func RequestRefsFromRequests(reqs []Request) []*RequestRef {
	rr := make([]*RequestRef, len(reqs))
	for i := range rr {
		rr[i] = RequestRefFromRequest(reqs[i])
	}
	return rr
}

func (ref *RequestRef) AsKey() RequestRefKey {
	var key RequestRefKey
	copy(key[:], ref.Bytes())
	return key
}

func (ref *RequestRef) IsFor(req Request) bool {
	if ref.ID != req.ID() {
		return false
	}
	return ref.Hash == RequestHash(req)
}

func (ref *RequestRef) Bytes() []byte {
	return append(ref.Hash[:], ref.ID[:]...)
}

func (ref *RequestRef) String() string {
	return fmt.Sprintf("{requestRef, id=%v, hash=%v}", ref.ID.String(), ref.Hash.Hex())
}

func (ref *RequestRef) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(ref.ID[:])
	rr.ReadN(ref.Hash[:])
	return rr.Err
}

func (ref *RequestRef) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(ref.ID[:])
	ww.WriteN(ref.Hash[:])
	return ww.Err
}

func RequestRefFromBytes(data []byte) (*RequestRef, error) {
	reqID, err := RequestIDFromBytes(data[hashing.HashSize:])
	if err != nil {
		return nil, err
	}
	ret := &RequestRef{
		ID: reqID,
	}
	copy(ret.Hash[:], data[:hashing.HashSize])

	return ret, nil
}

// RequestLookupDigest is shortened version of the request id. It is guaranteed to be unique
// within one block, however it may collide globally. Used for quick checking for most requests
// if it was never seen
type RequestLookupDigest [RequestIDDigestLen + 2]byte

func NewRequestID(txid iotago.TransactionID, index uint16) RequestID {
	return RequestID(iotago.OutputIDFromTransactionIDAndIndex(txid, index))
}

func RequestIDFromBytes(data []byte) (ret RequestID, err error) {
	_, err = rwutil.ReadFromBytes(data, &ret)
	return
}

func RequestIDFromString(s string) (ret RequestID, err error) {
	data, err := iotago.DecodeHex(s)
	if err != nil {
		return RequestID{}, err
	}

	if len(data) != iotago.OutputIDLength {
		return ret, errors.New("error parsing requestID: wrong length")
	}

	requestID := RequestID{}
	copy(requestID[:], data)
	return requestID, nil
}

func (rid RequestID) OutputID() iotago.OutputID {
	return iotago.OutputID(rid)
}

func (rid RequestID) LookupDigest() RequestLookupDigest {
	ret := RequestLookupDigest{}
	copy(ret[:RequestIDDigestLen], rid[:RequestIDDigestLen])
	// last 2 bytes are the outputindex
	copy(ret[RequestIDDigestLen:RequestIDDigestLen+2], rid[len(rid)-2:])
	return ret
}

func (rid RequestID) Bytes() []byte {
	return rid[:]
}

func (rid RequestID) Equals(other RequestID) bool {
	return rid == other
}

func (rid RequestID) String() string {
	return iotago.EncodeHex(rid[:])
}

func (rid RequestID) Short() string {
	ridString := rid.String()
	return fmt.Sprintf("%s..%s", ridString[2:6], ridString[len(ridString)-4:])
}

func (rid *RequestID) Read(r io.Reader) error {
	return rwutil.ReadN(r, rid[:])
}

func (rid *RequestID) Write(w io.Writer) error {
	return rwutil.WriteN(w, rid[:])
}

func ShortRequestIDs(ids []RequestID) []string {
	ret := make([]string, len(ids))
	for i := range ret {
		ret[i] = ids[i].Short()
	}
	return ret
}

func ShortRequestIDsFromRequests(reqs []Request) []string {
	requestIDs := make([]RequestID, len(reqs))
	for i := range reqs {
		requestIDs[i] = reqs[i].ID()
	}
	return ShortRequestIDs(requestIDs)
}

// endregion ////////////////////////////////////////////////////////////

// region RequestMetadata //////////////////////////////////////////////////

type RequestMetadata struct {
	SenderContract Hname `json:"senderContract"`
	// ID of the target smart contract
	TargetContract Hname `json:"targetContract"`
	// entry point code
	EntryPoint Hname `json:"entryPoint"`
	// request arguments
	Params dict.Dict `json:"params"`
	// Allowance intended to the target contract to take. Nil means zero allowance
	Allowance *Assets `json:"allowance"`
	// gas budget
	GasBudget uint64 `json:"gasBudget"`
}

func requestMetadataFromFeatureSet(set iotago.FeatureSet) (*RequestMetadata, error) {
	metadataFeatBlock := set.MetadataFeature()
	if metadataFeatBlock == nil {
		// IMPORTANT: this cannot return an empty `&RequestMetadata{}` object because that could cause `isInternalUTXO` check to fail
		return nil, nil
	}
	return RequestMetadataFromBytes(metadataFeatBlock.Data)
}

func RequestMetadataFromBytes(data []byte) (*RequestMetadata, error) {
	return rwutil.ReadFromBytes(data, new(RequestMetadata))
}

// returns nil if nil pointer receiver is cloned
func (meta *RequestMetadata) Clone() *RequestMetadata {
	if meta == nil {
		return nil
	}

	return &RequestMetadata{
		SenderContract: meta.SenderContract,
		TargetContract: meta.TargetContract,
		EntryPoint:     meta.EntryPoint,
		Params:         meta.Params.Clone(),
		Allowance:      meta.Allowance.Clone(),
		GasBudget:      meta.GasBudget,
	}
}

func (meta *RequestMetadata) Bytes() []byte {
	return rwutil.WriteToBytes(meta)
}

func (meta *RequestMetadata) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.Read(&meta.SenderContract)
	rr.Read(&meta.TargetContract)
	rr.Read(&meta.EntryPoint)
	meta.GasBudget = rr.ReadUint64()
	meta.Params = dict.New()
	rr.Read(&meta.Params)
	meta.Allowance = NewEmptyAssets()
	rr.Read(meta.Allowance)
	return rr.Err
}

func (meta *RequestMetadata) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(&meta.SenderContract)
	ww.Write(&meta.TargetContract)
	ww.Write(&meta.EntryPoint)
	ww.WriteUint64(meta.GasBudget)
	ww.Write(&meta.Params)
	ww.Write(meta.Allowance)
	return ww.Err
}
