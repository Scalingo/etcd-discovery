// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/Scalingo/etcd-discovery/v7/service (interfaces: HostResponse)

// Package servicemock is a generated GoMock package.
package servicemock

import (
	service "github.com/Scalingo/etcd-discovery/v7/service"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockHostResponse is a mock of HostResponse interface
type MockHostResponse struct {
	ctrl     *gomock.Controller
	recorder *MockHostResponseMockRecorder
}

// MockHostResponseMockRecorder is the mock recorder for MockHostResponse
type MockHostResponseMockRecorder struct {
	mock *MockHostResponse
}

// NewMockHostResponse creates a new mock instance
func NewMockHostResponse(ctrl *gomock.Controller) *MockHostResponse {
	mock := &MockHostResponse{ctrl: ctrl}
	mock.recorder = &MockHostResponseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockHostResponse) EXPECT() *MockHostResponseMockRecorder {
	return m.recorder
}

// Err mocks base method
func (m *MockHostResponse) Err() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

// Err indicates an expected call of Err
func (mr *MockHostResponseMockRecorder) Err() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*MockHostResponse)(nil).Err))
}

// Host mocks base method
func (m *MockHostResponse) Host() (*service.Host, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Host")
	ret0, _ := ret[0].(*service.Host)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Host indicates an expected call of Host
func (mr *MockHostResponseMockRecorder) Host() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Host", reflect.TypeOf((*MockHostResponse)(nil).Host))
}

// PrivateURL mocks base method
func (m *MockHostResponse) PrivateURL(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PrivateURL", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PrivateURL indicates an expected call of PrivateURL
func (mr *MockHostResponseMockRecorder) PrivateURL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PrivateURL", reflect.TypeOf((*MockHostResponse)(nil).PrivateURL), arg0, arg1)
}

// URL mocks base method
func (m *MockHostResponse) URL(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "URL", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// URL indicates an expected call of URL
func (mr *MockHostResponseMockRecorder) URL(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "URL", reflect.TypeOf((*MockHostResponse)(nil).URL), arg0, arg1)
}
