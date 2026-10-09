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
      {
        path: 'notice',
        name: 'SystemNotice',
        component: () => import('@/pages/system/notice/index.vue'),
        meta: { title: { zh_CN: '通知公告', en_US: 'Notice' } },
      },
      {
        path: 'log',
        name: 'SystemLog',
        redirect: '/system/log/operlog',
        meta: { title: { zh_CN: '日志管理', en_US: 'Logs' } },
        children: [
          {
            path: 'operlog',
            name: 'SystemOperLog',
            component: () => import('@/pages/system/log/operlog/index.vue'),
            meta: { title: { zh_CN: '操作日志', en_US: 'Operation Logs' } },
          },
          {
            path: 'loginlog',
            name: 'SystemLoginLog',
            component: () => import('@/pages/system/log/loginlog/index.vue'),
            meta: { title: { zh_CN: '登录日志', en_US: 'Login Logs' } },
          },
        ],
      },
    ],
  },
] satisfies RouteRecordRaw[];
