package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"project/compiled"
)

var (
	ErrInvalidOTP   = errors.New("invalid OTP")
	ErrUserNotFound = errors.New("user not found")
)

const validOTP = "123456"

type AuthService struct {
	queries *compiled.Queries
}

func NewAuthService(queries *compiled.Queries) *AuthService {
	return &AuthService{queries: queries}
}

func (s *AuthService) Login(ctx context.Context, email, otp string) (string, error) {
	if otp != validOTP {
		return "", ErrInvalidOTP
	}

	user, err := s.queries.FindUserByEmail(ctx, email)
	if err != nil {
		return "", ErrUserNotFound
	}

	token := generateToken(10)

	err = s.queries.UpdateUserToken(ctx, compiled.UpdateUserTokenParams{
		ID:    user.ID,
		Token: pgtype.Text{String: token, Valid: true},
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func generateToken(length int) string {
	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}
