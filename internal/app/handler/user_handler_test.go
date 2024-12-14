package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
	"github.com/patraden/ya-practicum-go-mart/internal/app/handler"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
	"github.com/patraden/ya-practicum-go-mart/internal/app/mock"
)

const mockToken = "mockToken"

type testCase struct {
	name           string
	requestBody    string
	mockUseCase    func(mockUC *mock.MockIUserUseCase)
	mockEncoder    func(_ string) (string, error)
	expectedStatus int
	expectedHeader string
	expectedCookie bool
}

func setupHandler(t *testing.T, mockTokenEncoder jwtauth.TokenEncoder) (
	*gomock.Controller,
	*mock.MockIUserUseCase,
	*handler.UserHandler,
) {
	t.Helper()

	ctrl := gomock.NewController(t)
	log := logger.NewLogger(zerolog.InfoLevel).GetZeroLog()
	mockUseCase := mock.NewMockIUserUseCase(ctrl)
	handler := handler.NewUserHandler(mockUseCase, mockTokenEncoder, log)

	return ctrl, mockUseCase, handler
}

func executeRegisterUserHandlerTest(t *testing.T, testCase testCase) {
	t.Helper()

	ctrl, mockUC, handler := setupHandler(t, testCase.mockEncoder)
	defer ctrl.Finish()

	testCase.mockUseCase(mockUC)

	r := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte(testCase.requestBody)))
	w := httptest.NewRecorder()

	handler.RegisterUser(w, r)
	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, testCase.expectedStatus, res.StatusCode)
	assert.Equal(t, testCase.expectedHeader, res.Header.Get("Authorization"))

	if testCase.expectedCookie {
		cookies := res.Cookies()
		require.NotEmpty(t, cookies)
		assert.Equal(t, jwtauth.JWTCookie, cookies[0].Name)
	}
}

func mockedCreateUser(err error) func(mockUC *mock.MockIUserUseCase) {
	if err == nil {
		return func(mockUC *mock.MockIUserUseCase) {
			mockUC.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(model.NewUser("testuser"), nil)
		}
	}

	return func(mockUC *mock.MockIUserUseCase) {
		mockUC.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil, err)
	}
}

func TestRegisterUserHandler(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:           "successfully registers user",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedCreateUser(nil),
			mockEncoder:    func(_ string) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusOK, expectedHeader: "Bearer " + mockToken, expectedCookie: true,
		},
		{
			name:           "invalid request format",
			requestBody:    `invalid_json`,
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			mockEncoder:    func(_ string) (string, error) { return "", nil },
			expectedStatus: http.StatusBadRequest, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "conflict with existing user",
			requestBody:    `{"username":"existinguser","password":"password123"}`,
			mockUseCase:    mockedCreateUser(e.ErrRepoUserExists),
			mockEncoder:    func(_ string) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusConflict, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during user creation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedCreateUser(e.ErrUseCaseInternal),
			mockEncoder:    func(_ string) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusInternalServerError, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during token generation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			mockEncoder:    func(_ string) (string, error) { return "", e.ErrTesting },
			expectedStatus: http.StatusInternalServerError, expectedHeader: "", expectedCookie: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			executeRegisterUserHandlerTest(t, testCase)
		})
	}
}

func mockedValidateUser(err error) func(mockUC *mock.MockIUserUseCase) {
	if err == nil {
		return func(mockUC *mock.MockIUserUseCase) {
			mockUC.EXPECT().ValidateUser(gomock.Any(), gomock.Any()).Return(nil)
		}
	}

	return func(mockUC *mock.MockIUserUseCase) {
		mockUC.EXPECT().ValidateUser(gomock.Any(), gomock.Any()).Return(err)
	}
}

func executeValidateUserHandlerTest(t *testing.T, testCase testCase) {
	t.Helper()

	ctrl, mockUC, handler := setupHandler(t, testCase.mockEncoder)
	defer ctrl.Finish()

	testCase.mockUseCase(mockUC)

	r := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader([]byte(testCase.requestBody)))
	w := httptest.NewRecorder()

	handler.LoginUser(w, r)
	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, testCase.expectedStatus, res.StatusCode)
	assert.Equal(t, testCase.expectedHeader, res.Header.Get("Authorization"))

	if testCase.expectedCookie {
		cookies := res.Cookies()
		require.NotEmpty(t, cookies)
		assert.Equal(t, jwtauth.JWTCookie, cookies[0].Name)
	}
}

func TestLoginUserHandler(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		{
			name:           "successfully logs in user",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedValidateUser(nil),
			mockEncoder:    func(_ string) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusOK, expectedHeader: "Bearer " + mockToken, expectedCookie: true,
		},
		{
			name:           "invalid request format",
			requestBody:    `invalid_json`,
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			mockEncoder:    func(_ string) (string, error) { return "", nil },
			expectedStatus: http.StatusBadRequest, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "user not found",
			requestBody:    `{"username":"nonexistentuser","password":"password123"}`,
			mockUseCase:    mockedValidateUser(e.ErrRepoUserNotFound),
			mockEncoder:    func(_ string) (string, error) { return "", nil },
			expectedStatus: http.StatusUnauthorized, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "password mismatch",
			requestBody:    `{"username":"testuser","password":"wrongpassword"}`,
			mockUseCase:    mockedValidateUser(e.ErrRepoUserPassMismatch),
			mockEncoder:    func(_ string) (string, error) { return "", nil },
			expectedStatus: http.StatusUnauthorized, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during validation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedValidateUser(e.ErrUseCaseInternal),
			mockEncoder:    func(_ string) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusInternalServerError, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during token generation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			mockEncoder:    func(_ string) (string, error) { return "", e.ErrTesting },
			expectedStatus: http.StatusInternalServerError, expectedHeader: "", expectedCookie: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			executeValidateUserHandlerTest(t, testCase)
		})
	}
}
