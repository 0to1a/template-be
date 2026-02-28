package handler

import (
	"context"
	"log"
	"strings"
	"sync"

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
	"/api.API/Health":          true,
	"/api.API/Login":           true,
	"/api.API/RequestLoginOTP": true,
}

type Handler struct {
	compiled.UnimplementedAPIServer
	authService    *service.AuthService
	companyService *service.CompanyService
	queries        *compiled.Queries
	tokenCache     sync.Map
}

func NewHandler(authService *service.AuthService, companyService *service.CompanyService, queries *compiled.Queries) *Handler {
	return &Handler{
		authService:    authService,
		companyService: companyService,
		queries:        queries,
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

type AuthenticatedUser struct {
	ID                int32
	Email             string
	Name              string
	SelectedCompanyID int32
	CreatedAt         string
	Token             string
}

func (h *Handler) authenticate(ctx context.Context) (*AuthenticatedUser, error) {
	token, err := extractToken(ctx)
	if err != nil {
		return nil, err
	}

	if cached, ok := h.tokenCache.Load(token); ok {
		return cached.(*AuthenticatedUser), nil
	}

	row, err := h.queries.FindUserByToken(ctx, token)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	var selectedCompanyID int32
	if row.SelectedCompanyID.Valid {
		selectedCompanyID = row.SelectedCompanyID.Int32
	}

	user := &AuthenticatedUser{
		ID:                row.ID,
		Email:             row.Email,
		Name:              row.Name,
		SelectedCompanyID: selectedCompanyID,
		CreatedAt:         row.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		Token:             token,
	}
	h.cacheSetToken(token, user)

	return user, nil
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

func UserFromContext(ctx context.Context) (*AuthenticatedUser, bool) {
	user, ok := ctx.Value(UserContextKey).(*AuthenticatedUser)
	return user, ok
}

func (h *Handler) LoadTokenCache(ctx context.Context) {
	rows, err := h.queries.GetAllUsersWithToken(ctx)
	if err != nil {
		log.Printf("Failed to load token cache: %v", err)
		return
	}

	for _, row := range rows {
		var selectedCompanyID int32
		if row.SelectedCompanyID.Valid {
			selectedCompanyID = row.SelectedCompanyID.Int32
		}

		h.tokenCache.Store(row.Token.String, &AuthenticatedUser{
			ID:                row.ID,
			Email:             row.Email,
			Name:              row.Name,
			SelectedCompanyID: selectedCompanyID,
			CreatedAt:         row.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			Token:             row.Token.String,
		})
	}

	log.Printf("Token cache loaded: %d entries", len(rows))
}

func (h *Handler) cacheSetToken(token string, user *AuthenticatedUser) {
	h.tokenCache.Store(token, user)
}

func (h *Handler) cacheDeleteByUserID(userID int32) {
	h.tokenCache.Range(func(key, value any) bool {
		if value.(*AuthenticatedUser).ID == userID {
			h.tokenCache.Delete(key)
			return false
		}
		return true
	})
}
