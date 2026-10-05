import { defineStore } from 'pinia';

import { fetchUserInfo, login as loginApi, logout as logoutApi, refreshSession } from '@/api/user';
import { t } from '@/locales';
import { usePermissionStore } from '@/store';
import type { UserInfo } from '@/types/interface';

const InitUserInfo: UserInfo = {
  name: '', // 用户名，用于展示在页面右上角头像处
  roles: [], // 前端权限模型使用 如果使用请配置modules/permission-fe.ts使用
};

// 单飞的 refresh：并发 401 共享同一次刷新，避免轮换中的 refresh token 被重复使用
let refreshPromise: Promise<string> | null = null;

async function redirectToLogin() {
  const { default: router } = await import('@/router');
  const current = router.currentRoute.value;
  if (current.path !== '/login') {
    router.push({ path: '/login', query: { redirect: current.fullPath } });
  }
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    refreshToken: '',
    userInfo: { ...InitUserInfo },
  }),
  getters: {
    roles: (state) => {
      return state.userInfo?.roles;
    },
  },
  actions: {
    setSession(tokens: { accessToken: string; refreshToken: string }) {
      this.token = tokens.accessToken;
      this.refreshToken = tokens.refreshToken;
    },
    clearSession() {
      this.token = '';
      this.refreshToken = '';
      this.userInfo = { ...InitUserInfo };
    },
    async login(userInfo: Record<string, unknown>) {
      const { account, password } = userInfo as { account: string; password: string };
      const res = await loginApi({ username: account, password });
      this.setSession(res);
    },
    async getUserInfo() {
      const res = await fetchUserInfo();
      this.userInfo = {
        ...InitUserInfo,
        name: res.username || res.account,
      };
    },
    async refreshAccessToken(): Promise<string> {
      if (!this.refreshToken) {
        this.clearSession();
        await redirectToLogin();
        throw new Error(t('pages.login.expired'));
      }
      if (!refreshPromise) {
        const currentRefreshToken = this.refreshToken;
        refreshPromise = (async () => {
          try {
            const res = await refreshSession(currentRefreshToken);
            this.setSession(res);
            return res.accessToken;
          } catch (error) {
            this.clearSession();
            await redirectToLogin();
            throw error;
          } finally {
            refreshPromise = null;
          }
        })();
      }
      return refreshPromise;
    },
    async logout() {
      const { token, refreshToken } = this;
      if (token || refreshToken) {
        try {
          await logoutApi(refreshToken || undefined);
        } catch {
          // 退出接口失败不阻塞本地会话清理
        }
      }
      this.clearSession();
    },
  },
  persist: {
    afterHydrate: () => {
      const permissionStore = usePermissionStore();
      permissionStore.initRoutes();
    },
    key: 'user',
    pick: ['token', 'refreshToken'],
  },
});
