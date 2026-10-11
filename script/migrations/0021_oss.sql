-- 0021: 对象存储（OSS配置 / 文件管理 / OSS处理规则）
--   新增三表 system_oss_config / system_oss / system_oss_rule；
--   菜单「对象存储」M id=60（parent=3 系统管理，复用 path storage）
--     + C 61 OSS配置 / 62 文件管理 / 63 OSS处理规则 + F 按钮 6101-6104 / 6201-6206 / 6301-6304
--     + system/admin 角色绑定；删除旧「存储管理」菜单 48/4801/4802 及其角色绑定（能力被吸收）。
--   种子：system_oss_config 插入 config_key='local' 模板（status='0' 非默认，运行时回退 conf.yaml）；
--         system_oss_rule 预置阿里/腾讯/华为常用图片处理规则 + 本地停用示例。
-- 幂等：重复执行不产生重复行/列；全表 public. 限定。
-- spec: docs/compose/spec/system-oss.md S2.1 S2.2

BEGIN;

-- ========== 对象存储配置表（字段照抄 ruoyi sys_oss_config，审计用本仓惯例） ==========
CREATE TABLE IF NOT EXISTS public.system_oss_config (
    id            BIGSERIAL PRIMARY KEY,
    config_key    VARCHAR(64)  NOT NULL,
    access_key    VARCHAR(255) NOT NULL DEFAULT '',
    secret_key    VARCHAR(255) NOT NULL DEFAULT '',
    bucket_name   VARCHAR(255) NOT NULL,
    prefix        VARCHAR(255) NOT NULL DEFAULT '',
    endpoint      VARCHAR(255) NOT NULL,
    domain        VARCHAR(255) NOT NULL DEFAULT '',
    is_https      CHAR(1)      NOT NULL DEFAULT 'N',
    region        VARCHAR(64)  NOT NULL DEFAULT '',
    access_policy CHAR(1)      NOT NULL DEFAULT '1',  -- 0=private 1=public 2=custom
    status        CHAR(1)      NOT NULL DEFAULT '0',  -- 是否默认：1=默认（全表至多一条）0=否
    ext1          VARCHAR(255) NOT NULL DEFAULT '',
    remark        VARCHAR(500) NOT NULL DEFAULT '',
    del_flag      SMALLINT     NOT NULL DEFAULT 0,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_system_oss_config_key UNIQUE (config_key)
);
COMMENT ON TABLE public.system_oss_config IS '对象存储配置';
CREATE INDEX IF NOT EXISTS idx_oss_config_status ON public.system_oss_config(status) WHERE del_flag = 0;

-- ========== 文件台账表（照抄 ruoyi sys_oss，去租户/分类，保留 is_lock/service） ==========
CREATE TABLE IF NOT EXISTS public.system_oss (
    id            BIGSERIAL PRIMARY KEY,
    file_name     VARCHAR(255) NOT NULL,               -- 对象 key
    original_name VARCHAR(255) NOT NULL,               -- 上传原始文件名
    file_suffix   VARCHAR(32)  NOT NULL DEFAULT '',    -- 小写后缀（不带点）
    url           VARCHAR(500) NOT NULL,               -- 上传完成时按 domain/is_https 拼的访问 URL
    size          BIGINT       NOT NULL DEFAULT 0,
    content_type  VARCHAR(128) NOT NULL DEFAULT '',
    is_lock       BOOLEAN      NOT NULL DEFAULT FALSE, -- 锁定文件禁止删除
    service       VARCHAR(64)  NOT NULL DEFAULT '',    -- 上传时使用的 config_key
    create_by     VARCHAR(64)  NOT NULL DEFAULT '',    -- 上传人
    del_flag      SMALLINT     NOT NULL DEFAULT 0,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_oss IS 'OSS文件台账';
CREATE INDEX IF NOT EXISTS idx_system_oss_create_by ON public.system_oss(create_by, created_at) WHERE del_flag = 0;
CREATE INDEX IF NOT EXISTS idx_system_oss_original ON public.system_oss(original_name) WHERE del_flag = 0;

-- ========== OSS处理规则表（照抄 ruoyi sys_oss_rule） ==========
CREATE TABLE IF NOT EXISTS public.system_oss_rule (
    id           BIGSERIAL PRIMARY KEY,
    rule_name    VARCHAR(255) NOT NULL,                -- 如 80x80，输出字段 url_80x80
    domain       VARCHAR(255) NOT NULL,                -- 匹配域名（URL contains 命中）
    mime_type    VARCHAR(255) NOT NULL DEFAULT '',     -- 逗号分隔；contentType 或文件后缀匹配
    rule         TEXT         NOT NULL,                -- #{#var} 占位：domain/path/fullPath/filename/url/fullUrl/query
    is_overwrite CHAR(1)      NOT NULL DEFAULT 'N',    -- Y=覆盖主 url
    is_default   CHAR(1)      NOT NULL DEFAULT 'N',    -- 默认规则
    status       CHAR(1)      NOT NULL DEFAULT '1',    -- 1 启用 0 停用
    rule_sort    INTEGER      NOT NULL DEFAULT 0,
    remark       VARCHAR(500) NOT NULL DEFAULT '',
    del_flag     SMALLINT     NOT NULL DEFAULT 0,
    created_at   TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP    NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_oss_rule IS 'OSS处理规则';
CREATE UNIQUE INDEX IF NOT EXISTS uk_oss_rule_domain_name ON public.system_oss_rule(domain, rule_name) WHERE del_flag = 0;
CREATE INDEX IF NOT EXISTS idx_oss_rule_status ON public.system_oss_rule(status) WHERE del_flag = 0;

-- ========== OSS处理规则常用种子（spec system-oss S2.6 修订：预置各云常用配置）==========
-- 全部 is_overwrite=N：只产出变体字段（url_180x800 等），绝不改写主 URL
--   （主 URL 是 private 桶预签名，拼接 query 会破坏签名导致 AccessDenied）。
-- 阿里/腾讯/华为按各自处理域名匹配，本地 URL 不含这些域名 → 不会误命中。
-- 本地示例默认停用：RustFS/S3 不处理图片参数，带未知 query 的 GET 会被 S3 拒绝，
--   仅作规则引擎调试模板（对接 CDN/图片网关后可启用）。
INSERT INTO public.system_oss_rule (rule_name, domain, mime_type, rule, is_overwrite, is_default, status, rule_sort, remark)
SELECT v.rule_name, v.domain, v.mime_type, v.rule, 'N', 'N', v.status, v.sort, v.remark
FROM (VALUES
    -- 阿里云 OSS（数据万象 x-oss-process）
    ('180x180', 'aliyuncs.com', 'image',
     '#{#url}?x-oss-process=image/auto-orient,1/resize,m_lfit,w_180,h_180/quality,q_90',
     '1', 10, '阿里云 180 缩略（头像/列表小图）'),
    ('800x800', 'aliyuncs.com', 'image',
     '#{#url}?x-oss-process=image/auto-orient,1/resize,m_lfit,w_800,h_800/quality,q_90',
     '1', 11, '阿里云 800 展示图'),
    -- 腾讯云 COS（数据万象 imageMogr2）
    ('180x180', 'myqcloud.com', 'image',
     '#{#url}?imageMogr2/thumbnail/180x180/quality/90',
     '1', 20, '腾讯云 COS 180 缩略'),
    ('800x800', 'myqcloud.com', 'image',
     '#{#url}?imageMogr2/thumbnail/800x800/quality/90',
     '1', 21, '腾讯云 COS 800 展示图'),
    -- 华为云 OBS（x-obs-process）
    ('180x180', 'myhuaweicloud.com', 'image',
     '#{#url}?x-obs-process=image/resize,m_lfit,w_180,h_180/quality,q_90',
     '1', 30, '华为云 OBS 180 缩略'),
    ('800x800', 'myhuaweicloud.com', 'image',
     '#{#url}?x-obs-process=image/resize,m_lfit,w_800,h_800/quality,q_90',
     '1', 31, '华为云 OBS 800 展示图'),
    -- 本地（RustFS/S3 无图片处理；停用示例，接 CDN/图片网关后按需启用）
    ('local-demo', '127.0.0.1', 'image',
     '#{#url}?x-oss-process=image/resize,m_lfit,w_800,h_800/quality,q_90',
     '0', 90, '本地示例（默认停用）：原生 S3 不处理图片参数，带 query 会 AccessDenied；对接 CDN/图片网关后启用')
) AS v(rule_name, domain, mime_type, rule, status, sort, remark)
WHERE NOT EXISTS (
    SELECT 1 FROM public.system_oss_rule r
    WHERE r.domain = v.domain AND r.rule_name = v.rule_name AND r.del_flag = 0
);

-- ========== 配置模板种子（status=0 非默认：无默认行时运行时回退 conf.yaml；密钥与本地 dev conf 一致开箱可用，
--            生产在页面改密钥后「设为默认」） ==========
INSERT INTO public.system_oss_config
    (config_key, access_key, secret_key, bucket_name, prefix, endpoint, domain, is_https, region, access_policy, status, remark)
SELECT 'local', 'gosilite', 'gosilite-secret', 'skyx', '', 'http://127.0.0.1:19000', '127.0.0.1:19000', 'N', '', '0', '0',
       '本地 RustFS 模板：与 conf.yaml dev 默认一致；生产改密钥后在页面「设为默认」生效'
WHERE NOT EXISTS (SELECT 1 FROM public.system_oss_config WHERE config_key = 'local');

-- ========== 拆除旧「存储管理」菜单（能力吸收进 文件管理 62） ==========
DELETE FROM public.system_role_menu WHERE menu_id IN (48, 4801, 4802);
DELETE FROM public.system_menu WHERE id IN (48, 4801, 4802);

-- ========== M 对象存储（父级系统管理 id=3；60-63 已核对空闲） ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 60, 3, 'page:root:storage', 'M', '对象存储', 'storage',
       (SELECT COALESCE(MAX(sort), 0) + 1 FROM public.system_menu WHERE parent_id = 3),
       'layers', 'enabled', 'layout', 'show', '', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 60);

-- ========== C 三菜单 ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 61, 60, 'system:ossConfig:list', 'C', 'OSS配置', 'ossConfig', 1, 'file-code', 'enabled', 'system/ossConfig/index', 'show', 'system:ossConfig:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 61)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 60);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 62, 60, 'system:oss:list', 'C', '文件管理', 'oss', 2, 'folder', 'enabled', 'system/oss/index', 'show', 'system:oss:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 62)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 60);
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, icon, state, component, visible, perms, status)
SELECT 63, 60, 'system:ossRule:list', 'C', 'OSS处理规则', 'ossRule', 3, 'catalog', 'enabled', 'system/ossRule/index', 'show', 'system:ossRule:list', 0
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = 63)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 60);

-- ========== F 按钮：OSS配置 61xx ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT v.id, 61, v.code, 'F', v.title, '', v.ord, v.code, 'enabled', 'hide', 0
FROM (VALUES
    (6101, 'system:ossConfig:add',    '配置新增', 1),
    (6102, 'system:ossConfig:edit',   '配置修改', 2),
    (6103, 'system:ossConfig:remove', '配置删除', 3),
    (6104, 'system:ossConfig:test',   '配置测试', 4)
) AS v(id, code, title, ord)
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = v.id)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 61);

-- ========== F 按钮：文件管理 62xx ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT v.id, 62, v.code, 'F', v.title, '', v.ord, v.code, 'enabled', 'hide', 0
FROM (VALUES
    (6203, 'system:oss:edit',     '文件修改', 1),
    (6204, 'system:oss:remove',   '文件删除', 2),
    (6205, 'system:oss:lock',     '文件加锁解锁', 3)
) AS v(id, code, title, ord)
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = v.id)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 62);

-- ========== F 按钮：OSS处理规则 63xx ==========
INSERT INTO public.system_menu (id, parent_id, permission_code, menu_type, title, path, sort, perms, state, visible, status)
SELECT v.id, 63, v.code, 'F', v.title, '', v.ord, v.code, 'enabled', 'hide', 0
FROM (VALUES
    (6301, 'system:ossRule:add',     '规则新增', 1),
    (6302, 'system:ossRule:edit',    '规则修改', 2),
    (6303, 'system:ossRule:remove',  '规则删除', 3),
    (6304, 'system:ossRule:refresh', '刷新缓存', 4)
) AS v(id, code, title, ord)
WHERE NOT EXISTS (SELECT 1 FROM public.system_menu WHERE id = v.id)
  AND EXISTS (SELECT 1 FROM public.system_menu WHERE id = 63);

-- ========== 管理员角色绑定 ==========
INSERT INTO public.system_role_menu (role_id, menu_id)
SELECT r.id, m.id
FROM public.system_role r
CROSS JOIN (VALUES
    (60), (61), (6101), (6102), (6103), (6104),
    (62), (6203), (6204), (6205),
    (63), (6301), (6302), (6303), (6304)
) AS m(id)
WHERE r.code IN ('system', 'admin')
  AND NOT EXISTS (
    SELECT 1 FROM public.system_role_menu rm WHERE rm.role_id = r.id AND rm.menu_id = m.id
  );

COMMIT;
