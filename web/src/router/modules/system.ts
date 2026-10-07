import { SettingIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';
import type { RouteRecordRaw } from 'vue-router';

import { LAYOUT } from '@/utils/route/constant';

// 系统管理：与后端 system 菜单目录（部门/用户/角色/菜单/岗位）对应的静态路由
export default [
  {
    path: '/system',
    name: 'system',
    component: LAYOUT,
    redirect: '/system/contacts',
    meta: {
      title: { zh_CN: '系统管理', en_US: 'System' },
      icon: shallowRef(SettingIcon),
      orderNo: 3,
    },
    children: [
      {
        path: 'contacts',
        name: 'SystemContacts',
        component: () => import('@/pages/system/organization/index.vue'),
        meta: { title: { zh_CN: '组织架构', en_US: 'Organization' } },
      },
      {
        path: 'role',
        name: 'SystemRole',
        component: () => import('@/pages/system/role/index.vue'),
        meta: { title: { zh_CN: '角色管理', en_US: 'Role' } },
      },
      {
        path: 'menu',
        name: 'SystemMenu',
        component: () => import('@/pages/system/menu/index.vue'),
        meta: { title: { zh_CN: '菜单管理', en_US: 'Menu' } },
      },
      {
        path: 'post',
        name: 'SystemPost',
        component: () => import('@/pages/system/post/index.vue'),
        meta: { title: { zh_CN: '岗位管理', en_US: 'Post' } },
      },
      {
        path: 'dict',
        name: 'SystemDict',
        component: () => import('@/pages/system/dict/index.vue'),
        meta: { title: { zh_CN: '字典管理', en_US: 'Dictionary' } },
      },
      {
        path: 'config',
        name: 'SystemConfig',
        component: () => import('@/pages/system/config/index.vue'),
        meta: { title: { zh_CN: '参数管理', en_US: 'Config' } },
      },
    ],
  },
] satisfies RouteRecordRaw[];
