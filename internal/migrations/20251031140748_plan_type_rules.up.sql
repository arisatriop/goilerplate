-- Migration: plan_type_rules
-- Created at: 2025-10-31T14:07:48+07:00

-- Add your up migration here
CREATE TABLE plan_type_rules (
    id CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    plan_type_id CHAR(36) NOT NULL ,
    rule VARCHAR(255) NOT NULL,
    rule_value VARCHAR(255) NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1 ,
    created_by VARCHAR(255) NOT NULL ,
    updated_by VARCHAR(255) NOT NULL ,
    deleted_by VARCHAR(255) DEFAULT NULL ,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
    deleted_at TIMESTAMP NULL DEFAULT NULL 
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- Create indexes for better performance
CREATE INDEX idx_plan_type_rules_plan_type_id ON plan_type_rules(plan_type_id);
CREATE INDEX idx_plan_type_rules_rule ON plan_type_rules(rule);
CREATE INDEX idx_plan_type_rules_is_active ON plan_type_rules(is_active);
CREATE INDEX idx_plan_type_rules_deleted_at ON plan_type_rules(deleted_at);


-- Composite index for active records (MySQL doesn't support partial indexes)
CREATE INDEX idx_plan_type_rules_active ON plan_type_rules(plan_type_id, rule, is_active, deleted_at);

-- Foreign key constraint
ALTER TABLE plan_type_rules
ADD CONSTRAINT fk_plan_type_rules_plan_types
FOREIGN KEY (plan_type_id) REFERENCES plan_types(id)
ON DELETE CASCADE;
