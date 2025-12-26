package mocks

import (
	"go.uber.org/mock/gomock"
)

type Mock struct {
	App                  *MockApp
	AppParameters        *MockAppParameters
	DeleteURLsService    *MockDeleteURLsService
	AuditService         *MockAuditService
	AuthorizationService *MockAuthorizationService
	Connection           *MockConnection
	Store                *MockStore
	ColdStore            *MockColdStore
	Logger               *MockLogger
}

func NewMock(ctrl *gomock.Controller) *Mock {
	return &Mock{
		App:                  NewMockApp(ctrl),
		AppParameters:        NewMockAppParameters(ctrl),
		DeleteURLsService:    NewMockDeleteURLsService(ctrl),
		AuditService:         NewMockAuditService(ctrl),
		AuthorizationService: NewMockAuthorizationService(ctrl),
		Connection:           NewMockConnection(ctrl),
		Store:                NewMockStore(ctrl),
		ColdStore:            NewMockColdStore(ctrl),
		Logger:               NewMockLogger(ctrl),
	}
}
