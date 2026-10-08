import { useLocalStorage } from '@vueuse/core';
import type { GlobalConfigProvider } from 'tdesign-vue-next';
import { computed } from 'vue';

import type { SupportedLocale } from '@/locales/index';
import { i18n, localeConfigKey, supportedLocales } from '@/locales/index';

export function useLocale() {
  const locale = computed({
    get: () => i18n.global.locale.value,
    set: (val: string) => {
      i18n.global.locale.value = val;
    },
  });
  const storedLocale = useLocalStorage<SupportedLocale>(localeConfigKey, 'zh_CN');

  const changeLocale = (lang: string) => {
    const validLang = supportedLocales.includes(lang as SupportedLocale) ? (lang as SupportedLocale) : 'zh_CN';
    locale.value = validLang;
    storedLocale.value = validLang;
    // 通知中心为服务端数据（标题原文），语言切换无需再刷新本地 mock 消息
  };

  const getComponentsLocale = computed(() => {
    return (i18n.global.getLocaleMessage(locale.value) as Record<string, any>).componentsLocale as GlobalConfigProvider;
  });

  return {
    changeLocale,
    getComponentsLocale,
    locale,
  };
}
