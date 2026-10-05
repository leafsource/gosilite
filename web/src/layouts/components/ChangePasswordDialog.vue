<template>
  <t-dialog
    v-model:visible="visible"
    :header="t('pages.changePassword.title')"
    :confirm-btn="{ content: t('pages.changePassword.submit'), loading, disabled: loading }"
    :cancel-btn="{ content: t('pages.changePassword.cancel'), disabled: loading }"
    :close-on-confirm="false"
    @confirm="onConfirm"
  >
    <t-form ref="formRef" :data="formData" :rules="rules" label-width="96px" :disabled="loading" @submit="onConfirm">
      <t-form-item name="currentPassword" :label="t('pages.changePassword.currentPassword')">
        <t-input v-model="formData.currentPassword" type="password" autocomplete="off" />
      </t-form-item>
      <t-form-item name="newPassword" :label="t('pages.changePassword.newPassword')">
        <t-input v-model="formData.newPassword" type="password" autocomplete="new-password" />
      </t-form-item>
      <t-form-item name="confirmPassword" :label="t('pages.changePassword.confirmPassword')">
        <t-input v-model="formData.confirmPassword" type="password" autocomplete="new-password" />
      </t-form-item>
    </t-form>
  </t-dialog>
</template>
<script setup lang="ts">
import type { FormInstanceFunctions, FormRule } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import { changePassword } from '@/api/user';
import { t } from '@/locales';
import { useUserStore } from '@/store';

const userStore = useUserStore();
const router = useRouter();

const visible = ref(false);
const loading = ref(false);
const formRef = ref<FormInstanceFunctions>();

const INITIAL_DATA = {
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
};

const formData = ref({ ...INITIAL_DATA });

const rules = computed<Record<string, FormRule[]>>(() => ({
  currentPassword: [{ required: true, message: t('pages.changePassword.currentPassword'), type: 'error' }],
  newPassword: [
    { required: true, message: t('pages.changePassword.newPassword'), type: 'error' },
    { min: 8, message: t('pages.changePassword.tooShort'), type: 'error' },
  ],
  confirmPassword: [
    { required: true, message: t('pages.changePassword.confirmPassword'), type: 'error' },
    {
      validator: (val: string) => val === formData.value.newPassword,
      message: t('pages.changePassword.mismatch'),
      type: 'error',
    },
  ],
}));

const open = () => {
  formData.value = { ...INITIAL_DATA };
  visible.value = true;
};

const onConfirm = async () => {
  if (loading.value) {
    return;
  }
  const validateResult = await formRef.value?.validate();
  if (validateResult !== true) {
    return;
  }
  loading.value = true;
  try {
    await changePassword({
      currentPassword: formData.value.currentPassword,
      newPassword: formData.value.newPassword,
    });
    MessagePlugin.success(t('pages.changePassword.success'));
    visible.value = false;
    // 改密会撤销该用户现有 refresh token，这里直接结束会话回登录页
    await userStore.logout();
    router.push('/login');
  } catch (error) {
    MessagePlugin.error((error as Error).message);
  } finally {
    loading.value = false;
  }
};

defineExpose({ open });
</script>
