-- ============================================================
-- EQS 工程服务 SaaS 平台 · 初始数据库结构 (V6)
-- 对齐 PRD-v6.0 §8.1 共 21 张表
-- 说明：
--   1. 平台不设资金池；资金经持牌机构分账/存管，本地仅存流水
--   2. "仲裁"=专家评审+平台调解，保留当事人法定救济权利
--   3. 评论/审计/规则为后续迭代字段，MVP 保留结构
-- ============================================================

CREATE DATABASE IF NOT EXISTS eqs DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE eqs;

-- ------------------------------------------------------------
-- 1. 用户表
-- user_type: 1-甲方 2-服务方 3-管理员 4-评审专家
-- wx_openid 允许 NULL（手机号登录用户无微信），不再用空串避免唯一冲突
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    phone        VARCHAR(20)  NOT NULL COMMENT '注册手机号',
    wx_openid    VARCHAR(100) NULL COMMENT '微信小程序openid',
    wx_unionid   VARCHAR(100) NULL COMMENT '微信unionid',
    user_type    TINYINT      NOT NULL DEFAULT 1 COMMENT '1:甲方 2:服务方 3:管理员 4:评审专家',
    company_name VARCHAR(100) DEFAULT '' COMMENT '企业名称',
    credit_score DECIMAL(5,2) NOT NULL DEFAULT 100.00 COMMENT '信誉分 0-100',
    status       TINYINT      NOT NULL DEFAULT 1 COMMENT '0:禁用 1:正常',
    created_at   DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_phone (phone),
    UNIQUE KEY uk_wx_openid (wx_openid),
    KEY idx_user_type (user_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 2. 项目表
-- service_type: cost/supervision/geotech/design
-- publish_scope: public(公开) / invited(定向邀请)
-- status: 0-草稿 1-已发布 2-已接单 3-进行中 4-已完成
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS projects (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id       BIGINT UNSIGNED NOT NULL COMMENT '发布业主',
    project_type  VARCHAR(50)  NOT NULL COMMENT '项目类型',
    service_type  VARCHAR(50)  DEFAULT '' COMMENT '服务类型',
    title         VARCHAR(200) NOT NULL,
    description   TEXT         NULL,
    address       VARCHAR(300) DEFAULT '',
    longitude     DECIMAL(10,6) DEFAULT 0,
    latitude      DECIMAL(10,6) DEFAULT 0,
    budget_min    DECIMAL(12,2) DEFAULT 0,
    budget_max    DECIMAL(12,2) DEFAULT 0,
    publish_scope VARCHAR(20)  DEFAULT 'public',
    status        TINYINT      NOT NULL DEFAULT 0 COMMENT '0:草稿 1:已发布 2:已接单 3:进行中 4:已完成',
    publish_time  DATETIME     NULL,
    deadline      DATETIME     NULL,
    created_at    DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_user_id (user_id),
    KEY idx_status (status),
    KEY idx_service_type (service_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 3. 服务方资质表
-- verification_method: OCR/manual/external
-- verification_status: pending/approved/rejected/expired
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS supplier_qualifications (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    supplier_id         BIGINT UNSIGNED NOT NULL,
    qualification_type  VARCHAR(50)  NOT NULL COMMENT '资质类型',
    certificate_no      VARCHAR(100) DEFAULT '' COMMENT '证书编号',
    level               VARCHAR(50)  DEFAULT '' COMMENT '等级',
    scope               VARCHAR(200) DEFAULT '' COMMENT '执业范围',
    valid_from          DATETIME     NULL,
    valid_to            DATETIME     NULL,
    verification_method VARCHAR(20)  DEFAULT 'manual',
    verification_status VARCHAR(20)  DEFAULT 'pending',
    evidence_file_id    BIGINT UNSIGNED DEFAULT 0,
    reviewed_by         BIGINT UNSIGNED DEFAULT 0,
    reviewed_at         DATETIME     NULL,
    created_at          DATETIME     DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_supplier_id (supplier_id),
    KEY idx_verify_status (verification_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 4. 报价表
-- status: submitted/selected/rejected/withdrawn
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bids (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id       BIGINT UNSIGNED NOT NULL,
    supplier_id      BIGINT UNSIGNED NOT NULL,
    amount           DECIMAL(12,2) NOT NULL,
    service_days     INT DEFAULT 0,
    proposal_file_id BIGINT UNSIGNED DEFAULT 0,
    status           VARCHAR(20) NOT NULL DEFAULT 'submitted',
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_project_id (project_id),
    KEY idx_supplier_id (supplier_id),
    UNIQUE KEY uk_project_supplier (project_id, supplier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 5. 订单表
-- status: 0-待签约 1-进行中 2-待验收 3-已完成 4-纠纷中
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS orders (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id      BIGINT UNSIGNED NOT NULL,
    supplier_id     BIGINT UNSIGNED NOT NULL,
    selected_bid_id BIGINT UNSIGNED DEFAULT 0 COMMENT '中选报价',
    amount          DECIMAL(12,2) NOT NULL,
    status          TINYINT NOT NULL DEFAULT 0,
    signed_at       DATETIME NULL,
    completed_at    DATETIME NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_project_id (project_id),
    KEY idx_supplier_id (supplier_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 6. 付款节点表
-- status: pending/submitted/accepted/rejected/disputed/settled
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payment_milestones (
    id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id          BIGINT UNSIGNED NOT NULL,
    name              VARCHAR(100) NOT NULL,
    sequence          INT NOT NULL DEFAULT 0,
    ratio             DECIMAL(5,2) NOT NULL DEFAULT 0 COMMENT '占比% 合计100',
    amount            DECIMAL(12,2) NOT NULL DEFAULT 0,
    acceptance_due_at DATETIME NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending',
    accepted_by       BIGINT UNSIGNED DEFAULT 0,
    accepted_at       DATETIME NULL,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 7. 合同表
-- status: draft/signing/signed/voided
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS contracts (
    id               BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id         BIGINT UNSIGNED NOT NULL,
    template_id      BIGINT UNSIGNED DEFAULT 0,
    template_version VARCHAR(20) DEFAULT '1.0',
    contract_no      VARCHAR(50)  NOT NULL,
    contract_file_id BIGINT UNSIGNED DEFAULT 0,
    sign_provider    VARCHAR(50) DEFAULT 'mock' COMMENT '电子签章服务商',
    sign_flow_id     VARCHAR(100) NOT NULL COMMENT '第三方签署流程号',
    status           VARCHAR(20) DEFAULT 'draft',
    signed_at        DATETIME NULL,
    created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_order_id (order_id),
    UNIQUE KEY uk_contract_no (contract_no),
    UNIQUE KEY uk_sign_flow_id (sign_flow_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 8. 交付物表
-- status: 0-待审核 1-已通过 2-已驳回
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS deliverables (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id        BIGINT UNSIGNED NOT NULL,
    milestone_id    BIGINT UNSIGNED NOT NULL,
    milestone       VARCHAR(50) DEFAULT '' COMMENT '节点名称快照',
    file_name       VARCHAR(200) DEFAULT '',
    file_url        VARCHAR(500) NOT NULL,
    version         INT NOT NULL DEFAULT 1,
    status          TINYINT NOT NULL DEFAULT 0,
    checklist_result JSON NULL COMMENT '模板验收清单结果',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_order_id (order_id),
    KEY idx_milestone_id (milestone_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 9. 项目文件表
-- file_type: pdf/image/dwg/dxf/document
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_files (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id    BIGINT UNSIGNED DEFAULT 0,
    order_id      BIGINT UNSIGNED DEFAULT 0,
    uploader_id   BIGINT UNSIGNED DEFAULT 0,
    original_name VARCHAR(200) DEFAULT '',
    file_type     VARCHAR(20)  DEFAULT '',
    storage_key   VARCHAR(500) NOT NULL COMMENT 'COS 对象键',
    version       INT NOT NULL DEFAULT 1,
    parent_file_id BIGINT UNSIGNED DEFAULT 0,
    sha256        CHAR(64) DEFAULT '',
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_project_id (project_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 10. 文件批注表
-- status: active/resolved/deleted
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS file_annotations (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    file_id    BIGINT UNSIGNED NOT NULL,
    author_id  BIGINT UNSIGNED DEFAULT 0,
    page_no    INT DEFAULT 1,
    x_ratio    DECIMAL(5,4) DEFAULT 0,
    y_ratio    DECIMAL(5,4) DEFAULT 0,
    content    TEXT,
    status     VARCHAR(20) DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_file_id (file_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 11. 支付与结算流水表（非托管，仅记录经持牌机构结果）
-- type: payment/settlement/refund/freeze/unfreeze
-- status: 0-处理中 1-成功 2-失败
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS payment_transactions (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id                 BIGINT UNSIGNED NOT NULL,
    order_id                BIGINT UNSIGNED DEFAULT 0,
    milestone_id            BIGINT UNSIGNED DEFAULT 0,
    amount                  DECIMAL(12,2) NOT NULL,
    type                    VARCHAR(20) NOT NULL,
    channel                 VARCHAR(20) DEFAULT 'mock' COMMENT 'wechat/bank/持牌机构',
    external_transaction_id VARCHAR(100) NOT NULL COMMENT '持牌机构单号',
    status                  TINYINT NOT NULL DEFAULT 0,
    created_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_external_txn (external_transaction_id),
    KEY idx_user_id (user_id),
    KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 12. 现场打卡表
-- verification_status: valid/exception/manual_approved/rejected
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS attendance_records (
    id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id           BIGINT UNSIGNED NOT NULL,
    user_id            BIGINT UNSIGNED NOT NULL,
    check_in_at        DATETIME NOT NULL,
    longitude          DECIMAL(10,6) DEFAULT 0,
    latitude           DECIMAL(10,6) DEFAULT 0,
    distance_meters    INT DEFAULT 0 COMMENT '与工程点位距离',
    evidence_file_id   BIGINT UNSIGNED DEFAULT 0,
    verification_status VARCHAR(20) DEFAULT 'ok',
    created_at         DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_order_id (order_id),
    KEY idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 13. 标准交付模板
-- status: draft/active/retired
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS delivery_templates (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    service_type VARCHAR(50) DEFAULT '',
    name         VARCHAR(100) NOT NULL,
    version      VARCHAR(20) DEFAULT '1.0',
    file_id      BIGINT UNSIGNED DEFAULT 0,
    checklist    JSON NULL COMMENT '验收清单',
    status       VARCHAR(20) DEFAULT 'draft',
    effective_at DATETIME NULL,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 14. 合同模板
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS contract_templates (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    service_type VARCHAR(50) DEFAULT '',
    name         VARCHAR(100) NOT NULL,
    version      VARCHAR(20) DEFAULT '1.0',
    content      TEXT COMMENT '合同模板文本/占位符',
    status       VARCHAR(20) DEFAULT 'active',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 15. 争议表（专家评审+平台调解；保留法定救济权利）
-- status: evidence/review/mediation/reconsideration/closed
-- resolution_type: settlement/agreement/award/judgment
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS disputes (
    id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id           BIGINT UNSIGNED NOT NULL,
    milestone_id       BIGINT UNSIGNED DEFAULT 0,
    initiator_id       BIGINT UNSIGNED DEFAULT 0,
    reason             TEXT COMMENT '争议事由',
    claim              TEXT COMMENT '诉求',
    expert_result      JSON NULL COMMENT '专家评审结论',
    resolution_type    VARCHAR(30) DEFAULT '',
    resolution_file_id BIGINT UNSIGNED DEFAULT 0,
    status             VARCHAR(30) DEFAULT 'evidence',
    closed_at          DATETIME NULL,
    created_at         DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 16. 争议证据表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dispute_evidences (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    dispute_id BIGINT UNSIGNED NOT NULL,
    user_id    BIGINT UNSIGNED DEFAULT 0,
    file_id    BIGINT UNSIGNED DEFAULT 0,
    sha256     VARCHAR(64) DEFAULT '',
    content    TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_dispute_id (dispute_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 17. 争议专家指派表
-- recusal_status: recused/not_required/required
-- vote: support_client/support_supplier/partial
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS dispute_expert_assignments (
    id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    dispute_id        BIGINT UNSIGNED NOT NULL,
    expert_user_id    BIGINT UNSIGNED DEFAULT 0,
    conflict_declared TINYINT DEFAULT 0 COMMENT '是否披露利益冲突',
    recusal_status   VARCHAR(20) DEFAULT 'not_required',
    opinion          TEXT,
    vote             VARCHAR(20) DEFAULT '',
    submitted_at      DATETIME NULL,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_dispute_id (dispute_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 18. 评价表（1-5星，联动信用分）
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reviews (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id    BIGINT UNSIGNED NOT NULL,
    reviewer_id BIGINT UNSIGNED DEFAULT 0,
    reviewee_id BIGINT UNSIGNED DEFAULT 0,
    rating      TINYINT NOT NULL DEFAULT 5 COMMENT '1-5',
    content     TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_order_id (order_id),
    KEY idx_reviewee_id (reviewee_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 19. 站内消息表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS messages (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sender_id   BIGINT UNSIGNED DEFAULT 0,
    receiver_id BIGINT UNSIGNED NOT NULL,
    order_id    BIGINT UNSIGNED DEFAULT 0,
    content     TEXT,
    is_read     TINYINT DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_receiver_id (receiver_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 20. 通知表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notifications (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT UNSIGNED NOT NULL,
    title      VARCHAR(200) DEFAULT '',
    content    TEXT,
    type       VARCHAR(50) DEFAULT 'system' COMMENT 'order/system/payment',
    is_read    TINYINT DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ------------------------------------------------------------
-- 21. 审计日志表
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT UNSIGNED DEFAULT 0,
    action      VARCHAR(50) NOT NULL,
    target_type VARCHAR(50) DEFAULT '',
    target_id   BIGINT UNSIGNED DEFAULT 0,
    detail      JSON NULL,
    ip          VARCHAR(50) DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;