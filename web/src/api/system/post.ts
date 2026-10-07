import { request } from '@/utils/request';

export type PostDTO = {
  id: string;
  postCode: string;
  postName: string;
  orderNum: number;
  status: number;
};

export type PostQuery = {
  page?: number;
  pageSize?: number;
  keyword?: string;
};

export type PostSaveParams = {
  id?: string;
  postCode: string;
  postName: string;
  orderNum: number;
  status: number;
};

const Api = {
  List: '/system/post/list',
  Base: '/system/post',
};

export function fetchPostList(params?: PostQuery) {
  return request.get<{ list: PostDTO[]; total: number }>({ url: Api.List, params });
}

export function createPost(data: PostSaveParams) {
  return request.post<PostDTO>({ url: Api.Base, data });
}

export function updatePost(data: PostSaveParams) {
  return request.put<PostDTO>({ url: Api.Base, data });
}

export function removePost(id: string) {
  return request.delete<number>({ url: Api.Base, data: { id }, params: { id } });
}
