<template>
  <div class="user-page">
    <t-row :gutter="[24, 24]">
      <t-col :flex="1">
        <t-card class="profile-card" :bordered="false">
          <div class="profile-card__avatar">
            <!-- limit=1：单头像磁贴（达到 max 时隐藏新增框，避免多余入口）；
                 换头像流程 = hover 删除（X）→ 新增框出现 → 上传新图（服务端无清空头像接口，
                 删除仅清本地，刷新即恢复；上传失败回退服务端原头像） -->
            <s-upload-image
              v-model="avatarValue"
              :limit="1"
              :is-show-tip="false"
              :show-image-file-name="false"
              @success="onAvatarSuccess"
            />
          </div>
          <div class="profile-card__name">{{ displayName }}</div>
          <div class="profile-card__account">{{ profile?.account || userStore.userInfo.name }}</div>
          <div v-if="roleTags.length" class="profile-card__roles">
            <t-tag v-for="role in roleTags" :key="role" theme="primary" variant="light">
              {{ role }}
            </t-tag>
          </div>
          <t-button class="profile-card__action" variant="outline" @click="changePwdRef?.open()">
            {{ t('pages.user.changePassword') }}
          </t-button>
        </t-card>
      </t-col>

      <t-col :flex="3">
        <t-card class="info-card" :bordered="false" :title="t('pages.user.profileTitle')">
          <t-descriptions :column="2" item-layout="vertical">
            <t-descriptions-item :label="t('pages.user.account')">
              {{ profile?.account || noValue }}
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.username')">
              {{ displayName }}
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.email')">
              {{ profile?.email || noValue }}
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.mobile')">
              {{ profile?.mobilePhone || noValue }}
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.state')">
              <s-tag-dict v-if="profile?.state" dict-type="sys_state" :value="profile.state" />
              <span v-else>{{ noValue }}</span>
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.createdAt')">
              {{ profile?.createdAt || noValue }}
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.roles')" :span="2">
              <template v-if="roleTags.length">
                <t-tag v-for="role in roleTags" :key="role" theme="primary" variant="light" class="info-card__role">
                  {{ role }}
                </t-tag>
              </template>
              <span v-else>{{ noValue }}</span>
            </t-descriptions-item>
          </t-descriptions>
        </t-card>
      </t-col>
    </t-row>

    <change-password-dialog ref="changePwdRef" />
  </div>
</template>
<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref, watch } from 'vue';

import type { ProfileResponse } from '@/api/user';
import { fetchProfile, updateProfileAvatar } from '@/api/user';
import STagDict from '@/components/s-tag-dict/index.vue';
import type { SUploadSuccessFile } from '@/components/s-upload-image/index.vue';
import SUploadImage from '@/components/s-upload-image/index.vue';
import ChangePasswordDialog from '@/layouts/components/ChangePasswordDialog.vue';
import { t } from '@/locales';
import { useUserStore } from '@/store';

defineOptions({ name: 'UserIndex' });

const noValue = '-';

const userStore = useUserStore();
const profile = ref<ProfileResponse | null>(null);
const changePwdRef = ref<{ open: () => void }>();

/** 头像单源：登录态 userInfo（上传成功后由 getUserInfo 刷新） */
const avatarUrl = computed(() => userStore.userInfo.avatarURL || '');
/** 组件 v-model 本地值：上传/移除即时反馈，avatarUrl 变化时回写对齐服务端 */
const avatarValue = ref('');
watch(
  avatarUrl,
  (v) => {
    avatarValue.value = v;
  },
  { immediate: true },
);
const displayName = computed(() => profile.value?.username || userStore.userInfo.name || noValue);
const roleTags = computed(() => userStore.userInfo.roles ?? []);

async function loadProfile() {
  try {
    profile.value = await fetchProfile();
  } catch {
    // 资料接口异常时保底展示登录态已有信息
    profile.value = null;
  }
}

// 统一文件服务两步链路（spec S2.11）：common 上传拿 fileId → 按 id 更新头像（服务端回查台账）
async function onAvatarSuccess(files: SUploadSuccessFile[]) {
  const fileId = files[0]?.ossId;
  if (!fileId) {
    MessagePlugin.error(t('pages.user.uploadFailed'));
    return;
  }
  try {
    await updateProfileAvatar(fileId);
  } catch (e) {
    MessagePlugin.error((e as Error).message || t('pages.user.uploadFailed'));
    // 更新失败：回退展示服务端原头像，避免本地显示与实际不一致
    avatarValue.value = avatarUrl.value;
    return;
  }
  await Promise.all([userStore.getUserInfo(), loadProfile()]);
  MessagePlugin.success(t('pages.user.avatarUpdated'));
}

onMounted(() => {
  loadProfile();
  // 兜底刷新登录态（头像/角色），幂等
  userStore.getUserInfo().catch(() => undefined);
});
</script>
<style lang="less" scoped>
.user-page {
  width: 100%;
}

.profile-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--td-comp-paddingTB-l, 24px) var(--td-comp-paddingLR-l, 24px);

  &__avatar {
    margin-bottom: var(--td-comp-margin-l, 16px);

    // t-upload 图片列表为 flex + gap：Vue 模板的空白文本节点会成为 flex item 多占一格 gap，
    // 单磁贴 shrink-wrap 时被右侧 gap 拉偏（实测相对卡片中心左偏 8px）；gap 置 0 后单图自然居中
    :deep(.t-upload__card) {
      gap: 0;
      justify-content: center;
    }
  }

  &__name {
    font-size: var(--td-font-size-title-large, 20px);
    font-weight: 600;
    color: var(--td-text-color-primary);
  }

  &__account {
    margin-top: 4px;
    font-size: var(--td-font-size-body-small, 12px);
    color: var(--td-text-color-secondary);
  }

  &__roles {
    display: flex;
    flex-wrap: wrap;
    gap: var(--td-comp-margin-s, 8px);
    justify-content: center;
    margin-top: var(--td-comp-margin-m, 12px);
  }

  &__action {
    width: 100%;
    margin-top: var(--td-comp-margin-xl, 24px);
  }
}

.info-card {
  :deep(.t-descriptions__content) {
    color: var(--td-text-color-primary);
  }
}

.info-card__role {
  margin-right: var(--td-comp-margin-s, 8px);
}
</style>
