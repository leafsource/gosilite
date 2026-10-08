-- 0012: 通知中心消费侧——已读记录（notice-center）
--   新增 system_notice_read：user_id × notice_id 已读标记（未读 = 已发布 − 本人已读）。
--   全员可见（用户拍板）：无 receiver 表；撤回即用户侧不可见，已读行保留。
-- 幂等：重复执行不产生重复行/索引；全表 public. 限定。
-- spec: docs/compose/spec/notice-center.md S2.1

BEGIN;

CREATE TABLE IF NOT EXISTS public.system_notice_read (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    notice_id  BIGINT NOT NULL,
    read_at    TIMESTAMP NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE public.system_notice_read IS '通知已读记录（消费侧）';
CREATE UNIQUE INDEX IF NOT EXISTS uk_system_notice_read_user_notice
    ON public.system_notice_read(user_id, notice_id);
CREATE INDEX IF NOT EXISTS idx_system_notice_read_user
    ON public.system_notice_read(user_id);

COMMIT;
