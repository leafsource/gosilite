-- 0003_system_dict_param.sql
-- 字典 + 参数管理
BEGIN;

CREATE TABLE IF NOT EXISTS system_dict_type (
    id          BIGSERIAL PRIMARY KEY,
    dict_name   VARCHAR(64) NOT NULL,
    dict_type   VARCHAR(64) NOT NULL,
    status      SMALLINT NOT NULL DEFAULT 0,
    order_num   INT NOT NULL DEFAULT 0,
    remark      VARCHAR(255) NOT NULL DEFAULT '',
    del_flag    SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_dict_type_type ON system_dict_type(dict_type) WHERE del_flag = 0;

CREATE TABLE IF NOT EXISTS system_dict_data (
    id          BIGSERIAL PRIMARY KEY,
    dict_type   VARCHAR(64) NOT NULL,
    dict_label  VARCHAR(64) NOT NULL,
    dict_value  VARCHAR(64) NOT NULL,
    css_class   VARCHAR(64) NOT NULL DEFAULT '',
    list_order  INT NOT NULL DEFAULT 0,
    status      SMALLINT NOT NULL DEFAULT 0,
    is_default  SMALLINT NOT NULL DEFAULT 0,
    del_flag    SMALLINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_system_dict_data_type ON system_dict_data(dict_type);

CREATE TABLE IF NOT EXISTS system_config (
    id            BIGSERIAL PRIMARY KEY,
    config_name   VARCHAR(64) NOT NULL,
    config_key    VARCHAR(128) NOT NULL,
    config_value  TEXT NOT NULL DEFAULT '',
    config_type   VARCHAR(16) NOT NULL DEFAULT 'system',
    remark        VARCHAR(255) NOT NULL DEFAULT '',
    status        SMALLINT NOT NULL DEFAULT 0,
    del_flag      SMALLINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_config_key ON system_config(config_key) WHERE del_flag = 0;

-- 种子字典
INSERT INTO system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '状态', 'sys_status', 1, '通用状态'
WHERE NOT EXISTS (SELECT 1 FROM system_dict_type WHERE dict_type = 'sys_status');

INSERT INTO system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_status', '启用', '0', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_dict_data WHERE dict_type = 'sys_status' AND dict_value = '0');
INSERT INTO system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_status', '停用', '1', 2
WHERE NOT EXISTS (SELECT 1 FROM system_dict_data WHERE dict_type = 'sys_status' AND dict_value = '1');

INSERT INTO system_dict_type (dict_name, dict_type, order_num, remark)
SELECT '性别', 'sys_user_sex', 2, '用户性别'
WHERE NOT EXISTS (SELECT 1 FROM system_dict_type WHERE dict_type = 'sys_user_sex');
INSERT INTO system_dict_data (dict_type, dict_label, dict_value, list_order, is_default)
SELECT 'sys_user_sex', '男', '0', 1, 1
WHERE NOT EXISTS (SELECT 1 FROM system_dict_data WHERE dict_type = 'sys_user_sex' AND dict_value = '0');
INSERT INTO system_dict_data (dict_type, dict_label, dict_value, list_order)
SELECT 'sys_user_sex', '女', '1', 2
WHERE NOT EXISTS (SELECT 1 FROM system_dict_data WHERE dict_type = 'sys_user_sex' AND dict_value = '1');

-- 种子参数
INSERT INTO system_config (config_name, config_key, config_value, config_type, remark)
SELECT '系统名称', 'sys.app.name', 'gosilite', 'system', '站点名称'
WHERE NOT EXISTS (SELECT 1 FROM system_config WHERE config_key = 'sys.app.name');
INSERT INTO system_config (config_name, config_key, config_value, config_type, remark)
SELECT '分页默认大小', 'sys.page.size', '10', 'system', '列表默认 pageSize'
WHERE NOT EXISTS (SELECT 1 FROM system_config WHERE config_key = 'sys.page.size');

COMMIT;
