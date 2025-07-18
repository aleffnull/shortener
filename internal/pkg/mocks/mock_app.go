// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/app.go
//
// Generated by this command:
//
//	mockgen -source internal/app/app.go -destination internal/pkg/mocks/mock_app.go -package mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/aleffnull/shortener/models"
	gomock "go.uber.org/mock/gomock"
)

// MockApp is a mock of App interface.
type MockApp struct {
	ctrl     *gomock.Controller
	recorder *MockAppMockRecorder
	isgomock struct{}
}

// MockAppMockRecorder is the mock recorder for MockApp.
type MockAppMockRecorder struct {
	mock *MockApp
}

// NewMockApp creates a new mock instance.
func NewMockApp(ctrl *gomock.Controller) *MockApp {
	mock := &MockApp{ctrl: ctrl}
	mock.recorder = &MockAppMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApp) EXPECT() *MockAppMockRecorder {
	return m.recorder
}

// CheckStore mocks base method.
func (m *MockApp) CheckStore(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckStore", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CheckStore indicates an expected call of CheckStore.
func (mr *MockAppMockRecorder) CheckStore(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckStore", reflect.TypeOf((*MockApp)(nil).CheckStore), arg0)
}

// GetURL mocks base method.
func (m *MockApp) GetURL(arg0 context.Context, arg1 string) (string, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetURL", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetURL indicates an expected call of GetURL.
func (mr *MockAppMockRecorder) GetURL(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetURL", reflect.TypeOf((*MockApp)(nil).GetURL), arg0, arg1)
}

// Init mocks base method.
func (m *MockApp) Init() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init")
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockAppMockRecorder) Init() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockApp)(nil).Init))
}

// ShortenURL mocks base method.
func (m *MockApp) ShortenURL(arg0 context.Context, arg1 *models.ShortenRequest) (*models.ShortenResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShortenURL", arg0, arg1)
	ret0, _ := ret[0].(*models.ShortenResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShortenURL indicates an expected call of ShortenURL.
func (mr *MockAppMockRecorder) ShortenURL(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShortenURL", reflect.TypeOf((*MockApp)(nil).ShortenURL), arg0, arg1)
}

// ShortenURLBatch mocks base method.
func (m *MockApp) ShortenURLBatch(arg0 context.Context, arg1 []*models.ShortenBatchRequestItem) ([]*models.ShortenBatchResponseItem, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ShortenURLBatch", arg0, arg1)
	ret0, _ := ret[0].([]*models.ShortenBatchResponseItem)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ShortenURLBatch indicates an expected call of ShortenURLBatch.
func (mr *MockAppMockRecorder) ShortenURLBatch(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ShortenURLBatch", reflect.TypeOf((*MockApp)(nil).ShortenURLBatch), arg0, arg1)
}

// Shutdown mocks base method.
func (m *MockApp) Shutdown() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shutdown")
}

// Shutdown indicates an expected call of Shutdown.
func (mr *MockAppMockRecorder) Shutdown() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shutdown", reflect.TypeOf((*MockApp)(nil).Shutdown))
}
