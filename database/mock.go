package database

import (
	"github.com/stretchr/testify/mock"
)

type MockStore struct {
	mock.Mock
}

func (m *MockStore) Get(ID int) (int, error) {
	returnVals := m.Called(ID)
	return returnVals.Get(0).(int), returnVals.Error(1)
}
