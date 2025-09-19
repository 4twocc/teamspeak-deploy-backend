-- ===========================================================
-- 文件:        schema_mysql.sql
-- 作用:        Teamspeak One-Click Deploy 用户域数据库全量初始化脚本（丢弃旧表并创建新表）
-- 作者:        AI Assistant
-- 版本:        v1.0.0
-- 说明:        请在空闲维护窗口执行。本脚本包含 DROP 语句，请务必确认目标库无关数据需要保留。
-- 适用:        MySQL 8.0+
-- ===========================================================

SET NAMES utf8mb4;
SET time_zone = 'Asia/Shanghai';
SET sql_safe_updates = 0;

START TRANSACTION;

-- 关闭外键检查以便按任意顺序删除旧表
SET FOREIGN_KEY_CHECKS = 0;

-- 如果历史上存在同名表，这里将直接删除
DROP TABLE IF EXISTS verification_tokens;
DROP TABLE IF EXISTS login_audit;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users_auth;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS users;

-- 重新开启外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ===========================================================
-- 基础用户表：users（核心身份标识 + 登录标识）
-- 说明：
--  - uid 作为业务层友好 ID（UUID），避免对自增主键暴露；
--  - username/email/phone 可为空，但各自唯一；
--  - deleted_at 用于软删除；
-- ===========================================================
CREATE TABLE users (
  id                   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键，自增ID',
  uid                  CHAR(36)        NOT NULL DEFAULT (UUID()) COMMENT '业务UID，建议UUIDv4，8.0.13-可用',
  username             VARCHAR(50)     NULL     COMMENT '用户名，唯一，可为空',
  email                VARCHAR(254)    NULL     COMMENT '邮箱，唯一，可为空',
  phone                VARCHAR(32)     NULL     COMMENT '手机号（E.164），唯一，可为空',

  status               TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：1启用，2停用，3封禁',
  is_email_verified    TINYINT(1)       NOT NULL DEFAULT 0 COMMENT '邮箱是否已验证',
  is_phone_verified    TINYINT(1)       NOT NULL DEFAULT 0 COMMENT '手机是否已验证',

  created_at           TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at           TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  deleted_at           TIMESTAMP        NULL DEFAULT NULL COMMENT '软删除时间（NULL表示未删除）',

  PRIMARY KEY (id),
  UNIQUE KEY uk_users_uid      (uid),
  UNIQUE KEY uk_users_username (username),
  UNIQUE KEY uk_users_email    (email),
  UNIQUE KEY uk_users_phone    (phone),
  KEY        idx_users_status  (status),
  KEY        idx_users_ctime   (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户核心信息';

-- ===========================================================
-- 用户档案表：user_profiles（一对一）
-- 说明：补充用户展示/偏好等信息，与认证/鉴权解耦
-- ===========================================================
CREATE TABLE user_profiles (
  user_id     BIGINT UNSIGNED NOT NULL COMMENT '用户ID=users.id',
  full_name   VARCHAR(100)    NULL     COMMENT '显示名称',
  first_name  VARCHAR(60)     NULL     COMMENT '名',
  last_name   VARCHAR(60)     NULL     COMMENT '姓',
  avatar_url  VARCHAR(512)    NULL     COMMENT '头像',
  locale      VARCHAR(16)     NULL DEFAULT 'zh-CN'        COMMENT '语言/区域',
  timezone    VARCHAR(64)     NULL DEFAULT 'Asia/Shanghai' COMMENT '时区',
  gender      TINYINT UNSIGNED NULL DEFAULT 0             COMMENT '0未知,1男,2女,3其他',
  date_of_birth DATE          NULL     COMMENT '生日',
  bio         VARCHAR(512)    NULL     COMMENT '个人简介',

  created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (user_id),
  CONSTRAINT fk_user_profiles_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
    ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户档案信息（与认证解耦）';

-- ===========================================================
-- 认证表：users_auth（一对一，密码/MFA）
-- 说明：
--  - password_hash 存储完整编码串（如argon2id/bcrypt）；
--  - mfa_secret 可存储TOTP密钥（建议加密后存储或KMS托管）；
-- ===========================================================
CREATE TABLE users_auth (
  user_id                   BIGINT UNSIGNED NOT NULL COMMENT '用户ID=users.id',
  password_hash             VARCHAR(255)    NULL     COMMENT '密码哈希（argon2id/bcrypt编码串）',
  password_algo             ENUM('argon2id','bcrypt','scrypt') NOT NULL DEFAULT 'argon2id' COMMENT '密码哈希算法',
  mfa_enabled               TINYINT(1)      NOT NULL DEFAULT 0 COMMENT '是否启用MFA',
  mfa_secret                VARCHAR(128)    NULL     COMMENT 'TOTP密钥（建议加密或KMS）',
  last_password_changed_at  TIMESTAMP       NULL     COMMENT '上次修改密码时间',

  created_at                TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at                TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (user_id),
  CONSTRAINT fk_users_auth_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
    ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户认证信息（密码与MFA）';

-- ===========================================================
-- 角色表：roles（系统/业务角色）
-- 说明：通过 user_roles 建立多对多关系
-- ===========================================================
CREATE TABLE roles (
  id          SMALLINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  code        VARCHAR(64)       NOT NULL COMMENT '角色唯一编码（如admin/user）',
  name        VARCHAR(64)       NOT NULL COMMENT '角色名称',
  description VARCHAR(255)      NULL     COMMENT '角色描述',
  is_system   TINYINT(1)        NOT NULL DEFAULT 0 COMMENT '是否系统内置角色',

  created_at  TIMESTAMP         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at  TIMESTAMP         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),
  UNIQUE KEY uk_roles_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色定义';

-- ===========================================================
-- 用户角色表：user_roles（多对多）
-- 说明：复合主键(user_id, role_id)；删除用户时联动清理
-- ===========================================================
CREATE TABLE user_roles (
  user_id     BIGINT UNSIGNED    NOT NULL COMMENT '用户ID=users.id',
  role_id     SMALLINT UNSIGNED  NOT NULL COMMENT '角色ID=roles.id',
  created_at  TIMESTAMP          NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',

  PRIMARY KEY (user_id, role_id),
  KEY         idx_user_roles_role_id (role_id),

  CONSTRAINT fk_user_roles_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
    ON UPDATE RESTRICT,

  CONSTRAINT fk_user_roles_role
    FOREIGN KEY (role_id) REFERENCES roles (id)
    ON DELETE RESTRICT
    ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户与角色绑定';

-- ===========================================================
-- 第三方账号表：oauth_accounts（多条/用户）
-- 说明：
--  - provider + provider_account_id 保证同源唯一；
--  - profile_json 可缓存外部资料快照；
-- ===========================================================
CREATE TABLE oauth_accounts (
  id                    BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  user_id               BIGINT UNSIGNED NOT NULL COMMENT '用户ID=users.id',
  provider              VARCHAR(50)     NOT NULL COMMENT '提供方（github/google/wechat等）',
  provider_account_id   VARCHAR(191)    NOT NULL COMMENT '提供方用户唯一ID',
  access_token          TEXT            NULL     COMMENT '访问令牌（建议加密存储或不落库）',
  refresh_token         TEXT            NULL     COMMENT '刷新令牌（建议加密存储或不落库）',
  token_type            VARCHAR(32)     NULL     COMMENT '令牌类型',
  expires_at            TIMESTAMP       NULL     COMMENT '访问令牌过期时间',
  scope                 VARCHAR(255)    NULL     COMMENT '授权 Scope',
  profile_json          JSON            NULL     COMMENT '第三方用户信息快照',

  created_at            TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
  updated_at            TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  PRIMARY KEY (id),
  UNIQUE KEY uk_oauth_provider_account (provider, provider_account_id),
  KEY        idx_oauth_user_id (user_id),

  CONSTRAINT fk_oauth_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
    ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='第三方登录账号绑定';

-- ===========================================================
-- 登录审计表：login_audit（记录成功/失败登录）
-- 说明：
--  - user_id 可为空（未知用户或失败尝试）；
--  - ip_address 使用 VARBINARY(16) 以兼容IPv4/IPv6（建议应用层转换）；
-- ===========================================================
CREATE TABLE login_audit (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  user_id      BIGINT UNSIGNED NULL     COMMENT '用户ID，可为空',
  identity     VARCHAR(254)    NULL     COMMENT '尝试使用的用户名/邮箱/手机号',
  provider     ENUM('password','oauth') NOT NULL COMMENT '登录方式',
  ip_address   VARBINARY(16)   NULL     COMMENT 'IP（网络序，IPv4/IPv6）',
  user_agent   VARCHAR(255)    NULL     COMMENT 'UA',
  success      TINYINT(1)      NOT NULL COMMENT '是否成功',
  error_code   VARCHAR(64)     NULL     COMMENT '失败错误码（可选）',

  created_at   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',

  PRIMARY KEY (id),
  KEY          idx_login_audit_user_id (user_id),
  KEY          idx_login_audit_ctime (created_at),

  CONSTRAINT fk_login_audit_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE SET NULL
    ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='登录审计';

-- ===========================================================
-- 校验/重置令牌表：verification_tokens
-- 说明：
--  - 仅存 token_hash（如 SHA-256(token)），不存明文；
--  - 可承载邮箱验证/重置密码/MFA 挑战等；
-- ===========================================================
CREATE TABLE verification_tokens (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  user_id      BIGINT UNSIGNED NOT NULL COMMENT '用户ID=users.id',
  token_type   ENUM('email_verify','password_reset','mfa_challenge') NOT NULL COMMENT '令牌类型',
  token_hash   VARBINARY(32)   NOT NULL COMMENT '令牌摘要（SHA-256）',
  expires_at   TIMESTAMP       NOT NULL COMMENT '过期时间',
  consumed_at  TIMESTAMP       NULL     COMMENT '消费时间（NULL表示未使用）',
  metadata     JSON            NULL     COMMENT '可选额外信息',

  created_at   TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',

  PRIMARY KEY (id),
  UNIQUE KEY uk_verification_token_hash (token_hash),
  KEY         idx_verification_user (user_id, token_type, expires_at),

  CONSTRAINT fk_verification_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE
    ON UPDATE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='校验/重置令牌（存哈希）';

-- ===========================================================
-- 默认角色初始化（可按需调整）
-- ===========================================================
INSERT INTO roles (code, name, description, is_system)
VALUES
  ('admin', '管理员', '系统管理员，具备系统级权限', 1),
  ('user',  '用户',   '普通用户，默认权限集合',   1)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  is_system = VALUES(is_system);

COMMIT;

-- ===========================================================
-- 附：可选优化建议（非DDL）
-- 1) 如登录审计量较大，可考虑对 login_audit 按月份进行 RANGE 分区，或独立库/冷热分层；
-- 2) oauth_accounts 中敏感令牌建议加密存储或仅缓存短期令牌；
-- 3) users_auth.mfa_secret 建议KMS/硬件加密后落库，或改为外部密管；
-- 4) 低于 8.0.13 的 MySQL 请移除 users.uid 的 DEFAULT(UUID())，改由应用层生成；
-- ===========================================================
