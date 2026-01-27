package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"project/compiled"
	"project/service"
)

func (h *Handler) Login(ctx context.Context, req *compiled.LoginRequest) (*compiled.LoginResponse, error) {
	token, err := h.authService.Login(ctx, req.Email, req.Otp)
	if err != nil {
		switch err {
		case service.ErrInvalidOTP:
			return nil, status.Error(codes.InvalidArgument, "invalid OTP")
		case service.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, "user not found")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &compiled.LoginResponse{
		Token: token,
	}, nil
}
