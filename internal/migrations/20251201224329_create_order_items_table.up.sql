-- Create order_items table to store items in each order (MySQL)
-- Stores product snapshot at time of order (immutable after creation)

CREATE TABLE IF NOT EXISTS `order_items` (
    -- Primary key
    `id` CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    `order_id` CHAR(36) NOT NULL,
    `product_id` CHAR(36) NOT NULL,

    -- Product snapshot (captured at order time)
    `name` VARCHAR(255) NOT NULL ,
    `price` DECIMAL(15,2) NOT NULL ,
    `image` TEXT NULL ,
    `quantity` INT NOT NULL DEFAULT 1,
    `subtotal` DECIMAL(15,2) NOT NULL ,
    `notes` TEXT NULL ,

    -- Audit fields (simplified - items are immutable after creation)
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `created_by` CHAR(36) NOT NULL DEFAULT 'system',

    -- Indexes
    INDEX `idx_order_items_order_id` (`order_id`),

    -- Foreign keys
    CONSTRAINT `fk_order_items_order` FOREIGN KEY (`order_id`) REFERENCES `orders`(`id`) ON DELETE RESTRICT

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Stores individual items in each order with product snapshot';
