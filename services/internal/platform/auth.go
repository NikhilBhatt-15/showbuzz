package platform

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const CookieName = "sb_token"

type Claims struct {
	UserID string `json:"uid"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
	jwt.RegisteredClaims
}

type ctxKey struct{}

func secret() []byte { return []byte(Env("JWT_SECRET", "dev-secret-change-me")) }

// IssueToken signs a 30-day session token.
func IssueToken(c Claims) (string, time.Time, error) {
	exp := time.Now().Add(30 * 24 * time.Hour)
	c.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   c.UserID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(exp),
		Issuer:    "showbuzz",
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(secret())
	return tok, exp, err
}

// Authenticate attaches the caller's claims to the request context when a
// valid session cookie or bearer token is present. It never rejects.
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := ""
		if c, err := r.Cookie(CookieName); err == nil {
			raw = c.Value
		} else if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			raw = strings.TrimPrefix(h, "Bearer ")
		}
		if raw != "" {
			var c Claims
			tok, err := jwt.ParseWithClaims(raw, &c, func(*jwt.Token) (any, error) { return secret(), nil },
				jwt.WithValidMethods([]string{"HS256"}))
			if err == nil && tok.Valid {
				r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, &c))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// User returns the authenticated caller, or nil.
func User(ctx context.Context) *Claims {
	c, _ := ctx.Value(ctxKey{}).(*Claims)
	return c
}

// RequireUser rejects requests without a valid session.
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if User(r.Context()) == nil {
			Err(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Please sign in to continue")
			return
		}
		next.ServeHTTP(w, r)
	})
}
