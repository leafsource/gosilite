<template>
  <div class="user-page">
    <t-row :gutter="[24, 24]">
      <t-col :flex="1">
        <t-card class="profile-card" :bordered="false">
          <div class="profile-card__avatar">
            <s-upload-image
              :action="UPLOAD_AVATAR_ACTION"
              field="avatar"
              :url="avatarUrl"
              size="96px"
              :upload-text="t('pages.user.uploadAvatar')"
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
              <s-tag-dict
                v-if="profile?.state"
                dict-type="sys_state"
                :value="profile.state"
              />
              <span v-else>{{ noValue }}</span>
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.createdAt')">
              {{ profile?.createdAt || noValue }}
            </t-descriptions-item>
            <t-descriptions-item :label="t('pages.user.roles')" :span="2">
              <template v-if="roleTags.length">
                <t-tag
                  v-for="role in roleTags"
                  :key="role"
                  theme="primary"
                  variant="light"
                  class="info-card__role"
                >
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
import { computed, onMounted, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';

import { fetchProfile } from '@/api/user';
import type { ProfileResponse } from '@/api/user';
import ChangePasswordDialog from '@/layouts/components/ChangePasswordDialog.vue';
import SUploadImage from '@/components/s-upload-image/index.vue';
import STagDict from '@/components/s-tag-dict/index.vue';
import { t } from '@/locales';
import { useUserStore } from '@/store';

defineOptions({ name: 'UserIndex' });

const UPLOAD_AVATAR_ACTION = '/user/update/upload_avatar';
const noValue = '-';

const userStore = useUserStore();
const profile = ref<ProfileResponse | null>(null);
const changePwdRef = ref<{ open: () => void }>();

/** 头像单源：登录态 userInfo（上传成功后由 getUserInfo 刷新） */
const avatarUrl = computed(() => userStore.userInfo.avatarURL || '');
const displayName = computed(
  () => profile.value?.username || userStore.userInfo.name || noValue,
);
const roleTags = computed(() => userStore.userInfo.roles ?? []);

async function loadProfile() {
  try {
    profile.value = await fetchProfile();
  } catch {
    // 资料接口异常时保底展示登录态已有信息
    profile.value = null;
  }
}

async function onAvatarSuccess() {
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
