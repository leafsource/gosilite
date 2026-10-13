import { defineStore } from 'pinia';

import { getConfigByKey } from '@/api/system/config';
import { request } from '@/utils/request';

/**
 * 参数消费 store：
 * - loadGlobal：匿名拉取 /system/config/global（品牌类参数，如 sys.app.name），启动即调；
 * - loadPrivate：登录后拉取需登录的参数（如 sys.page.size），会话内只跑一次。
 * 失败不置 loaded，下次调用可重试；调用方以硬编码兜底（appName || t(...)、pageDefault=10）。
 */
export const useConfigStore = defineStore('config', {
  state: () => ({
    values: {} as Record<string, string>,
    globalLoaded: false,
    globalPending: false,
    privateLoaded: false,
    privatePending: false,
  }),
  getters: {
    /** 系统名称；空串表示未取到，调用方自行兜底 */
    appName: (state) => state.values['sys.app.name'] || '',
    /** 分页默认大小；参数缺失/非法（非正整数）时 10 */
    pageDefault: (state) => {
      const n = Number(state.values['sys.page.size']);
      return Number.isInteger(n) && n > 0 ? n : 10;
    },
  },
  actions: {
    async loadGlobal() {
      if (this.globalLoaded || this.globalPending) return;
      this.globalPending = true;
      try {
        const res = await request.get<{ list: Array<{ configKey: string; configValue: string }> }>(
          { url: '/system/config/global' },
          { withToken: false, skipAuthRefresh: true },
        );
        for (const item of res?.list || []) {
          this.values[item.configKey] = item.configValue;
        }
        this.globalLoaded = true;
        if (this.values['sys.app.name']) {
          document.title = this.values['sys.app.name'];
        }
      } catch {
        // 匿名拉取失败：保留静态 title 兜底，下次启动/调用重试
      } finally {
        this.globalPending = false;
      }
    },
    async loadPrivate() {
      if (this.privateLoaded || this.privatePending) return;
      this.privatePending = true;
      try {
        const res = await getConfigByKey('sys.page.size');
        if (res?.configKey) {
          this.values[res.configKey] = res.configValue;
        }
        this.privateLoaded = true;
      } catch {
        // 失败保持 pageDefault=10 兜底，下次调用重试
      } finally {
        this.privatePending = false;
      }
    },
  },
});

export function getConfigStore() {
  return useConfigStore();
}
