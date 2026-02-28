package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"project/compiled"
	"project/service"
)

func (h *Handler) RequestLoginOTP(ctx context.Context, req *compiled.RequestLoginOTPRequest) (*compiled.RequestLoginOTPResponse, error) {
	if err := h.authService.RequestOTP(ctx, req.Email); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &compiled.RequestLoginOTPResponse{Success: true}, nil
}

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

	row, err := h.queries.FindUserByToken(ctx, token)
	if err == nil {
		h.cacheDeleteByUserID(row.ID)

		var selectedCompanyID int32
		if row.SelectedCompanyID.Valid {
			selectedCompanyID = row.SelectedCompanyID.Int32
		}
		h.cacheSetToken(token, &AuthenticatedUser{
			ID:                row.ID,
			Email:             row.Email,
			Name:              row.Name,
			SelectedCompanyID: selectedCompanyID,
			CreatedAt:         row.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			Token:             token,
		})
	}

	return &compiled.LoginResponse{
		Token: token,
	}, nil
}
