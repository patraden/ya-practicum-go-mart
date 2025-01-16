package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
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

func setupTransactionsHandler(t *testing.T) (
	*gomock.Controller,
	*mock.MockITransactionsUseCase,
	*handler.TransactionsHandler,
) {
	t.Helper()

	ctrl := gomock.NewController(t)
	log := logger.NewLogger(zerolog.InfoLevel).GetZeroLog()
	mockUseCase := mock.NewMockITransactionsUseCase(ctrl)
	handler := handler.NewTransactionsHandler(mockUseCase, log)

	return ctrl, mockUseCase, handler
}

// User balance handler testing.
type testCaseUserBalance struct {
	name           string
	token          string
	mockUseCase    func(mockUC *mock.MockITransactionsUseCase)
	expectedStatus int
	expectedBody   string
}

func mockedGetUserBalance(balance *model.UserBalance, err error) func(mockUC *mock.MockITransactionsUseCase) {
	return func(mockUC *mock.MockITransactionsUseCase) {
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

	ctrl, mockUC, handler := setupTransactionsHandler(t)
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
			expectedBody:   `{"current":30.55,"withdrawn":10.35}`,
		},
		{
			name:           "unauthorized request due to missing token",
			mockUseCase:    func(_ *mock.MockITransactionsUseCase) {},
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "internal server error on use case failure",
			mockUseCase:    mockedGetUserBalance(nil, e.ErrTesting),
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
