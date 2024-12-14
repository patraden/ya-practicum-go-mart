// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/repository/user_repository.go
//
// Generated by this command:
//
//	mockgen -source=internal/app/repository/user_repository.go -destination=internal/app/mock/user_repository.go -package=mock UserRepository
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"

	model "github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

// MockUserRepository is a mock of UserRepository interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
	isgomock struct{}
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// CreateUser mocks base method.
func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, user)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserRepositoryMockRecorder) CreateUser(ctx, user any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserRepository)(nil).CreateUser), ctx, user)
}

// ValidateUser mocks base method.
func (m *MockUserRepository) ValidateUser(ctx context.Context, username, password string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateUser", ctx, username, password)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateUser indicates an expected call of ValidateUser.
func (mr *MockUserRepositoryMockRecorder) ValidateUser(ctx, username, password any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateUser", reflect.TypeOf((*MockUserRepository)(nil).ValidateUser), ctx, username, password)
}
