import { defineStore } from 'pinia';
import { computed, watch } from 'vue';

import { dictOptions } from '@/api/system/dict';
import type { DictOption } from '@/api/system/dictModel';

/**
 * 字典缓存 store：按 dictType 拉取 /system/dict/data/options 并缓存。
 * 拉取失败不写 loaded（下次 ensure 重试），调用方经 useDict 走硬编码兜底。
 */
export const useDictStore = defineStore('dict', {
  state: () => ({
    optionsMap: {} as Record<string, DictOption[]>,
    loaded: {} as Record<string, boolean>,
    pending: {} as Record<string, boolean>,
  }),
  actions: {
    async ensure(dictType: string) {
      if (!dictType || this.loaded[dictType] || this.pending[dictType]) return;
      this.pending[dictType] = true;
      try {
        const res = await dictOptions(dictType);
        this.optionsMap[dictType] = res.list || [];
        this.loaded[dictType] = true;
      } catch {
        // 失败不缓存：保留 fallback 展示，下次 ensure 重试
      } finally {
        this.pending[dictType] = false;
      }
    },
    /** 字典数据变更后失效缓存；不传 dictType 则全部失效 */
    invalidate(dictType?: string) {
      if (dictType) {
        this.loaded[dictType] = false;
        return;
      }
      this.loaded = {};
    },
  },
});

/** 各字典的硬编码兜底（与 0005 种子一致）：字典未返回/失败时保证页面文案不空白 */
const DEFAULT_FALLBACK: Record<string, DictOption[]> = {
  sys_status: [
    { label: '启用', value: '0' },
    { label: '停用', value: '1' },
  ],
  sys_state: [
    { label: '启用', value: 'enabled' },
    { label: '停用', value: 'disabled' },
  ],
  sys_data_scope: [
    { label: '全部数据', value: '1' },
    { label: '本部门', value: '2' },
    { label: '本部门及以下', value: '3' },
    { label: '仅本人', value: '4' },
    { label: '自定义部门', value: '5' },
  ],
  sys_menu_type: [
    { label: '目录', value: 'M' },
    { label: '菜单', value: 'C' },
    { label: '按钮', value: 'F' },
  ],
  sys_config_type: [
    { label: '系统', value: 'system' },
    { label: '业务', value: 'business' },
    { label: '全局', value: 'global' },
  ],
};

/**
 * 组合式字典：挂载时触发拉取；字典未返回/失败时用兜底项
 * （默认取 DEFAULT_FALLBACK[dictType]，可传参覆盖）。
 * 返回 options 与 labelOf（未命中回落原值字符串）。
 */
export function useDict(dictType: string, fallback: DictOption[] = DEFAULT_FALLBACK[dictType] ?? []) {
  const dictStore = useDictStore();
  void dictStore.ensure(dictType);
  // invalidate 置 loaded=false 后自动重拉：keep-alive 页面 setup 不再执行也能刷新
  watch(
    () => dictStore.loaded[dictType],
    (isLoaded) => {
      if (!isLoaded) void dictStore.ensure(dictType);
    },
  );

  const options = computed(() => {
    const cached = dictStore.optionsMap[dictType];
    return cached && cached.length ? cached : fallback;
  });

  const labelOf = (value: string | number | undefined | null): string => {
    if (value === undefined || value === null || value === '') return '';
    const hit = options.value.find((o) => String(o.value) === String(value));
    return hit ? hit.label : String(value);
  };

  return { options, labelOf };
}
