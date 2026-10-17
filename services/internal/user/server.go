// Package user handles passwordless sign-in (OTP to email or mobile) and
// profiles. OTP delivery is asynchronous: the service publishes an
// otp.requested event and the notify service sends it.
package user

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/NikhilBhatt-15/showbuzz/services/internal/platform"
)

const (
	otpTTL      = 5 * time.Minute
	otpCooldown = 30 * time.Second
	maxAttempts = 5
)

var (
	emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	phoneRe = regexp.MustCompile(`^[6-9]\d{9}$`)
)

type Server struct {
	db     *pgxpool.Pool
	rdb    *redis.Client
	pub    *platform.Producer
	devOTP bool
}

func New(db *pgxpool.Pool, rdb *redis.Client, pub *platform.Producer) *Server {
	return &Server{db: db, rdb: rdb, pub: pub, devOTP: platform.EnvBool("DEV_SHOW_OTP")}
}

func (s *Server) Routes(r chi.Router) {
	r.Route("/api/user/v1", func(r chi.Router) {
		r.Post("/otp", s.requestOTP)
		r.Post("/verify", s.verifyOTP)
		r.Post("/logout", s.logout)
		r.With(platform.RequireUser).Get("/me", s.me)
		r.With(platform.RequireUser).Put("/me", s.updateMe)
	})
}

type User struct {
	ID         string    `json:"id"`
	Identifier string    `json:"identifier"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
}

// normalise returns the identifier and its channel ("email" or "sms").
func normalise(raw string) (string, string, bool) {
	id := strings.ToLower(strings.TrimSpace(raw))
	id = strings.TrimPrefix(strings.ReplaceAll(id, " ", ""), "+91")
	switch {
	case emailRe.MatchString(id):
		return id, "email", true
	case phoneRe.MatchString(id):
		return id, "sms", true
	}
	return "", "", false
}

func otpKey(id string) string { return "otp:" + id }

func (s *Server) requestOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Identifier string `json:"identifier"`
	}
	if err := platform.Decode(r, &req); err != nil {
		platform.Err(w, 400, "BAD_REQUEST", "Invalid request body")
		return
	}
	id, channel, ok := normalise(req.Identifier)
	if !ok {
		platform.Err(w, 400, "INVALID_IDENTIFIER", "Enter a valid email address or 10-digit mobile number")
		return
	}
	if ok, err := s.rdb.SetNX(ctx, "otp_cooldown:"+id, 1, otpCooldown).Result(); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	} else if !ok {
		platform.Err(w, 429, "TOO_MANY_REQUESTS", "Please wait a few seconds before requesting another OTP")
		return
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	otp := fmt.Sprintf("%06d", n.Int64())
	if err := s.rdb.HSet(ctx, otpKey(id), "otp", otp, "attempts", 0).Err(); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	s.rdb.Expire(ctx, otpKey(id), otpTTL)
	s.pub.Publish(ctx, platform.TopicOTPRequested, id, map[string]string{"identifier": id, "channel": channel, "otp": otp})

	resp := map[string]any{"sent": true, "channel": channel, "identifier": id, "expiresIn": int(otpTTL.Seconds()),
		"resendIn": int(otpCooldown.Seconds())}
	if s.devOTP {
		resp["devOtp"] = otp
	}
	platform.JSON(w, 200, resp)
}

func setSession(w http.ResponseWriter, u User) error {
	tok, exp, err := platform.IssueToken(platform.Claims{UserID: u.ID, Name: u.Name, Email: u.Email, Phone: u.Phone})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: platform.CookieName, Value: tok, Path: "/", Expires: exp,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: platform.EnvBool("COOKIE_SECURE")})
	return nil
}

func (s *Server) verifyOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Identifier string `json:"identifier"`
		OTP        string `json:"otp"`
	}
	if err := platform.Decode(r, &req); err != nil {
		platform.Err(w, 400, "BAD_REQUEST", "Invalid request body")
		return
	}
	id, channel, ok := normalise(req.Identifier)
	if !ok {
		platform.Err(w, 400, "INVALID_IDENTIFIER", "Enter a valid email address or 10-digit mobile number")
		return
	}
	vals, err := s.rdb.HGetAll(ctx, otpKey(id)).Result()
	if err != nil || vals["otp"] == "" {
		platform.Err(w, 400, "OTP_EXPIRED", "OTP has expired. Please request a new one.")
		return
	}
	attempts, _ := s.rdb.HIncrBy(ctx, otpKey(id), "attempts", 1).Result()
	if attempts > maxAttempts {
		s.rdb.Del(ctx, otpKey(id))
		platform.Err(w, 429, "TOO_MANY_ATTEMPTS", "Too many incorrect attempts. Please request a new OTP.")
		return
	}
	if subtle.ConstantTimeCompare([]byte(vals["otp"]), []byte(strings.TrimSpace(req.OTP))) != 1 {
		platform.Err(w, 400, "OTP_INVALID", "Incorrect OTP. Please try again.")
		return
	}
	s.rdb.Del(ctx, otpKey(id))

	email, phone := "", ""
	if channel == "email" {
		email = id
	} else {
		phone = id
	}
	var u User
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (identifier, email, phone) VALUES ($1, $2, $3)
		ON CONFLICT (identifier) DO UPDATE SET identifier = EXCLUDED.identifier
		RETURNING id, identifier, email, phone, name, created_at`, id, email, phone).
		Scan(&u.ID, &u.Identifier, &u.Email, &u.Phone, &u.Name, &u.CreatedAt)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if err := setSession(w, u); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.JSON(w, 200, map[string]any{"user": u})
}

func (s *Server) logout(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: platform.CookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) load(r *http.Request) (User, error) {
	var u User
	err := s.db.QueryRow(r.Context(), `SELECT id, identifier, email, phone, name, created_at FROM users WHERE id = $1`,
		platform.User(r.Context()).UserID).Scan(&u.ID, &u.Identifier, &u.Email, &u.Phone, &u.Name, &u.CreatedAt)
	return u, err
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	u, err := s.load(r)
	if err != nil {
		platform.Err(w, 401, "UNAUTHENTICATED", "Please sign in to continue")
		return
	}
	platform.NoStore(w)
	platform.JSON(w, 200, map[string]any{"user": u})
}

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := platform.Decode(r, &req); err != nil {
		platform.Err(w, 400, "BAD_REQUEST", "Invalid request body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 60 || (req.Email != "" && !emailRe.MatchString(req.Email)) || (req.Phone != "" && !phoneRe.MatchString(req.Phone)) {
		platform.Err(w, 400, "INVALID_PROFILE", "Please check the details you entered")
		return
	}
	var u User
	err := s.db.QueryRow(r.Context(), `
		UPDATE users SET name = $2, email = COALESCE(NULLIF($3, ''), email), phone = COALESCE(NULLIF($4, ''), phone)
		WHERE id = $1 RETURNING id, identifier, email, phone, name, created_at`,
		platform.User(r.Context()).UserID, req.Name, req.Email, req.Phone).
		Scan(&u.ID, &u.Identifier, &u.Email, &u.Phone, &u.Name, &u.CreatedAt)
	if err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	if err := setSession(w, u); err != nil {
		platform.Err(w, 500, "INTERNAL", err.Error())
		return
	}
	platform.JSON(w, 200, map[string]any{"user": u})
}
