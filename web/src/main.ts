/* eslint-disable simple-import-sort/imports */
import { createApp } from 'vue';
import TDesign from 'tdesign-vue-next';

import App from './App.vue';
import router from './router';
import { getConfigStore, store } from './store';
import i18n from './locales';
import vPermission from './directives/permission';

import 'tdesign-vue-next/es/style/index.css';
import '@/style/index.less';
import './permission';

const app = createApp(App);

app.use(TDesign);
app.use(store);
app.use(router);
app.use(i18n);
app.directive('permission', vPermission);

// 启动即拉取全局参数（匿名）：系统名称 → document.title 等
void getConfigStore().loadGlobal();

app.mount('#app');
