DROP INDEX IF EXISTS idx_company_users_user_id;
DROP INDEX IF EXISTS idx_companies_owner_id;
ALTER TABLE users DROP COLUMN IF EXISTS selected_company_id;
DROP TABLE IF EXISTS company_users;
DROP TABLE IF EXISTS companies;
