-- Migration: plans
-- Created at: 2025-10-31T14:01:30+07:00

-- Add your up migration here
CREATE TABLE plans (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    plan_type_id CHAR(36) NOT NULL ,
    duration_in_days INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL ,
    is_active TINYINT(1) NOT NULL DEFAULT 1 ,
    created_by VARCHAR(255) NOT NULL ,
    updated_by VARCHAR(255) NOT NULL ,
    deleted_by VARCHAR(255) DEFAULT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
    deleted_at TIMESTAMP NULL DEFAULT NULL 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes for better performance
CREATE INDEX idx_plans_plan_type_id ON plans(plan_type_id);
CREATE INDEX idx_plans_is_active ON plans(is_active);
CREATE INDEX idx_plans_created_by ON plans(created_by);
CREATE INDEX idx_plans_deleted_at ON plans(deleted_at);

-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_plans_active ON plans(plan_type_id, is_active, deleted_at);

-- Foreign key constraint
ALTER TABLE plans
ADD CONSTRAINT fk_plans_plan_types
FOREIGN KEY (plan_type_id) REFERENCES plan_types(id)
ON DELETE CASCADE;