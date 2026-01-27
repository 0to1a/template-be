package handler

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"project/compiled"
	"project/service"
)

type contextKey string

const UserContextKey contextKey = "user"

var publicMethods = map[string]bool{
	"/api.API/Health": true,
	"/api.API/Login":  true,
}

type Handler struct {
	compiled.UnimplementedAPIServer
	authService *service.AuthService
	queries     *compiled.Queries
}

func NewHandler(authService *service.AuthService, queries *compiled.Queries) *Handler {
	return &Handler{
		authService: authService,
		queries:     queries,
	}
}

func (h *Handler) AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		user, err := h.authenticate(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, UserContextKey, user)
		return handler(ctx, req)
	}
}

func (h *Handler) authenticate(ctx context.Context) (*compiled.User, error) {
	token, err := extractToken(ctx)
	if err != nil {
		return nil, err
	}

	row, err := h.queries.FindUserByToken(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &compiled.User{
		ID:        row.ID,
		Email:     row.Email,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
	}, nil
}

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	parts := strings.SplitN(authHeader[0], " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	return parts[1], nil
}

func UserFromContext(ctx context.Context) (*compiled.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*compiled.User)
	return user, ok
}
