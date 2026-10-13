export const LAYOUT = () => import('@/layouts/index.vue');
export const BLANK_LAYOUT = () => import('@/layouts/blank.vue');
export const IFRAME = () => import('@/layouts/components/FrameBlank.vue');
/** 菜单 component 缺失时回退：404 页（非 500） */
export const NOT_FOUND_COMPONENT = () => import('@/pages/result/404/index.vue');
/** @deprecated 保留命名：旧代码引用；语义同 NOT_FOUND_COMPONENT */
export const EXCEPTION_COMPONENT = NOT_FOUND_COMPONENT;
export const PARENT_LAYOUT = () =>
  new Promise((resolve) => {
    resolve({ name: 'ParentLayout' });
  });

export const PAGE_NOT_FOUND_ROUTE = {
  path: '/:path(.*)*',
  name: '404Page',
  redirect: '/result/404',
};