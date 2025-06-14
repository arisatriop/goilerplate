CREATE TABLE roles (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	name VARCHAR(100) UNIQUE NOT NULL,
	"order" INT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	created_by VARCHAR(100) NOT NULL,
	updated_at TIMESTAMP NULL,
	updated_by VARCHAR(100) NULL,
	deleted_at TIMESTAMP NULL,
	deleted_by VARCHAR(100) null
);



