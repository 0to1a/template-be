package handler

import (
	"context"

	"project/compiled"
)

func (h *Handler) Health(ctx context.Context, req *compiled.HealthRequest) (*compiled.HealthResponse, error) {
	return &compiled.HealthResponse{
		Status: "ok",
	}, nil
}
