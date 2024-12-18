package jwtauth

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
)

type TokenEncoder func(string, uuid.UUID) (string, error)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation. This technique
// for defining context keys was copied from Go 1.7's new use of context in net/http.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "jwtauth context value " + k.name
}

const (
	defaultTokenDuration = 24 * time.Hour

	JWTCookie   = `jwt`
	TokenCtxKey = "Token"
	ErrorCtxKey = "Error"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}

func (c Claims) Validate() error {
	if c.Username == `` {
		return e.ErrAuthInvalidToken
	}

	return nil
}

type JWTAuth struct {
	keyFunc jwt.Keyfunc
	log     *zerolog.Logger
}

func NewJWTAuth(keyFunc jwt.Keyfunc, log *zerolog.Logger) *JWTAuth {
	return &JWTAuth{
		keyFunc: keyFunc,
		log:     log,
	}
}

func (auth *JWTAuth) Encoder() TokenEncoder {
	return func(username string, userID uuid.UUID) (string, error) {
		now := time.Now()

		claims := &Claims{
			UserID:   userID,
			Username: username,
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(defaultTokenDuration)),
			},
		}

		auth.log.Info().
			Str("method", "HS256").
			Str("user_id", userID.String()).
			Str("username", username).
			Msg("generating new user token")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		signingKey, err := auth.keyFunc(token)
		if err != nil {
			auth.log.Error().Err(err).
				Str("user_id", userID.String()).
				Str("username", username).
				Msg(`failed to retrieve signing key`)

			return ``, e.ErrAuthGenerateToken
		}

		tokenString, err := token.SignedString(signingKey)
		if err != nil {
			auth.log.Error().Err(err).
				Str("user_id", userID.String()).
				Str("username", username).
				Msg(`failed to sign token`)

			return ``, e.ErrAuthGenerateToken
		}

		return tokenString, nil
	}
}

func (auth *JWTAuth) VerifyToken(tokenString string) (*jwt.Token, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// add signature method validation
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, e.ErrAuthInvalidKeyType
		}

		return auth.keyFunc(t)
	})
	if err != nil {
		auth.log.Error().
			Err(err).
			Str(`token`, tokenString).
			Msg(`failed to parse JWT token`)

		return nil, e.ErrAuthInvalidToken
	}

	if !token.Valid {
		auth.log.Error().
			Str(`token`, tokenString).
			Msg("token validation failed")

		return nil, e.ErrAuthInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Validate() != nil {
		auth.log.Error().
			Str("token", tokenString).
			Msg("invalid claims")

		return nil, e.ErrAuthUnknownClaims
	}

	return token, nil
}

func (auth *JWTAuth) VerifyRequest(r *http.Request, extractors ...TokenExtractor) (*jwt.Token, error) {
	var tokenString string

	// Extract token string from the request by calling token find functions in
	// the order they where provided. Further extraction stops if a function
	// returns a non-empty string.
	for _, fn := range extractors {
		tokenString = fn(r)
		if tokenString != `` {
			auth.log.
				Info().
				Msg("token extracted successfully")

			break
		}
	}

	if tokenString == `` {
		auth.log.Error().
			Msg("token not found in request")

		return nil, e.ErrAuthNoToken
	}

	return auth.VerifyToken(tokenString)
}

func Verify(auth *JWTAuth, extractors ...TokenExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			token, err := auth.VerifyRequest(r, extractors...)
			ctx = NewContext(ctx, token, err)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(hfn)
	}
}

// The Verifier always calls the next http handler in sequence, which can either
// be the generic `jwtauth.Authenticator` middleware or your own custom handler
// which checks the request context jwt token and error to prepare a custom
// http response.
func Verifier(auth *JWTAuth) func(http.Handler) http.Handler {
	return Verify(auth, TokenFromHeader, TokenFromCookie)
}

// Authenticator is a default authentication middleware to enforce access from the
// Verifier middleware request context values. The Authenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through. It's just fine
// until you decide to write something similar and customize your client response.
func Authenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := FromContext(r.Context())
			if err != nil {
				http.Error(w, "Unauthorized: Invalid authentication data", http.StatusUnauthorized)

				return
			}

			if token == nil {
				http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)

				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(hfn)
	}
}

func NewContext(ctx context.Context, token *jwt.Token, err error) context.Context {
	ctx = context.WithValue(ctx, contextKey{TokenCtxKey}, token)
	ctx = context.WithValue(ctx, contextKey{ErrorCtxKey}, err)

	return ctx
}

func FromContext(ctx context.Context) (*jwt.Token, *Claims, error) {
	var claims *Claims

	token, _ := ctx.Value(contextKey{TokenCtxKey}).(*jwt.Token)
	err, _ := ctx.Value(contextKey{ErrorCtxKey}).(error)

	if token != nil {
		if c, ok := token.Claims.(*Claims); ok {
			claims = c
		} else {
			return token, nil, e.ErrAuthUnknownClaims
		}
	}

	return token, claims, err
}
