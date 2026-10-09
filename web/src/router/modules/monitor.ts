import { PreciseMonitorIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';
import type { RouteRecordRaw } from 'vue-router';

import { LAYOUT } from '@/utils/route/constant';

// 系统监控：顶级菜单，位于系统管理下方
export default [
  {
    path: '/monitor',
    name: 'monitor',
    component: LAYOUT,
    redirect: '/monitor/online',
    meta: {
      title: { zh_CN: '系统监控', en_US: 'Monitor' },
      icon: shallowRef(PreciseMonitorIcon),
      orderNo: 4,
    },
    children: [
      {
        path: 'online',
        name: 'MonitorOnline',
        component: () => import('@/pages/monitor/online/index.vue'),
        meta: { title: { zh_CN: '在线用户', en_US: 'Online Users' } },
      },
      {
        path: 'data',
        name: 'MonitorData',
        component: () => import('@/pages/monitor/data/index.vue'),
        meta: { title: { zh_CN: '数据监控', en_US: 'Data Monitor' } },
      },
      {
        path: 'cache',
        name: 'MonitorCache',
        component: () => import('@/pages/monitor/cache/index.vue'),
        meta: { title: { zh_CN: '缓存监控', en_US: 'Cache Monitor' } },
      },
    ],
  },
] satisfies RouteRecordRaw[];
