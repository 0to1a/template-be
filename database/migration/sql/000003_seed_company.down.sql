UPDATE users SET selected_company_id = NULL WHERE id = 1;
DELETE FROM company_users WHERE user_id = 1 AND company_id = 1;
DELETE FROM companies WHERE id = 1;
