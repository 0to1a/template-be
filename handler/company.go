package handler

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"project/compiled"
	"project/service"
)

func (h *Handler) CreateCompany(ctx context.Context, req *compiled.CreateCompanyRequest) (*compiled.CreateCompanyResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.CompanyName == "" {
		return nil, status.Error(codes.InvalidArgument, "company_name is required")
	}

	company, err := h.companyService.CreateCompany(ctx, user.ID, req.CompanyName)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create company")
	}

	user.SelectedCompanyID = company.ID
	h.cacheSetToken(user.Token, user)

	return &compiled.CreateCompanyResponse{
		Id:          int64(company.ID),
		CompanyName: company.CompanyName,
		CreatedAt:   company.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (h *Handler) SelectCompany(ctx context.Context, req *compiled.SelectCompanyRequest) (*compiled.SelectCompanyResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.CompanyId == 0 {
		return nil, status.Error(codes.InvalidArgument, "company_id is required")
	}

	company, err := h.companyService.SelectCompany(ctx, user.ID, int32(req.CompanyId))
	if err != nil {
		if errors.Is(err, service.ErrNotCompanyMember) {
			return nil, status.Error(codes.PermissionDenied, "user is not a member of this company")
		}
		return nil, status.Error(codes.Internal, "failed to select company")
	}

	user.SelectedCompanyID = company.ID
	h.cacheSetToken(user.Token, user)

	role, _ := h.companyService.GetCompanyUserRole(ctx, company.ID, user.ID)
	isOwner := company.OwnerID == user.ID

	return &compiled.SelectCompanyResponse{
		Success: true,
		SelectedCompany: &compiled.CompanyInfo{
			Id:        int64(company.ID),
			Name:      company.CompanyName,
			Role:      role,
			IsOwner:   isOwner,
			CreatedAt: company.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

func (h *Handler) InviteUser(ctx context.Context, req *compiled.InviteUserRequest) (*compiled.InviteUserResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if user.SelectedCompanyID == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no company selected")
	}

	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Role != "admin" && req.Role != "member" {
		return nil, status.Error(codes.InvalidArgument, "role must be 'admin' or 'member'")
	}

	invitedUser, err := h.companyService.InviteUser(ctx, user.ID, user.SelectedCompanyID, req.Email, req.Name, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrNotAdmin) {
			return nil, status.Error(codes.PermissionDenied, "only admins can invite users")
		}
		if errors.Is(err, service.ErrUserAlreadyMember) {
			return nil, status.Error(codes.AlreadyExists, "user is already a member of this company")
		}
		return nil, status.Error(codes.Internal, "failed to invite user")
	}

	return &compiled.InviteUserResponse{
		UserId: int64(invitedUser.ID),
		Email:  invitedUser.Email,
		Name:   invitedUser.Name,
		Role:   req.Role,
	}, nil
}

func (h *Handler) ListCompanyMembers(ctx context.Context, req *compiled.ListCompanyMembersRequest) (*compiled.ListCompanyMembersResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if user.SelectedCompanyID == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no company selected")
	}

	members, err := h.companyService.GetCompanyMembers(ctx, user.SelectedCompanyID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get company members")
	}

	var result []*compiled.CompanyMember
	for _, m := range members {
		result = append(result, &compiled.CompanyMember{
			UserId: int64(m.ID),
			Name:   m.Name,
			Email:  m.Email,
			Role:   m.Role,
		})
	}

	return &compiled.ListCompanyMembersResponse{Members: result}, nil
}

func (h *Handler) RemoveCompanyMember(ctx context.Context, req *compiled.RemoveCompanyMemberRequest) (*compiled.RemoveCompanyMemberResponse, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	if user.SelectedCompanyID == 0 {
		return nil, status.Error(codes.FailedPrecondition, "no company selected")
	}
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := h.companyService.RemoveCompanyMember(ctx, user.ID, user.SelectedCompanyID, int32(req.UserId))
	if err != nil {
		if errors.Is(err, service.ErrCannotRemoveSelf) {
			return nil, status.Error(codes.InvalidArgument, "cannot remove yourself from the company")
		}
		if errors.Is(err, service.ErrNotAdmin) {
			return nil, status.Error(codes.PermissionDenied, "only admins can remove members")
		}
		if errors.Is(err, service.ErrNotCompanyMember) {
			return nil, status.Error(codes.NotFound, "user is not a member of this company")
		}
		return nil, status.Error(codes.Internal, "failed to remove company member")
	}

	return &compiled.RemoveCompanyMemberResponse{Success: true}, nil
}
