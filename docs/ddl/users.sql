CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	username VARCHAR(100) UNIQUE NOT NULL,
	fullname VARCHAR(100) NOT NULL,
	email VARCHAR(100) UNIQUE NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	created_by VARCHAR(100) NOT NULL,
	updated_at TIMESTAMP NULL,
	updated_by VARCHAR(100) NULL,
	deleted_at TIMESTAMP NULL,
	deleted_by VARCHAR(100) null
);

CREATE INDEX idx_users_fullname ON users(fullname);



