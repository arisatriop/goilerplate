-- Create orders table for admin order management (MySQL)
-- Supports both admin orders and future guest orders (user_id can be NULL)

CREATE TABLE IF NOT EXISTS `orders` (
    `id` CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    `order_number` VARCHAR(50) UNIQUE NOT NULL,
    `queue_number` VARCHAR(20) NULL,
    `table_number` VARCHAR(20) NULL,
    `order_type` VARCHAR(20) NOT NULL COMMENT 'e.g., dine-in, takeout',
    `order_status` VARCHAR(20) NOT NULL DEFAULT 'processing' COMMENT 'e.g., processing, completed, cancelled',
    `notes` TEXT NULL,
    `amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    `store_id` CHAR(36) NOT NULL,
    `customer_name` VARCHAR(255) NULL,
    `customer_phone` VARCHAR(20) NULL,
    `customer_email` VARCHAR(255) NULL,    

    -- Audit fields
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by` CHAR(36) NOT NULL DEFAULT 'system',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `updated_by` CHAR(36) NULL,



    -- Indexes
    INDEX `idx_orders_created_at` (`created_at`),
    INDEX `idx_orders_queue_number` (`queue_number`),

    -- Foreign keys
    CONSTRAINT `fk_orders_store` FOREIGN KEY (`store_id`) REFERENCES `stores`(`id`) ON DELETE RESTRICT

    -- No check constraints - validation done in application layer
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Stores customer orders placed by admin or guests';
