package handler_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
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
	mockEncoder    func(string, uuid.UUID) (string, error)
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

// RegisterUser handler test

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
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusOK, expectedHeader: "Bearer " + mockToken, expectedCookie: true,
		},
		{
			name:           "invalid request format",
			requestBody:    `invalid_json`,
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return "", nil },
			expectedStatus: http.StatusBadRequest, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "conflict with existing user",
			requestBody:    `{"username":"existinguser","password":"password123"}`,
			mockUseCase:    mockedCreateUser(e.ErrRepoUserExists),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusConflict, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during user creation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedCreateUser(e.ErrUseCaseInternal),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusInternalServerError, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during token generation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedCreateUser(nil),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return "", e.ErrTesting },
			expectedStatus: http.StatusOK, expectedHeader: "", expectedCookie: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			executeRegisterUserHandlerTest(t, testCase)
		})
	}
}

// LoginUser handler test

func mockedValidateUser(err error) func(mockUC *mock.MockIUserUseCase) {
	if err == nil {
		return func(mockUC *mock.MockIUserUseCase) {
			mockUC.EXPECT().ValidateUser(gomock.Any(), gomock.Any()).Return(model.NewUser("testuser"), nil)
		}
	}

	return func(mockUC *mock.MockIUserUseCase) {
		mockUC.EXPECT().ValidateUser(gomock.Any(), gomock.Any()).Return(nil, err)
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
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusOK, expectedHeader: "Bearer " + mockToken, expectedCookie: true,
		},
		{
			name:           "invalid request format",
			requestBody:    `invalid_json`,
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return "", nil },
			expectedStatus: http.StatusBadRequest, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "user not found",
			requestBody:    `{"username":"nonexistentuser","password":"password123"}`,
			mockUseCase:    mockedValidateUser(e.ErrRepoUserNotFound),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return "", nil },
			expectedStatus: http.StatusUnauthorized, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "password mismatch",
			requestBody:    `{"username":"testuser","password":"wrongpassword"}`,
			mockUseCase:    mockedValidateUser(e.ErrRepoUserPassMismatch),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return "", nil },
			expectedStatus: http.StatusUnauthorized, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during validation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedValidateUser(e.ErrUseCaseInternal),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return mockToken, nil },
			expectedStatus: http.StatusInternalServerError, expectedHeader: "", expectedCookie: false,
		},
		{
			name:           "internal server error during token generation",
			requestBody:    `{"username":"testuser","password":"password123"}`,
			mockUseCase:    mockedValidateUser(nil),
			mockEncoder:    func(_ string, _ uuid.UUID) (string, error) { return "", e.ErrTesting },
			expectedStatus: http.StatusOK, expectedHeader: "", expectedCookie: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			executeValidateUserHandlerTest(t, testCase)
		})
	}
}

// User balance handler testing.
type testCaseUserBalance struct {
	name           string
	token          string
	mockUseCase    func(mockUC *mock.MockIUserUseCase)
	expectedStatus int
	expectedBody   string
}

func mockedGetUserBalance(balance *model.UserBalance, err error) func(mockUC *mock.MockIUserUseCase) {
	return func(mockUC *mock.MockIUserUseCase) {
		mockUC.EXPECT().GetUserBalance(gomock.Any(), gomock.Any()).Return(balance, err)
	}
}

func mockedToken(t *testing.T, user *model.User) (string, *jwtauth.JWTAuth) {
	t.Helper()

	logger := logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
	mockKeyFunc := func(_ *jwt.Token) (interface{}, error) { return []byte("jwtSecret"), nil }
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)
	encoder := auth.Encoder()

	token, err := encoder(user.Username, user.ID)
	require.NoError(t, err)

	return token, auth
}

func executeUserBalanceHandlerTest(t *testing.T, testCase testCaseUserBalance, auth *jwtauth.JWTAuth) {
	t.Helper()

	ctrl, mockUC, handler := setupHandler(t, nil)
	defer ctrl.Finish()

	testCase.mockUseCase(mockUC)

	r := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+testCase.token)

	verifier, authenticator := jwtauth.Verifier(auth), jwtauth.Authenticator()
	verifier(authenticator(http.HandlerFunc(handler.UserBalance))).ServeHTTP(w, r)
	res := w.Result()

	defer res.Body.Close()

	assert.Equal(t, testCase.expectedStatus, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	if testCase.expectedBody != `` {
		assert.JSONEq(t, testCase.expectedBody, string(body))
	}
}

func TestUserBalanceHandler(t *testing.T) {
	t.Parallel()

	user := model.NewUser("testuser")
	token, auth := mockedToken(t, user)
	balance := model.NewUserBalance(user.ID)
	require.NoError(t, balance.Accrual(decimal.NewFromFloat(30.55)))
	require.NoError(t, balance.Withdraw(decimal.NewFromFloat(10.35)))

	tests := []testCaseUserBalance{
		{
			name:           "successfully retrieves user balance",
			mockUseCase:    mockedGetUserBalance(balance, nil),
			token:          token,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"current":"30.55","withdrawn":"10.35"}`,
		},
		{
			name:           "unauthorized request due to missing token",
			mockUseCase:    func(_ *mock.MockIUserUseCase) {},
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "internal server error on use case failure",
			mockUseCase:    mockedGetUserBalance(nil, e.ErrUseCaseInternal),
			token:          token,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			executeUserBalanceHandlerTest(t, testCase, auth)
		})
	}
}
