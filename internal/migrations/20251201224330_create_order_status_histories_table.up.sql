-- Create order_status_histories table for audit trail (MySQL)
-- Tracks all status changes for orders with who changed it and when

CREATE TABLE IF NOT EXISTS `order_status_histories` (
    -- Primary key
    `id` CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    `order_id` CHAR(36) NOT NULL,
    `status` VARCHAR(20) NOT NULL ,

    -- Audit fields
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by` CHAR(36) NOT NULL DEFAULT 'system',

    -- Indexes
    INDEX `idx_order_status_histories_created_at` (`created_at`),

    -- Foreign keys
    CONSTRAINT `fk_order_status_history_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE RESTRICT

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Tracks all status changes for orders (audit trail)';
