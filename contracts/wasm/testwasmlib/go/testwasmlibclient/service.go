// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package testwasmlibclient

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmclient"
)

const (
	ArgAddress     = "address"
	ArgAgentID     = "agentID"
	ArgBlockIndex  = "blockIndex"
	ArgBool        = "bool"
	ArgBytes       = "bytes"
	ArgChainID     = "chainID"
	ArgColor       = "color"
	ArgHash        = "hash"
	ArgHname       = "hname"
	ArgIndex       = "index"
	ArgInt16       = "int16"
	ArgInt32       = "int32"
	ArgInt64       = "int64"
	ArgInt8        = "int8"
	ArgKey         = "key"
	ArgName        = "name"
	ArgParam       = "this"
	ArgRecordIndex = "recordIndex"
	ArgRequestID   = "requestID"
	ArgString      = "string"
	ArgUint16      = "uint16"
	ArgUint32      = "uint32"
	ArgUint64      = "uint64"
	ArgUint8       = "uint8"
	ArgValue       = "value"

	ResCount  = "count"
	ResIotas  = "iotas"
	ResLength = "length"
	ResRandom = "random"
	ResRecord = "record"
	ResValue  = "value"
)

///////////////////////////// arrayAppend /////////////////////////////

type ArrayAppendFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *ArrayAppendFunc) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *ArrayAppendFunc) Value(v string) {
	f.args.Set(ArgValue, f.args.FromString(v))
}

func (f *ArrayAppendFunc) Post() wasmclient.Request {
	f.args.Mandatory(ArgName)
	f.args.Mandatory(ArgValue)
	return f.ClientFunc.Post(0x612f835f, &f.args)
}

///////////////////////////// arrayClear /////////////////////////////

type ArrayClearFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *ArrayClearFunc) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *ArrayClearFunc) Post() wasmclient.Request {
	f.args.Mandatory(ArgName)
	return f.ClientFunc.Post(0x88021821, &f.args)
}

///////////////////////////// arraySet /////////////////////////////

type ArraySetFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *ArraySetFunc) Index(v uint32) {
	f.args.Set(ArgIndex, f.args.FromUint32(v))
}

func (f *ArraySetFunc) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *ArraySetFunc) Value(v string) {
	f.args.Set(ArgValue, f.args.FromString(v))
}

func (f *ArraySetFunc) Post() wasmclient.Request {
	f.args.Mandatory(ArgIndex)
	f.args.Mandatory(ArgName)
	f.args.Mandatory(ArgValue)
	return f.ClientFunc.Post(0x2c4150b3, &f.args)
}

///////////////////////////// mapClear /////////////////////////////

type MapClearFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *MapClearFunc) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *MapClearFunc) Post() wasmclient.Request {
	f.args.Mandatory(ArgName)
	return f.ClientFunc.Post(0x027f215a, &f.args)
}

///////////////////////////// mapSet /////////////////////////////

type MapSetFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *MapSetFunc) Key(v string) {
	f.args.Set(ArgKey, f.args.FromString(v))
}

func (f *MapSetFunc) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *MapSetFunc) Value(v string) {
	f.args.Set(ArgValue, f.args.FromString(v))
}

func (f *MapSetFunc) Post() wasmclient.Request {
	f.args.Mandatory(ArgKey)
	f.args.Mandatory(ArgName)
	f.args.Mandatory(ArgValue)
	return f.ClientFunc.Post(0xf2260404, &f.args)
}

///////////////////////////// paramTypes /////////////////////////////

type ParamTypesFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *ParamTypesFunc) Address(v wasmclient.Address) {
	f.args.Set(ArgAddress, f.args.FromAddress(v))
}

func (f *ParamTypesFunc) AgentID(v wasmclient.AgentID) {
	f.args.Set(ArgAgentID, f.args.FromAgentID(v))
}

func (f *ParamTypesFunc) Bool(v bool) {
	f.args.Set(ArgBool, f.args.FromBool(v))
}

func (f *ParamTypesFunc) Bytes(v []byte) {
	f.args.Set(ArgBytes, f.args.FromBytes(v))
}

func (f *ParamTypesFunc) ChainID(v wasmclient.ChainID) {
	f.args.Set(ArgChainID, f.args.FromChainID(v))
}

func (f *ParamTypesFunc) Color(v wasmclient.Color) {
	f.args.Set(ArgColor, f.args.FromColor(v))
}

func (f *ParamTypesFunc) Hash(v wasmclient.Hash) {
	f.args.Set(ArgHash, f.args.FromHash(v))
}

func (f *ParamTypesFunc) Hname(v wasmclient.Hname) {
	f.args.Set(ArgHname, f.args.FromHname(v))
}

func (f *ParamTypesFunc) Int16(v int16) {
	f.args.Set(ArgInt16, f.args.FromInt16(v))
}

func (f *ParamTypesFunc) Int32(v int32) {
	f.args.Set(ArgInt32, f.args.FromInt32(v))
}

func (f *ParamTypesFunc) Int64(v int64) {
	f.args.Set(ArgInt64, f.args.FromInt64(v))
}

func (f *ParamTypesFunc) Int8(v int8) {
	f.args.Set(ArgInt8, f.args.FromInt8(v))
}

func (f *ParamTypesFunc) Param(v []byte) {
	f.args.Set(ArgParam, f.args.FromBytes(v))
}

func (f *ParamTypesFunc) RequestID(v wasmclient.RequestID) {
	f.args.Set(ArgRequestID, f.args.FromRequestID(v))
}

func (f *ParamTypesFunc) String(v string) {
	f.args.Set(ArgString, f.args.FromString(v))
}

func (f *ParamTypesFunc) Uint16(v uint16) {
	f.args.Set(ArgUint16, f.args.FromUint16(v))
}

func (f *ParamTypesFunc) Uint32(v uint32) {
	f.args.Set(ArgUint32, f.args.FromUint32(v))
}

func (f *ParamTypesFunc) Uint64(v uint64) {
	f.args.Set(ArgUint64, f.args.FromUint64(v))
}

func (f *ParamTypesFunc) Uint8(v uint8) {
	f.args.Set(ArgUint8, f.args.FromUint8(v))
}

func (f *ParamTypesFunc) Post() wasmclient.Request {
	return f.ClientFunc.Post(0x6921c4cd, &f.args)
}

///////////////////////////// random /////////////////////////////

type RandomFunc struct {
	wasmclient.ClientFunc
}

func (f *RandomFunc) Post() wasmclient.Request {
	return f.ClientFunc.Post(0xe86c97ca, nil)
}

///////////////////////////// triggerEvent /////////////////////////////

type TriggerEventFunc struct {
	wasmclient.ClientFunc
	args wasmclient.Arguments
}

func (f *TriggerEventFunc) Address(v wasmclient.Address) {
	f.args.Set(ArgAddress, f.args.FromAddress(v))
}

func (f *TriggerEventFunc) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *TriggerEventFunc) Post() wasmclient.Request {
	f.args.Mandatory(ArgAddress)
	f.args.Mandatory(ArgName)
	return f.ClientFunc.Post(0xd5438ac6, &f.args)
}

///////////////////////////// arrayLength /////////////////////////////

type ArrayLengthView struct {
	wasmclient.ClientView
	args wasmclient.Arguments
}

func (f *ArrayLengthView) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *ArrayLengthView) Call() ArrayLengthResults {
	f.args.Mandatory(ArgName)
	f.ClientView.Call("arrayLength", &f.args)
	return ArrayLengthResults{res: f.Results()}
}

type ArrayLengthResults struct {
	res wasmclient.Results
}

func (r *ArrayLengthResults) Length() uint32 {
	return r.res.ToUint32(r.res.Get(ResLength))
}

///////////////////////////// arrayValue /////////////////////////////

type ArrayValueView struct {
	wasmclient.ClientView
	args wasmclient.Arguments
}

func (f *ArrayValueView) Index(v uint32) {
	f.args.Set(ArgIndex, f.args.FromUint32(v))
}

func (f *ArrayValueView) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *ArrayValueView) Call() ArrayValueResults {
	f.args.Mandatory(ArgIndex)
	f.args.Mandatory(ArgName)
	f.ClientView.Call("arrayValue", &f.args)
	return ArrayValueResults{res: f.Results()}
}

type ArrayValueResults struct {
	res wasmclient.Results
}

func (r *ArrayValueResults) Value() string {
	return r.res.ToString(r.res.Get(ResValue))
}

///////////////////////////// blockRecord /////////////////////////////

type BlockRecordView struct {
	wasmclient.ClientView
	args wasmclient.Arguments
}

func (f *BlockRecordView) BlockIndex(v uint32) {
	f.args.Set(ArgBlockIndex, f.args.FromUint32(v))
}

func (f *BlockRecordView) RecordIndex(v uint32) {
	f.args.Set(ArgRecordIndex, f.args.FromUint32(v))
}

func (f *BlockRecordView) Call() BlockRecordResults {
	f.args.Mandatory(ArgBlockIndex)
	f.args.Mandatory(ArgRecordIndex)
	f.ClientView.Call("blockRecord", &f.args)
	return BlockRecordResults{res: f.Results()}
}

type BlockRecordResults struct {
	res wasmclient.Results
}

func (r *BlockRecordResults) Record() []byte {
	return r.res.ToBytes(r.res.Get(ResRecord))
}

///////////////////////////// blockRecords /////////////////////////////

type BlockRecordsView struct {
	wasmclient.ClientView
	args wasmclient.Arguments
}

func (f *BlockRecordsView) BlockIndex(v uint32) {
	f.args.Set(ArgBlockIndex, f.args.FromUint32(v))
}

func (f *BlockRecordsView) Call() BlockRecordsResults {
	f.args.Mandatory(ArgBlockIndex)
	f.ClientView.Call("blockRecords", &f.args)
	return BlockRecordsResults{res: f.Results()}
}

type BlockRecordsResults struct {
	res wasmclient.Results
}

func (r *BlockRecordsResults) Count() uint32 {
	return r.res.ToUint32(r.res.Get(ResCount))
}

///////////////////////////// getRandom /////////////////////////////

type GetRandomView struct {
	wasmclient.ClientView
}

func (f *GetRandomView) Call() GetRandomResults {
	f.ClientView.Call("getRandom", nil)
	return GetRandomResults{res: f.Results()}
}

type GetRandomResults struct {
	res wasmclient.Results
}

func (r *GetRandomResults) Random() uint64 {
	return r.res.ToUint64(r.res.Get(ResRandom))
}

///////////////////////////// iotaBalance /////////////////////////////

type IotaBalanceView struct {
	wasmclient.ClientView
}

func (f *IotaBalanceView) Call() IotaBalanceResults {
	f.ClientView.Call("iotaBalance", nil)
	return IotaBalanceResults{res: f.Results()}
}

type IotaBalanceResults struct {
	res wasmclient.Results
}

func (r *IotaBalanceResults) Iotas() uint64 {
	return r.res.ToUint64(r.res.Get(ResIotas))
}

///////////////////////////// mapValue /////////////////////////////

type MapValueView struct {
	wasmclient.ClientView
	args wasmclient.Arguments
}

func (f *MapValueView) Key(v string) {
	f.args.Set(ArgKey, f.args.FromString(v))
}

func (f *MapValueView) Name(v string) {
	f.args.Set(ArgName, f.args.FromString(v))
}

func (f *MapValueView) Call() MapValueResults {
	f.args.Mandatory(ArgKey)
	f.args.Mandatory(ArgName)
	f.ClientView.Call("mapValue", &f.args)
	return MapValueResults{res: f.Results()}
}

type MapValueResults struct {
	res wasmclient.Results
}

func (r *MapValueResults) Value() string {
	return r.res.ToString(r.res.Get(ResValue))
}

///////////////////////////// TestWasmLibService /////////////////////////////

type TestWasmLibService struct {
	wasmclient.Service
}

func NewTestWasmLibService(cl *wasmclient.ServiceClient, chainID string) (*TestWasmLibService, error) {
	s := &TestWasmLibService{}
	err := s.Service.Init(cl, chainID, 0x89703a45)
	return s, err
}

func (s *TestWasmLibService) NewEventHandler() *TestWasmLibEvents {
	return &TestWasmLibEvents{}
}

func (s *TestWasmLibService) ArrayAppend() ArrayAppendFunc {
	return ArrayAppendFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) ArrayClear() ArrayClearFunc {
	return ArrayClearFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) ArraySet() ArraySetFunc {
	return ArraySetFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) MapClear() MapClearFunc {
	return MapClearFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) MapSet() MapSetFunc {
	return MapSetFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) ParamTypes() ParamTypesFunc {
	return ParamTypesFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) Random() RandomFunc {
	return RandomFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) TriggerEvent() TriggerEventFunc {
	return TriggerEventFunc{ClientFunc: s.AsClientFunc()}
}

func (s *TestWasmLibService) ArrayLength() ArrayLengthView {
	return ArrayLengthView{ClientView: s.AsClientView()}
}

func (s *TestWasmLibService) ArrayValue() ArrayValueView {
	return ArrayValueView{ClientView: s.AsClientView()}
}

func (s *TestWasmLibService) BlockRecord() BlockRecordView {
	return BlockRecordView{ClientView: s.AsClientView()}
}

func (s *TestWasmLibService) BlockRecords() BlockRecordsView {
	return BlockRecordsView{ClientView: s.AsClientView()}
}

func (s *TestWasmLibService) GetRandom() GetRandomView {
	return GetRandomView{ClientView: s.AsClientView()}
}

func (s *TestWasmLibService) IotaBalance() IotaBalanceView {
	return IotaBalanceView{ClientView: s.AsClientView()}
}

func (s *TestWasmLibService) MapValue() MapValueView {
	return MapValueView{ClientView: s.AsClientView()}
}