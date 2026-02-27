-- Seed company for user id 1
INSERT INTO companies (company_name, owner_id) VALUES ('Default Company', 1);
INSERT INTO company_users (company_id, user_id, role) VALUES (1, 1, 'admin');
UPDATE users SET selected_company_id = 1 WHERE id = 1;
