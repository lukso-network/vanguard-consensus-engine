// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/prysmaticlabs/prysm/proto/eth/v1alpha1 (interfaces: BeaconChain_StreamChainHeadServer,BeaconChain_StreamAttestationsServer,BeaconChain_StreamBlocksServer,BeaconChain_StreamValidatorsInfoServer,BeaconChain_StreamIndexedAttestationsServer,BeaconChain_StreamMinimalConsensusInfoServer,BeaconChain_StreamNewPendingBlocksServer)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	metadata "google.golang.org/grpc/metadata"
)

// MockBeaconChain_StreamChainHeadServer is a mock of BeaconChain_StreamChainHeadServer interface
type MockBeaconChain_StreamChainHeadServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamChainHeadServerMockRecorder
}

// MockBeaconChain_StreamChainHeadServerMockRecorder is the mock recorder for MockBeaconChain_StreamChainHeadServer
type MockBeaconChain_StreamChainHeadServerMockRecorder struct {
	mock *MockBeaconChain_StreamChainHeadServer
}

// NewMockBeaconChain_StreamChainHeadServer creates a new mock instance
func NewMockBeaconChain_StreamChainHeadServer(ctrl *gomock.Controller) *MockBeaconChain_StreamChainHeadServer {
	mock := &MockBeaconChain_StreamChainHeadServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamChainHeadServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamChainHeadServer) EXPECT() *MockBeaconChain_StreamChainHeadServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) Send(arg0 *eth.ChainHead) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamChainHeadServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamChainHeadServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamChainHeadServer)(nil).SetTrailer), arg0)
}

// MockBeaconChain_StreamAttestationsServer is a mock of BeaconChain_StreamAttestationsServer interface
type MockBeaconChain_StreamAttestationsServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamAttestationsServerMockRecorder
}

// MockBeaconChain_StreamAttestationsServerMockRecorder is the mock recorder for MockBeaconChain_StreamAttestationsServer
type MockBeaconChain_StreamAttestationsServerMockRecorder struct {
	mock *MockBeaconChain_StreamAttestationsServer
}

// NewMockBeaconChain_StreamAttestationsServer creates a new mock instance
func NewMockBeaconChain_StreamAttestationsServer(ctrl *gomock.Controller) *MockBeaconChain_StreamAttestationsServer {
	mock := &MockBeaconChain_StreamAttestationsServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamAttestationsServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamAttestationsServer) EXPECT() *MockBeaconChain_StreamAttestationsServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) Send(arg0 *eth.Attestation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamAttestationsServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamAttestationsServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamAttestationsServer)(nil).SetTrailer), arg0)
}

// MockBeaconChain_StreamBlocksServer is a mock of BeaconChain_StreamBlocksServer interface
type MockBeaconChain_StreamBlocksServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamBlocksServerMockRecorder
}

// MockBeaconChain_StreamBlocksServerMockRecorder is the mock recorder for MockBeaconChain_StreamBlocksServer
type MockBeaconChain_StreamBlocksServerMockRecorder struct {
	mock *MockBeaconChain_StreamBlocksServer
}

// NewMockBeaconChain_StreamBlocksServer creates a new mock instance
func NewMockBeaconChain_StreamBlocksServer(ctrl *gomock.Controller) *MockBeaconChain_StreamBlocksServer {
	mock := &MockBeaconChain_StreamBlocksServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamBlocksServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamBlocksServer) EXPECT() *MockBeaconChain_StreamBlocksServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamBlocksServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamBlocksServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamBlocksServer) Send(arg0 *eth.SignedBeaconBlock) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamBlocksServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamBlocksServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamBlocksServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamBlocksServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamBlocksServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamBlocksServer)(nil).SetTrailer), arg0)
}

// MockBeaconChain_StreamValidatorsInfoServer is a mock of BeaconChain_StreamValidatorsInfoServer interface
type MockBeaconChain_StreamValidatorsInfoServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamValidatorsInfoServerMockRecorder
}

// MockBeaconChain_StreamValidatorsInfoServerMockRecorder is the mock recorder for MockBeaconChain_StreamValidatorsInfoServer
type MockBeaconChain_StreamValidatorsInfoServerMockRecorder struct {
	mock *MockBeaconChain_StreamValidatorsInfoServer
}

// NewMockBeaconChain_StreamValidatorsInfoServer creates a new mock instance
func NewMockBeaconChain_StreamValidatorsInfoServer(ctrl *gomock.Controller) *MockBeaconChain_StreamValidatorsInfoServer {
	mock := &MockBeaconChain_StreamValidatorsInfoServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamValidatorsInfoServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamValidatorsInfoServer) EXPECT() *MockBeaconChain_StreamValidatorsInfoServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).Context))
}

// Recv mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) Recv() (*eth.ValidatorChangeSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Recv")
	ret0, _ := ret[0].(*eth.ValidatorChangeSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Recv indicates an expected call of Recv
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) Recv() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Recv", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).Recv))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) Send(arg0 *eth.ValidatorInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamValidatorsInfoServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamValidatorsInfoServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamValidatorsInfoServer)(nil).SetTrailer), arg0)
}

// MockBeaconChain_StreamIndexedAttestationsServer is a mock of BeaconChain_StreamIndexedAttestationsServer interface
type MockBeaconChain_StreamIndexedAttestationsServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder
}

// MockBeaconChain_StreamIndexedAttestationsServerMockRecorder is the mock recorder for MockBeaconChain_StreamIndexedAttestationsServer
type MockBeaconChain_StreamIndexedAttestationsServerMockRecorder struct {
	mock *MockBeaconChain_StreamIndexedAttestationsServer
}

// NewMockBeaconChain_StreamIndexedAttestationsServer creates a new mock instance
func NewMockBeaconChain_StreamIndexedAttestationsServer(ctrl *gomock.Controller) *MockBeaconChain_StreamIndexedAttestationsServer {
	mock := &MockBeaconChain_StreamIndexedAttestationsServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamIndexedAttestationsServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamIndexedAttestationsServer) EXPECT() *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) Send(arg0 *eth.IndexedAttestation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamIndexedAttestationsServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamIndexedAttestationsServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamIndexedAttestationsServer)(nil).SetTrailer), arg0)
}

// MockBeaconChain_StreamMinimalConsensusInfoServer is a mock of BeaconChain_StreamMinimalConsensusInfoServer interface
type MockBeaconChain_StreamMinimalConsensusInfoServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder
}

// MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder is the mock recorder for MockBeaconChain_StreamMinimalConsensusInfoServer
type MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder struct {
	mock *MockBeaconChain_StreamMinimalConsensusInfoServer
}

// NewMockBeaconChain_StreamMinimalConsensusInfoServer creates a new mock instance
func NewMockBeaconChain_StreamMinimalConsensusInfoServer(ctrl *gomock.Controller) *MockBeaconChain_StreamMinimalConsensusInfoServer {
	mock := &MockBeaconChain_StreamMinimalConsensusInfoServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) EXPECT() *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) Send(arg0 *eth.MinimalConsensusInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamMinimalConsensusInfoServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamMinimalConsensusInfoServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamMinimalConsensusInfoServer)(nil).SetTrailer), arg0)
}

// MockBeaconChain_StreamNewPendingBlocksServer is a mock of BeaconChain_StreamNewPendingBlocksServer interface
type MockBeaconChain_StreamNewPendingBlocksServer struct {
	ctrl     *gomock.Controller
	recorder *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder
}

// MockBeaconChain_StreamNewPendingBlocksServerMockRecorder is the mock recorder for MockBeaconChain_StreamNewPendingBlocksServer
type MockBeaconChain_StreamNewPendingBlocksServerMockRecorder struct {
	mock *MockBeaconChain_StreamNewPendingBlocksServer
}

// NewMockBeaconChain_StreamNewPendingBlocksServer creates a new mock instance
func NewMockBeaconChain_StreamNewPendingBlocksServer(ctrl *gomock.Controller) *MockBeaconChain_StreamNewPendingBlocksServer {
	mock := &MockBeaconChain_StreamNewPendingBlocksServer{ctrl: ctrl}
	mock.recorder = &MockBeaconChain_StreamNewPendingBlocksServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockBeaconChain_StreamNewPendingBlocksServer) EXPECT() *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder {
	return m.recorder
}

// Context mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) Context() context.Context {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Context")
	ret0, _ := ret[0].(context.Context)
	return ret0
}

// Context indicates an expected call of Context
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) Context() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Context", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).Context))
}

// RecvMsg mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) RecvMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecvMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RecvMsg indicates an expected call of RecvMsg
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) RecvMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecvMsg", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).RecvMsg), arg0)
}

// Send mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) Send(arg0 *eth.BeaconBlock) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) Send(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).Send), arg0)
}

// SendHeader mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) SendHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendHeader indicates an expected call of SendHeader
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) SendHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendHeader", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).SendHeader), arg0)
}

// SendMsg mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) SendMsg(arg0 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMsg", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMsg indicates an expected call of SendMsg
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) SendMsg(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMsg", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).SendMsg), arg0)
}

// SetHeader mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) SetHeader(arg0 metadata.MD) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetHeader", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetHeader indicates an expected call of SetHeader
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) SetHeader(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetHeader", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).SetHeader), arg0)
}

// SetTrailer mocks base method
func (m *MockBeaconChain_StreamNewPendingBlocksServer) SetTrailer(arg0 metadata.MD) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTrailer", arg0)
}

// SetTrailer indicates an expected call of SetTrailer
func (mr *MockBeaconChain_StreamNewPendingBlocksServerMockRecorder) SetTrailer(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTrailer", reflect.TypeOf((*MockBeaconChain_StreamNewPendingBlocksServer)(nil).SetTrailer), arg0)
}
