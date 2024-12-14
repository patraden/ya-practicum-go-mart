package jwtauth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
	"github.com/patraden/ya-practicum-go-mart/internal/app/logger"
	"github.com/patraden/ya-practicum-go-mart/internal/app/middleware/jwtauth"
)

const (
	jwtSecret = "secret"
	testuser  = "testuser"
)

func mockKeyFunc(_ *jwt.Token) (interface{}, error) {
	return []byte(jwtSecret), nil
}

func ExpiredToken() (string, error) {
	now := time.Now()

	claims := jwtauth.Claims{
		Username: testuser,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return ``, e.ErrTesting
	}

	return signedToken, nil
}

func setupLogger() *zerolog.Logger {
	return logger.NewLogger(zerolog.DebugLevel).GetZeroLog()
}

func TestJWTAuthEncoder(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)

	encoder := auth.Encoder()
	token, err := encoder(testuser)

	require.NoError(t, err, "Token generation should not error")
	assert.NotEmpty(t, token, "Token should be generated successfully")
}

func TestJWTAuthVerifyTokenValid(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)

	encoder := auth.Encoder()
	tokenString, err := encoder(testuser)
	require.NoError(t, err)

	token, err := auth.VerifyToken(tokenString)
	require.NoError(t, err, "Token verification should succeed")
	assert.NotNil(t, token, "Token should not be nil")
	assert.True(t, token.Valid, "Token should be valid")

	claims, ok := token.Claims.(*jwtauth.Claims)
	assert.True(t, ok, "Claims should be of type *middleware.Claims")
	assert.Equal(t, testuser, claims.Username, "Username in claims should match")
}

func TestJWTAuthVerifyTokenInvalid(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)

	// invalid
	token, err := auth.VerifyToken("invalid.token")
	require.Error(t, err, "Verification of an invalid token should error")
	assert.Nil(t, token)

	// empty user
	encoder := auth.Encoder()
	tokenString, err := encoder("")

	require.NoError(t, err)

	token, err = auth.VerifyToken(tokenString)

	require.Error(t, err)
	assert.Nil(t, token)

	// expired
	expiredToken, err := ExpiredToken()
	require.NoError(t, err)
	token, err = auth.VerifyToken(expiredToken)
	require.Error(t, err)
	assert.Nil(t, token)
}

func TestJWTAuthVerifyRequestNoToken(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token, err := auth.VerifyRequest(req)

	require.Error(t, err)
	assert.Nil(t, token)
}

func TestJWTAuthVerifyRequestTokenInHeader(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)
	encoder := auth.Encoder()
	tokenString, err := encoder(testuser)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	token, err := auth.VerifyRequest(req, jwtauth.TokenFromHeader)
	require.NoError(t, err)
	assert.NotNil(t, token)
}

func TestJWTAuthVerifyRequestTokenInCookie(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	jwt := jwtauth.NewJWTAuth(mockKeyFunc, logger)
	encoder := jwt.Encoder()
	tokenString, err := encoder(testuser)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: jwtauth.JWTCookie, Value: tokenString})

	token, err := jwt.VerifyRequest(req, jwtauth.TokenFromCookie)
	require.NoError(t, err)
	assert.NotNil(t, token)
}

func TestVerifierMiddleware(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)
	verifier := jwtauth.Verifier(auth)

	handler := verifier(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, claims, err := jwtauth.FromContext(r.Context())
		// Assert the values injected by Verifier
		if token != nil && claims != nil {
			assert.NoError(t, err)
			assert.NotNil(t, claims)
			w.WriteHeader(http.StatusOK)

			return
		}
		// Invalid or missing tokens
		assert.Error(t, err)
		assert.Nil(t, claims)
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("ValidToken", func(t *testing.T) {
		t.Parallel()

		token, err := auth.Encoder()(testuser)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		req.Header.Set("Authorization", "Bearer "+token)
		handler.ServeHTTP(rec, req)
		resp := rec.Result()

		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NoToken", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		resp := rec.Result()

		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestAuthenticatorMiddleware(t *testing.T) {
	t.Parallel()

	logger := setupLogger()
	auth := jwtauth.NewJWTAuth(mockKeyFunc, logger)
	authenticator := jwtauth.Authenticator(auth)

	handler := authenticator(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("ValidContextToken", func(t *testing.T) {
		tokenString, err := auth.Encoder()(testuser)
		require.NoError(t, err)
		token, err := auth.VerifyToken(tokenString)
		require.NoError(t, err)

		ctx := jwtauth.NewContext(context.Background(), token, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		resp := rec.Result()

		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NoContextToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		resp := rec.Result()

		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
