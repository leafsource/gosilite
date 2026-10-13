import type { Directive } from 'vue';

import { useUserStore } from '@/store';

// 超管角色：与后端 ctxutil.IsSuperRole 保持一致
const SUPER_ROLES = ['admin', 'system'];

export function hasPerm(value: string | string[]): boolean {
  const userStore = useUserStore();
  if ((userStore.roles ?? []).some((role) => SUPER_ROLES.includes(role))) {
    return true;
  }
  const perms = userStore.permissions ?? [];
  const required = Array.isArray(value) ? value : [value];
  // 数组 = 任一命中即保留
  return required.some((code) => perms.includes(code));
}

/**
 * 按钮级权限指令：v-permission="'system:user:add'" 或 v-permission="['a','b']"
 * 不满足权限码时从 DOM 移除该元素。
 */
export const vPermission: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    if (binding.value && !hasPerm(binding.value)) {
      el.parentNode?.removeChild(el);
    }
  },
};

export default vPermission;
