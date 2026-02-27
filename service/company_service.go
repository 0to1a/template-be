package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"project/compiled"
)

var (
	ErrAlreadyOwner      = errors.New("user already owns a company")
	ErrNotCompanyMember  = errors.New("user is not a member of this company")
	ErrNotAdmin          = errors.New("user is not an admin of the selected company")
	ErrNoSelectedCompany = errors.New("no company selected")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserAlreadyMember = errors.New("user is already a member of this company")
)

type CompanyService struct {
	queries *compiled.Queries
}

func NewCompanyService(queries *compiled.Queries) *CompanyService {
	return &CompanyService{queries: queries}
}

func (s *CompanyService) CreateCompany(ctx context.Context, userID int32, companyName string) (*compiled.CreateCompanyRow, error) {
	// Check if user already owns a company
	isOwner, err := s.queries.IsUserCompanyOwner(ctx, userID)
	if err != nil {
		return nil, err
	}
	if isOwner {
		return nil, ErrAlreadyOwner
	}

	// Create company
	company, err := s.queries.CreateCompany(ctx, compiled.CreateCompanyParams{
		CompanyName: companyName,
		OwnerID:     userID,
	})
	if err != nil {
		return nil, err
	}

	// Add user as admin
	_, err = s.queries.AddUserToCompany(ctx, compiled.AddUserToCompanyParams{
		CompanyID: company.ID,
		UserID:    userID,
		Role:      "admin",
	})
	if err != nil {
		return nil, err
	}

	// Set as selected company
	err = s.queries.UpdateUserSelectedCompany(ctx, compiled.UpdateUserSelectedCompanyParams{
		SelectedCompanyID: pgtype.Int4{Int32: company.ID, Valid: true},
		ID:                userID,
	})
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *CompanyService) SelectCompany(ctx context.Context, userID int32, companyID int32) (*compiled.GetCompanyByIDRow, error) {
	// Check if user is member of this company
	isMember, err := s.queries.IsUserMemberOfCompany(ctx, compiled.IsUserMemberOfCompanyParams{
		CompanyID: companyID,
		UserID:    userID,
	})
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotCompanyMember
	}

	// Update selected company
	err = s.queries.UpdateUserSelectedCompany(ctx, compiled.UpdateUserSelectedCompanyParams{
		SelectedCompanyID: pgtype.Int4{Int32: companyID, Valid: true},
		ID:                userID,
	})
	if err != nil {
		return nil, err
	}

	// Return company info
	company, err := s.queries.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *CompanyService) GetUserCompanies(ctx context.Context, userID int32) ([]compiled.GetUserCompaniesRow, error) {
	return s.queries.GetUserCompanies(ctx, userID)
}

func (s *CompanyService) IsUserOwner(ctx context.Context, userID int32) (bool, error) {
	return s.queries.IsUserCompanyOwner(ctx, userID)
}

func (s *CompanyService) InviteUser(ctx context.Context, inviterID int32, selectedCompanyID int32, email, name, role string) (*compiled.CreateUserRow, error) {
	// Check if inviter is admin in selected company
	inviterRole, err := s.queries.GetCompanyUserRole(ctx, compiled.GetCompanyUserRoleParams{
		CompanyID: selectedCompanyID,
		UserID:    inviterID,
	})
	if err != nil {
		return nil, ErrNotAdmin
	}
	if inviterRole != "admin" {
		return nil, ErrNotAdmin
	}

	// Check if user already exists
	existingUser, err := s.queries.FindUserByEmail(ctx, email)
	if err == nil {
		// User exists, check if already member
		isMember, err := s.queries.IsUserMemberOfCompany(ctx, compiled.IsUserMemberOfCompanyParams{
			CompanyID: selectedCompanyID,
			UserID:    existingUser.ID,
		})
		if err != nil {
			return nil, err
		}
		if isMember {
			return nil, ErrUserAlreadyMember
		}

		// Add existing user to company
		_, err = s.queries.AddUserToCompany(ctx, compiled.AddUserToCompanyParams{
			CompanyID: selectedCompanyID,
			UserID:    existingUser.ID,
			Role:      role,
		})
		if err != nil {
			return nil, err
		}

		return &compiled.CreateUserRow{
			ID:                existingUser.ID,
			Email:             existingUser.Email,
			Name:              existingUser.Name,
			SelectedCompanyID: pgtype.Int4{},
			CreatedAt:         existingUser.CreatedAt,
		}, nil
	}

	// Create new user with selected company set to the inviting company
	newUser, err := s.queries.CreateUser(ctx, compiled.CreateUserParams{
		Email:             email,
		Name:              name,
		SelectedCompanyID: pgtype.Int4{Int32: selectedCompanyID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Add user to company
	_, err = s.queries.AddUserToCompany(ctx, compiled.AddUserToCompanyParams{
		CompanyID: selectedCompanyID,
		UserID:    newUser.ID,
		Role:      role,
	})
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (s *CompanyService) GetCompanyByID(ctx context.Context, companyID int32) (*compiled.GetCompanyByIDRow, error) {
	company, err := s.queries.GetCompanyByID(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (s *CompanyService) GetCompanyUserRole(ctx context.Context, companyID, userID int32) (string, error) {
	return s.queries.GetCompanyUserRole(ctx, compiled.GetCompanyUserRoleParams{
		CompanyID: companyID,
		UserID:    userID,
	})
}
