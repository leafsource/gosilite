import { request } from '@/utils/request';

export interface LoginParams {
  username: string;
  password: string;
}

export interface SessionTokenResponse {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
  refreshToken: string;
}

export interface UserProfile {
  id: number;
  account: string;
  username: string;
  avatarURL: string;
}

export interface UserAuthInfo {
  user: UserProfile;
  roles: string[];
  permissions: string[];
}

export interface ChangePasswordParams {
  currentPassword: string;
  newPassword: string;
}

const Api = {
  Login: '/user/login',
  Info: '/user/info',
  Refresh: '/user/refresh',
  Logout: '/user/logout',
  Password: '/user/password',
};

export function login(data: LoginParams) {
  return request.post<SessionTokenResponse>({ url: Api.Login, data }, { withToken: false, skipAuthRefresh: true });
}

export function fetchUserInfo() {
  return request.get<UserAuthInfo>({ url: Api.Info });
}

export function refreshSession(refreshToken: string) {
  return request.post<SessionTokenResponse>(
    { url: Api.Refresh, data: { refreshToken } },
    { withToken: false, skipAuthRefresh: true },
  );
}

export function logout(refreshToken?: string) {
  return request.post({ url: Api.Logout, data: { refreshToken } }, { skipAuthRefresh: true });
}

export function changePassword(data: ChangePasswordParams) {
  return request.post({ url: Api.Password, data }, { skipAuthRefresh: true });
}
