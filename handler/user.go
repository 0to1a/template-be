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

	// Get user companies
	companies, err := h.companyService.GetUserCompanies(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user companies")
	}

	// Check if user is an owner of any company
	isOwner, err := h.companyService.IsUserOwner(ctx, user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check company ownership")
	}

	// Build company info list
	companyInfoList := make([]*compiled.CompanyInfo, 0, len(companies))
	var selectedCompany *compiled.CompanyInfo

	for _, c := range companies {
		info := &compiled.CompanyInfo{
			Id:        int64(c.ID),
			Name:      c.CompanyName,
			Role:      c.Role,
			IsOwner:   c.OwnerID == user.ID,
			CreatedAt: c.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
		companyInfoList = append(companyInfoList, info)

		if user.SelectedCompanyID != 0 && c.ID == user.SelectedCompanyID {
			selectedCompany = info
		}
	}

	return &compiled.GetProfileResponse{
		Id:              int64(user.ID),
		Email:           user.Email,
		Name:            user.Name,
		CreatedAt:       user.CreatedAt,
		Companies:       companyInfoList,
		SelectedCompany: selectedCompany,
		IsOwner:         isOwner,
	}, nil
}
