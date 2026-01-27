package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"project/compiled"
)

func (h *Handler) GetProfile(ctx context.Context, req *compiled.GetProfileRequest) (*compiled.GetProfileResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	return &compiled.GetProfileResponse{
		Id:        int64(user.ID),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}
