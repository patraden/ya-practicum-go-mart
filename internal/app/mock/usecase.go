// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/usecase/usecase.go
//
// Generated by this command:
//
//	mockgen -source=internal/app/usecase/usecase.go -destination=internal/app/mock/usecase.go -package=mock IUserUseCase
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	model "github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	dto "github.com/patraden/ya-practicum-go-mart/internal/app/dto"
	gomock "go.uber.org/mock/gomock"
)

// MockIUserUseCase is a mock of IUserUseCase interface.
type MockIUserUseCase struct {
	ctrl     *gomock.Controller
	recorder *MockIUserUseCaseMockRecorder
	isgomock struct{}
}

// MockIUserUseCaseMockRecorder is the mock recorder for MockIUserUseCase.
type MockIUserUseCaseMockRecorder struct {
	mock *MockIUserUseCase
}

// NewMockIUserUseCase creates a new mock instance.
func NewMockIUserUseCase(ctrl *gomock.Controller) *MockIUserUseCase {
	mock := &MockIUserUseCase{ctrl: ctrl}
	mock.recorder = &MockIUserUseCaseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIUserUseCase) EXPECT() *MockIUserUseCaseMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockIUserUseCase) CreateUser(ctx context.Context, creds *dto.UserCredentials) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, creds)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockIUserUseCaseMockRecorder) CreateUser(ctx, creds any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockIUserUseCase)(nil).CreateUser), ctx, creds)
}

// ValidateUser mocks base method.
func (m *MockIUserUseCase) ValidateUser(ctx context.Context, creds *dto.UserCredentials) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateUser", ctx, creds)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateUser indicates an expected call of ValidateUser.
func (mr *MockIUserUseCaseMockRecorder) ValidateUser(ctx, creds any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateUser", reflect.TypeOf((*MockIUserUseCase)(nil).ValidateUser), ctx, creds)
}
