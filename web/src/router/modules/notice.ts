import type { RouteRecordRaw } from 'vue-router';

import { LAYOUT } from '@/utils/route/constant';

// 我的通知（通知中心消费侧）：入口 = 头部铃铛「查看全部」，不进侧栏
// （侧栏业务菜单以后端菜单为唯一来源，见 store/modules/permission.ts）
export default [
  {
    path: '/notice',
    component: LAYOUT,
    redirect: '/notice/list',
    name: 'notice',
    meta: {
      title: { zh_CN: '我的通知', en_US: 'My Notices' },
      orderNo: 10,
    },
    children: [
      {
        path: 'list',
        name: 'NoticeList',
        component: () => import('@/pages/notice/index.vue'),
        meta: { title: { zh_CN: '我的通知', en_US: 'My Notices' } },
      },
    ],
  },
] satisfies RouteRecordRaw[];
