package testutils

import (
	"github.com/aleffnull/shortener/internal/pkg/mocks"
	"go.uber.org/mock/gomock"
)

type Mock struct {
	App       *mocks.MockApp
	Store     *mocks.MockStore
	ColdStore *mocks.MockColdStore
	Logger    *mocks.MockLogger
}

func NewMock(ctrl *gomock.Controller) *Mock {
	return &Mock{
		App:       mocks.NewMockApp(ctrl),
		Store:     mocks.NewMockStore(ctrl),
		ColdStore: mocks.NewMockColdStore(ctrl),
		Logger:    mocks.NewMockLogger(ctrl),
	}
}
