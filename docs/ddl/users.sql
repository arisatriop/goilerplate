CREATE TABLE users (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(100) NOT NULL,
	email VARCHAR(100) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
	token VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	created_by VARCHAR(100) NOT NULL,
	updated_at TIMESTAMP NULL,
	updated_by VARCHAR(100) NULL,
	deleted_at TIMESTAMP NULL,
	deleted_by VARCHAR(100) null
);

CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_token ON users(token);



