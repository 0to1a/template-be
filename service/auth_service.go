package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/resend/resend-go/v2"

	"project/compiled"
)

var (
	ErrInvalidOTP   = errors.New("invalid OTP")
	ErrUserNotFound = errors.New("user not found")
)

const validOTP = "123456"

type AuthService struct {
	queries      *compiled.Queries
	resendAPIKey string
}

func NewAuthService(queries *compiled.Queries, resendAPIKey string) *AuthService {
	return &AuthService{queries: queries, resendAPIKey: resendAPIKey}
}

func (s *AuthService) RequestOTP(ctx context.Context, email string) error {
	user, err := s.queries.FindUserByEmail(ctx, email)
	if err != nil {
		return nil // silent success for non-existent users
	}

	if strings.HasSuffix(user.Email, "@localhost") {
		return nil // localhost users use hardcoded OTP
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	err = s.queries.UpdateUserOTP(ctx, compiled.UpdateUserOTPParams{
		Otp:          pgtype.Text{String: otp, Valid: true},
		OtpExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		ID:           user.ID,
	})
	if err != nil {
		return err
	}

	if s.resendAPIKey != "" {
		client := resend.NewClient(s.resendAPIKey)
		_, err = client.Emails.Send(&resend.SendEmailRequest{
			From:    "noreply@yourdomain.com",
			To:      []string{email},
			Subject: "Your login OTP",
			Text:    fmt.Sprintf("Your OTP is: %s\nIt expires in 5 minutes.", otp),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, email, otp string) (string, error) {
	user, err := s.queries.FindUserByEmail(ctx, email)
	if err != nil {
		return "", ErrUserNotFound
	}

	if strings.HasSuffix(email, "@localhost") {
		if otp != validOTP {
			return "", ErrInvalidOTP
		}
	} else {
		if !user.Otp.Valid || user.Otp.String != otp {
			return "", ErrInvalidOTP
		}
		if !user.OtpExpiresAt.Valid || time.Now().After(user.OtpExpiresAt.Time) {
			return "", ErrInvalidOTP
		}
		// Clear OTP after successful use
		_ = s.queries.UpdateUserOTP(ctx, compiled.UpdateUserOTPParams{
			Otp:          pgtype.Text{Valid: false},
			OtpExpiresAt: pgtype.Timestamp{Valid: false},
			ID:           user.ID,
		})
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

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
