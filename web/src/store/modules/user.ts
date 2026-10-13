import { defineStore } from 'pinia';

import { fetchUserInfo, login as loginApi, logout as logoutApi, refreshSession } from '@/api/user';
import { t } from '@/locales';
import { usePermissionStore } from '@/store';
import type { UserInfo } from '@/types/interface';

const InitUserInfo: UserInfo = {
  name: '', // 用户名，用于展示在页面右上角头像处
  avatarURL: '', // 头像地址，空串表示未设置（回退占位图标）
  roles: [], // 角色码（超管判定 admin/system）
  permissions: [], // 权限码，供 v-permission 指令使用
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
    permissions: (state) => {
      return state.userInfo?.permissions ?? [];
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
        name: res.user?.username || res.user?.account,
        avatarURL: res.user?.avatarURL || '',
        roles: res.roles ?? [],
        permissions: res.permissions ?? [],
      };
    },
    /**
     * 上传前确保 token 新鲜（t-upload 原生 XHR 绕过 axios 401→refresh 拦截器，
     * token 过期会直接 401）：解析本 token exp，距过期 <60s 才走单飞刷新。
     */
    async ensureFreshAccessToken(): Promise<string> {
      if (!this.token) {
        return this.refreshAccessToken();
      }
      try {
        const payload = JSON.parse(atob(this.token.split('.')[1] || ''));
        const expSec = Number(payload.exp || 0);
        if (expSec * 1000 - Date.now() > 60_000) {
          return this.token;
        }
      } catch {
        // 解析失败按临期处理，交给刷新
      }
      return this.refreshAccessToken();
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
          } catch {
            // 会话已失效（过期/被强退）：清理并回登录页，
            // 抛友好文案而非后端内部错误，避免 toast「Service Internal Error」
            this.clearSession();
            await redirectToLogin();
            throw new Error(t('pages.login.expired'));
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
