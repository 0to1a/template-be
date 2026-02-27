CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE company_users (
    id SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES companies(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'member')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(company_id, user_id)
);

ALTER TABLE users ADD COLUMN selected_company_id INTEGER REFERENCES companies(id);

CREATE INDEX idx_companies_owner_id ON companies(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_company_users_user_id ON company_users(user_id) WHERE deleted_at IS NULL;
