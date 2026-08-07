CREATE DATABASE IF NOT EXISTS eqs DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE eqs;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone VARCHAR(20) NOT NULL UNIQUE,
    user_type TINYINT NOT NULL DEFAULT 1 COMMENT '1:甲方 2:服务方 3:管理员',
    company_name VARCHAR(100) DEFAULT '',
    credit_score DECIMAL(5,2) DEFAULT 100.00,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '0:禁用 1:正常',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_phone (phone),
    INDEX idx_user_type (user_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    project_type VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    budget_min DECIMAL(12,2) DEFAULT 0,
    budget_max DECIMAL(12,2) DEFAULT 0,
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0:草稿 1:已发布 2:已接单 3:进行中 4:已完成',
    publish_time DATETIME DEFAULT NULL,
    deadline DATETIME DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id BIGINT UNSIGNED NOT NULL,
    supplier_id BIGINT UNSIGNED NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0:待签约 1:进行中 2:待验收 3:已完成 4:纠纷中',
    signed_at DATETIME DEFAULT NULL,
    completed_at DATETIME DEFAULT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id),
    FOREIGN KEY (supplier_id) REFERENCES users(id),
    INDEX idx_project_id (project_id),
    INDEX idx_supplier_id (supplier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Deliverables table
CREATE TABLE IF NOT EXISTS deliverables (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT UNSIGNED NOT NULL,
    milestone VARCHAR(50) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    version INT DEFAULT 1,
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0:待审核 1:已通过 2:已驳回',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id),
    INDEX idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Payouts table
CREATE TABLE IF NOT EXISTS payouts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    order_id BIGINT UNSIGNED DEFAULT NULL,
    amount DECIMAL(12,2) NOT NULL,
    type VARCHAR(20) NOT NULL COMMENT 'recharge, withdraw, payment, refund',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '0:处理中 1:成功 2:失败',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (order_id) REFERENCES orders(id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
