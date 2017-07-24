package messages_test

import (
	"fmt"
	"testing"

	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"

	logrus "github.com/FactomProject/logrus"
	gomock "github.com/golang/mock/gomock"
)

// Using this allows us to set the return values of the IMsg interface and control error conditions
// Here is an example setting errors on GetSignature() and MarshalForSignature(), things that are hard
// to usually set
func TestVerifyAndSign(t *testing.T) {
	var err error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := NewMockSignable(ctrl)

	// Setup Mock for:
	//		Good marshal
	//		Bad get
	m.EXPECT().IsValid().AnyTimes().Return(false)
	m.EXPECT().GetSignature().Times(1).Return(nil)
	m.EXPECT().MarshalForSignature().Return([]byte{}, nil)
	//	m.EXPECT().MarshalForSignature().Return([]byte{0x00}, nil)

	_, err = VerifyMessage(m)
	if err == nil {
		t.Error("Should error as MarshalForSignature should fail")
	}

	// Setup Mock for:
	//		bad marshal
	//		good get
	// Fail on marshal sig

	m.EXPECT().GetSignature().AnyTimes().Return(new(primitives.Signature))
	m.EXPECT().MarshalForSignature().AnyTimes().Return(nil, fmt.Errorf("Mock defined error"))
	_, err = SignSignable(m, nil)
	if err == nil {
		t.Error("Should error as MarshalForSignature should fail")
	}

	_, err = VerifyMessage(m)
	if err == nil {
		t.Error("Should error as MarshalForSignature should fail")
	}
}

// MOCK STRUCTURE
// Mock of Signable interface
type MockSignable struct {
	ctrl     *gomock.Controller
	recorder *_MockSignableRecorder
}

// Recorder for MockSignable (not exported)
type _MockSignableRecorder struct {
	mock *MockSignable
}

func NewMockSignable(ctrl *gomock.Controller) *MockSignable {
	mock := &MockSignable{ctrl: ctrl}
	mock.recorder = &_MockSignableRecorder{mock}
	return mock
}

func (_m *MockSignable) EXPECT() *_MockSignableRecorder {
	return _m.recorder
}

func (_m *MockSignable) Sign(_param0 Signer) error {
	ret := _m.ctrl.Call(_m, "Sign", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSignableRecorder) Sign(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Sign", arg0)
}

func (_m *MockSignable) MarshalForSignature() ([]byte, error) {
	ret := _m.ctrl.Call(_m, "MarshalForSignature")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockSignableRecorder) MarshalForSignature() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "MarshalForSignature")
}

func (_m *MockSignable) GetSignature() IFullSignature {
	ret := _m.ctrl.Call(_m, "GetSignature")
	ret0, _ := ret[0].(IFullSignature)
	return ret0
}

func (_mr *_MockSignableRecorder) GetSignature() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetSignature")
}

func (_m *MockSignable) IsValid() bool {
	ret := _m.ctrl.Call(_m, "IsValid")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockSignableRecorder) IsValid() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsValid")
}

func (_m *MockSignable) SetValid() {
	_m.ctrl.Call(_m, "SetValid")
}

func (_mr *_MockSignableRecorder) SetValid() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetValid")
}

// Mock of IMsg interface
type MockIMsg struct {
	ctrl     *gomock.Controller
	recorder *_MockIMsgRecorder
}

// Recorder for MockIMsg (not exported)
type _MockIMsgRecorder struct {
	mock *MockIMsg
}

func NewMockIMsg(ctrl *gomock.Controller) *MockIMsg {
	mock := &MockIMsg{ctrl: ctrl}
	mock.recorder = &_MockIMsgRecorder{mock}
	return mock
}

func (_m *MockIMsg) EXPECT() *_MockIMsgRecorder {
	return _m.recorder
}

func (_m *MockIMsg) JSONByte() ([]byte, error) {
	ret := _m.ctrl.Call(_m, "JSONByte")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockIMsgRecorder) JSONByte() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "JSONByte")
}

func (_m *MockIMsg) JSONString() (string, error) {
	ret := _m.ctrl.Call(_m, "JSONString")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockIMsgRecorder) JSONString() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "JSONString")
}

func (_m *MockIMsg) String() string {
	ret := _m.ctrl.Call(_m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockIMsgRecorder) String() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "String")
}

func (_m *MockIMsg) MarshalBinary() ([]byte, error) {
	ret := _m.ctrl.Call(_m, "MarshalBinary")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockIMsgRecorder) MarshalBinary() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "MarshalBinary")
}

func (_m *MockIMsg) UnmarshalBinaryData(data []byte) ([]byte, error) {
	ret := _m.ctrl.Call(_m, "UnmarshalBinaryData", data)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockIMsgRecorder) UnmarshalBinaryData(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "UnmarshalBinaryData", arg0)
}

func (_m *MockIMsg) UnarmshalBinary(data []byte) error {
	ret := _m.ctrl.Call(_m, "UnarmshalBinary", data)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockIMsgRecorder) UnarmshalBinary(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "UnarmshalBinary", arg0)
}

func (_m *MockIMsg) GetAck() IMsg {
	ret := _m.ctrl.Call(_m, "GetAck")
	ret0, _ := ret[0].(IMsg)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetAck() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetAck")
}

func (_m *MockIMsg) PutAck(_param0 IMsg) {
	_m.ctrl.Call(_m, "PutAck", _param0)
}

func (_mr *_MockIMsgRecorder) PutAck(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PutAck", arg0)
}

func (_m *MockIMsg) Type() byte {
	ret := _m.ctrl.Call(_m, "Type")
	ret0, _ := ret[0].(byte)
	return ret0
}

func (_mr *_MockIMsgRecorder) Type() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Type")
}

func (_m *MockIMsg) IsLocal() bool {
	ret := _m.ctrl.Call(_m, "IsLocal")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) IsLocal() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsLocal")
}

func (_m *MockIMsg) SetLocal(_param0 bool) {
	_m.ctrl.Call(_m, "SetLocal", _param0)
}

func (_mr *_MockIMsgRecorder) SetLocal(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetLocal", arg0)
}

func (_m *MockIMsg) GetOrigin() int {
	ret := _m.ctrl.Call(_m, "GetOrigin")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetOrigin() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetOrigin")
}

func (_m *MockIMsg) SetOrigin(_param0 int) {
	_m.ctrl.Call(_m, "SetOrigin", _param0)
}

func (_mr *_MockIMsgRecorder) SetOrigin(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetOrigin", arg0)
}

func (_m *MockIMsg) GetNetworkOrigin() string {
	ret := _m.ctrl.Call(_m, "GetNetworkOrigin")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetNetworkOrigin() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetNetworkOrigin")
}

func (_m *MockIMsg) SetNetworkOrigin(_param0 string) {
	_m.ctrl.Call(_m, "SetNetworkOrigin", _param0)
}

func (_mr *_MockIMsgRecorder) SetNetworkOrigin(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetNetworkOrigin", arg0)
}

func (_m *MockIMsg) GetTimestamp() Timestamp {
	ret := _m.ctrl.Call(_m, "GetTimestamp")
	ret0, _ := ret[0].(Timestamp)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetTimestamp() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetTimestamp")
}

func (_m *MockIMsg) GetRepeatHash() IHash {
	ret := _m.ctrl.Call(_m, "GetRepeatHash")
	ret0, _ := ret[0].(IHash)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetRepeatHash() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetRepeatHash")
}

func (_m *MockIMsg) GetHash() IHash {
	ret := _m.ctrl.Call(_m, "GetHash")
	ret0, _ := ret[0].(IHash)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetHash() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetHash")
}

func (_m *MockIMsg) GetMsgHash() IHash {
	ret := _m.ctrl.Call(_m, "GetMsgHash")
	ret0, _ := ret[0].(IHash)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetMsgHash() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetMsgHash")
}

func (_m *MockIMsg) GetFullMsgHash() IHash {
	ret := _m.ctrl.Call(_m, "GetFullMsgHash")
	ret0, _ := ret[0].(IHash)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetFullMsgHash() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFullMsgHash")
}

func (_m *MockIMsg) IsPeer2Peer() bool {
	ret := _m.ctrl.Call(_m, "IsPeer2Peer")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) IsPeer2Peer() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsPeer2Peer")
}

func (_m *MockIMsg) SetPeer2Peer(_param0 bool) {
	_m.ctrl.Call(_m, "SetPeer2Peer", _param0)
}

func (_mr *_MockIMsgRecorder) SetPeer2Peer(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetPeer2Peer", arg0)
}

func (_m *MockIMsg) Validate(_param0 IState) int {
	ret := _m.ctrl.Call(_m, "Validate", _param0)
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockIMsgRecorder) Validate(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Validate", arg0)
}

func (_m *MockIMsg) ComputeVMIndex(_param0 IState) {
	_m.ctrl.Call(_m, "ComputeVMIndex", _param0)
}

func (_mr *_MockIMsgRecorder) ComputeVMIndex(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ComputeVMIndex", arg0)
}

func (_m *MockIMsg) LeaderExecute(_param0 IState) {
	_m.ctrl.Call(_m, "LeaderExecute", _param0)
}

func (_mr *_MockIMsgRecorder) LeaderExecute(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LeaderExecute", arg0)
}

func (_m *MockIMsg) GetLeaderChainID() IHash {
	ret := _m.ctrl.Call(_m, "GetLeaderChainID")
	ret0, _ := ret[0].(IHash)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetLeaderChainID() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetLeaderChainID")
}

func (_m *MockIMsg) SetLeaderChainID(_param0 IHash) {
	_m.ctrl.Call(_m, "SetLeaderChainID", _param0)
}

func (_mr *_MockIMsgRecorder) SetLeaderChainID(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetLeaderChainID", arg0)
}

func (_m *MockIMsg) FollowerExecute(_param0 IState) {
	_m.ctrl.Call(_m, "FollowerExecute", _param0)
}

func (_mr *_MockIMsgRecorder) FollowerExecute(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "FollowerExecute", arg0)
}

func (_m *MockIMsg) SendOut(_param0 IState, _param1 IMsg) {
	_m.ctrl.Call(_m, "SendOut", _param0, _param1)
}

func (_mr *_MockIMsgRecorder) SendOut(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SendOut", arg0, arg1)
}

func (_m *MockIMsg) GetNoResend() bool {
	ret := _m.ctrl.Call(_m, "GetNoResend")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetNoResend() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetNoResend")
}

func (_m *MockIMsg) SetNoResend(_param0 bool) {
	_m.ctrl.Call(_m, "SetNoResend", _param0)
}

func (_mr *_MockIMsgRecorder) SetNoResend(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetNoResend", arg0)
}

func (_m *MockIMsg) Process(dbheight uint32, state IState) bool {
	ret := _m.ctrl.Call(_m, "Process", dbheight, state)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) Process(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Process", arg0, arg1)
}

func (_m *MockIMsg) GetVMIndex() int {
	ret := _m.ctrl.Call(_m, "GetVMIndex")
	ret0, _ := ret[0].(int)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetVMIndex() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetVMIndex")
}

func (_m *MockIMsg) SetVMIndex(_param0 int) {
	_m.ctrl.Call(_m, "SetVMIndex", _param0)
}

func (_mr *_MockIMsgRecorder) SetVMIndex(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetVMIndex", arg0)
}

func (_m *MockIMsg) GetVMHash() []byte {
	ret := _m.ctrl.Call(_m, "GetVMHash")
	ret0, _ := ret[0].([]byte)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetVMHash() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetVMHash")
}

func (_m *MockIMsg) SetVMHash(_param0 []byte) {
	_m.ctrl.Call(_m, "SetVMHash", _param0)
}

func (_mr *_MockIMsgRecorder) SetVMHash(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetVMHash", arg0)
}

func (_m *MockIMsg) GetMinute() byte {
	ret := _m.ctrl.Call(_m, "GetMinute")
	ret0, _ := ret[0].(byte)
	return ret0
}

func (_mr *_MockIMsgRecorder) GetMinute() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetMinute")
}

func (_m *MockIMsg) SetMinute(_param0 byte) {
	_m.ctrl.Call(_m, "SetMinute", _param0)
}

func (_mr *_MockIMsgRecorder) SetMinute(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetMinute", arg0)
}

func (_m *MockIMsg) MarkSentInvalid(_param0 bool) {
	_m.ctrl.Call(_m, "MarkSentInvalid", _param0)
}

func (_mr *_MockIMsgRecorder) MarkSentInvalid(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "MarkSentInvalid", arg0)
}

func (_m *MockIMsg) SentInvalid() bool {
	ret := _m.ctrl.Call(_m, "SentInvalid")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) SentInvalid() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SentInvalid")
}

func (_m *MockIMsg) IsStalled() bool {
	ret := _m.ctrl.Call(_m, "IsStalled")
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) IsStalled() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "IsStalled")
}

func (_m *MockIMsg) SetStall(_param0 bool) {
	_m.ctrl.Call(_m, "SetStall", _param0)
}

func (_mr *_MockIMsgRecorder) SetStall(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SetStall", arg0)
}

func (_m *MockIMsg) Resend(_param0 IState) bool {
	ret := _m.ctrl.Call(_m, "Resend", _param0)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) Resend(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Resend", arg0)
}

func (_m *MockIMsg) Expire(_param0 IState) bool {
	ret := _m.ctrl.Call(_m, "Expire", _param0)
	ret0, _ := ret[0].(bool)
	return ret0
}

func (_mr *_MockIMsgRecorder) Expire(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Expire", arg0)
}

func (_m *MockIMsg) LogFields() logrus.Fields {
	ret := _m.ctrl.Call(_m, "LogFields")
	ret0, _ := ret[0].(logrus.Fields)
	return ret0
}

func (_mr *_MockIMsgRecorder) LogFields() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "LogFields")
}
