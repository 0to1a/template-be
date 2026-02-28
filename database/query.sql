-- name: FindUserByEmail :one
SELECT id, email, name, token, created_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: FindUserByToken :one
SELECT id, email, name, selected_company_id, created_at
FROM users
WHERE token = $1::text AND deleted_at IS NULL;

-- name: UpdateUserToken :exec
UPDATE users SET token = $1 WHERE id = $2;

-- name: CreateUser :one
INSERT INTO users (email, name, selected_company_id)
VALUES ($1, $2, $3)
RETURNING id, email, name, selected_company_id, created_at;

-- name: UpdateUserSelectedCompany :exec
UPDATE users SET selected_company_id = $1 WHERE id = $2;

-- Company queries
-- name: CreateCompany :one
INSERT INTO companies (company_name, owner_id)
VALUES ($1, $2)
RETURNING id, company_name, owner_id, created_at;

-- name: GetCompanyByID :one
SELECT id, company_name, owner_id, created_at
FROM companies
WHERE id = $1 AND deleted_at IS NULL;

-- name: IsUserCompanyOwner :one
SELECT EXISTS(
    SELECT 1 FROM companies
    WHERE owner_id = $1 AND deleted_at IS NULL
) AS is_owner;

-- Company users queries
-- name: AddUserToCompany :one
INSERT INTO company_users (company_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING id, company_id, user_id, role, created_at;

-- name: GetCompanyUser :one
SELECT id, company_id, user_id, role, created_at
FROM company_users
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetUserCompanies :many
SELECT c.id, c.company_name, c.owner_id, c.created_at, cu.role
FROM companies c
JOIN company_users cu ON cu.company_id = c.id
WHERE cu.user_id = $1 AND cu.deleted_at IS NULL AND c.deleted_at IS NULL;

-- name: IsUserMemberOfCompany :one
SELECT EXISTS(
    SELECT 1 FROM company_users
    WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL
) AS is_member;

-- name: GetCompanyUserRole :one
SELECT role FROM company_users
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: GetCompanyMembers :many
SELECT u.id, u.name, u.email, cu.role
FROM company_users cu
JOIN users u ON cu.user_id = u.id
WHERE cu.company_id = $1 AND cu.deleted_at IS NULL AND u.deleted_at IS NULL;

-- name: RemoveUserFromCompany :exec
UPDATE company_users SET deleted_at = NOW()
WHERE company_id = $1 AND user_id = $2 AND deleted_at IS NULL;
